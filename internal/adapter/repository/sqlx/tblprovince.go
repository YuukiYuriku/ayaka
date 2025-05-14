package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblprovince"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblProvinceRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (r *TblProvinceRepository) FetchProvinces(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblprovince WHERE ProvName LIKE ?"
	if err := r.DB.GetContext(ctx, &totalRecords, countQuery, search); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	var totalPages int
	var offset int

	// Calculate pagination details
	if param != nil {
		totalPages, offset = pagination.CountPagination(param, totalRecords)

		log.Printf("Calculated values - Total Records: %d, Total Pages: %d, Offset: %d",
			totalRecords, totalPages, offset)
	} else {
		param = &pagination.PaginationParam{
			PageSize: totalRecords,
			Page:     1,
		}
		totalPages = 1
		offset = 0
	}

	var provinces []*tblprovince.ReadTblProvince
	query := `
		SELECT 
			p.ProvCode, 
        	p.ProvName,
			c.CntCode, 
        	c.CntName, 
        	p.CreateDt
		FROM tblprovince p
   		JOIN tblcountry c ON p.CntCode = c.CntCode
   		WHERE p.ProvName LIKE ? 
        LIMIT ? OFFSET ?`

	if err := r.DB.SelectContext(ctx, &provinces, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblprovince.ReadTblProvince, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Provinces: %w", err)
	}

	j := offset
	result := make([]*tblprovince.ReadTblProvince, len(provinces))
	for i, province := range provinces {
		j++
		result[i] = &tblprovince.ReadTblProvince{
			Number:     uint(j),
			ProvCode:   province.ProvCode,
			ProvName:   province.ProvName,
			CntCode:    province.CntCode,
			CntName:    province.CntName, // Menyertakan CountryCode
			CreateDate: share.FormatDate(province.CreateDate),
		}
	}
	return &pagination.PaginationResponse{
		Data:         result,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}, nil
}

func (r *TblProvinceRepository) Create(ctx context.Context, data *tblprovince.CreateTblProvince) (*tblprovince.CreateTblProvince, error) {
	query := "INSERT INTO tblprovince (ProvCode, ProvName, CntCode, CreateBy, CreateDt, LastUpBy, LastUpDt) VALUES (?, ?, ?, ?, ?, ?, ?)"

	log.Printf("Executing query: %s", query)
	_, err := r.DB.ExecContext(ctx, query, data.ProvCode, data.ProvName, data.CountryCode, data.CreateBy, data.CreateDate, data.CreateBy, data.CreateDate)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Province: %w", err)
	}

	return data, nil
}

func (r *TblProvinceRepository) DetailProvince(ctx context.Context, provCode string) (*tblprovince.DetailTblProvince, error) {
	query := "SELECT p.ProvCode, p.ProvName, c.CntCode, c.CntName, p.CreateBy, p.CreateDt FROM tblprovince p JOIN tblcountry c ON c.CntCode = p.CntCode  WHERE p.ProvCode = ?"

	var province tblprovince.DetailTblProvince

	log.Printf("Executing query: %s with code: %s", query, provCode)
	if err := r.DB.GetContext(ctx, &province, query, provCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("province not found: %w", err)
		}
		return nil, fmt.Errorf("error Get Detail Province: %w", err)
	}

	return &province, nil
}

func (r *TblProvinceRepository) Update(ctx context.Context, data *tblprovince.UpdateTblProvince, userCode string) error {
	query := "SELECT ProvName, CntCode AS CntName FROM tblprovince WHERE ProvCode = ?"
	var existing tblprovince.DetailTblProvince

	if err := r.DB.GetContext(ctx, &existing, query, data.ProvCode); err != nil {
		return fmt.Errorf("error fetching existing province: %w", err)
	}

	if existing.ProvName == data.ProvName && existing.CntName == data.CountryCode {
		return customerrors.ErrNoDataEdited
	}

	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	query = "UPDATE tblprovince SET ProvName = ?, CntCode = ?, LastUpBy = ?, LastUpDt = ? WHERE ProvCode = ?"
	_, err = tx.ExecContext(ctx, query, data.ProvName, data.CountryCode, userCode, data.LastUpdateDate, data.ProvCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
		}
		return fmt.Errorf("error updating Province: %w", err)
	}

	// Log the activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"
	_, err = tx.ExecContext(ctx, query, userCode, data.ProvCode, "Province", data.LastUpdateDate)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
		}
		return fmt.Errorf("error inserting log activity: %w", err)
	}

	return tx.Commit()
}

func (t *TblProvinceRepository) GetGroupProvinces(ctx context.Context) ([]*datagroup.DataGroup, error) {
	var data []*datagroup.DataGroup
	var provinces []*tblprovince.ReadTblProvince
	var result []*tblprovince.GroupProvinceByCountry

	query := `
		SELECT 
			GROUP_CONCAT(
				CONCAT_WS(' - ', p.ProvCode, p.ProvName, p.CntCode, p.CreateDt)
				SEPARATOR ' | '
			) as GroupedData,
			c.CntName as CountryName
		FROM tblprovince p
		LEFT JOIN tblcountry c
		ON p.CntCode = c.CntCode
		GROUP BY c.CntName`

	log.Printf("Executing query: %s", query)

	if err := t.DB.SelectContext(ctx, &result, query); err != nil {
		log.Printf("Detailed error: %+v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, fmt.Errorf("error FetchProvinces: %w", err)
	}

	countProvince := 0
	countCountry := 0

	// Iterasi untuk setiap negara
	for _, group := range result {
		provinces = []*tblprovince.ReadTblProvince{} // Reset provinces untuk setiap negara

		// Split data yang dipisahkan dengan ' | '
		provinceStrings := strings.Split(group.GroupedData, " | ")

		// Iterasi untuk setiap data provinsi dalam negara
		for i, provStr := range provinceStrings {
			// Split data yang dipisahkan dengan ' - '
			parts := strings.Split(provStr, " - ")
			if len(parts) >= 4 {
				province := &tblprovince.ReadTblProvince{
					Number:     uint(i + 1),
					ProvCode:   parts[0],
					ProvName:   parts[1],
					CntName:    parts[2],
					CreateDate: parts[3],
				}
				provinces = append(provinces, province)
			} else {
				province := &tblprovince.ReadTblProvince{
					Number:     uint(i + 1),
					ProvCode:   parts[0],
					ProvName:   parts[1],
					CntName:    parts[2],
					CreateDate: "null",
				}
				provinces = append(provinces, province)
			}
			countProvince++
		}

		// Buat DataGroup untuk negara ini
		dataGroup := &datagroup.DataGroup{
			InGroup: group.CountryName,
			Data:    provinces,
		}
		data = append(data, dataGroup)
		countCountry++
	}

	return data, nil
}
