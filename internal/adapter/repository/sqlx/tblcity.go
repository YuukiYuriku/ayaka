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
	"gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	"gitlab.com/ayaka/internal/domain/tblcity"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCityRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblCityRepository) FetchCities(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblcity WHERE CityName LIKE ?"
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	var totalPages int
	var offset int

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

	var cities []*tblcity.ReadTblCity
	query := `
        SELECT 
            c.CityCode, 
            c.CityName,
			p.ProvCode, 
            p.ProvName, 
            c.RingCode, 
			c.LocationCode,
            c.CreateDt 
        FROM tblcity c 
        JOIN tblprovince p ON c.ProvCode = p.ProvCode 
        WHERE CityName LIKE ? 
        LIMIT ? OFFSET ?`

	if err := t.DB.SelectContext(ctx, &cities, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblcity.ReadTblCity, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch City: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tblcity.ReadTblCity, len(cities))
	for i, city := range cities {
		j++
		result[i] = &tblcity.ReadTblCity{
			Number:     uint(j),
			CityCode:   city.CityCode,
			CityName:   city.CityName,
			ProvCode:   city.ProvCode,
			Province:   city.Province,
			RingArea:   city.RingArea,
			CreateDate: share.FormatDate(city.CreateDate),
		}
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         result,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}

	return response, nil
}

func (t *TblCityRepository) Create(ctx context.Context, data *tblcity.CreateTblCity) (*tblcity.CreateTblCity, error) {
	query := "INSERT INTO tblcity (CityCode, CityName, ProvCode, RingCode, LocationCode, CreateBy, CreateDt) VALUES (?, ?, ?, ?, ?, ?, ?)"

	log.Printf("Executing query: %s", query)
	_, err := t.DB.ExecContext(ctx, query, data.CityCode, data.CityName, data.Province, data.RingArea, data.Location, data.CreateBy, data.CreateDate)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create City: %w", err)
	}

	return data, nil
}

func (t *TblCityRepository) Update(ctx context.Context, data *tblcity.UpdateTblCity) (*tblcity.UpdateTblCity, error) {
	query := "SELECT CityName, ProvCode, RingCode, LocationCode FROM tblcity WHERE CityCode = ?"
	var check tblcity.ReadTblCity

	if err := t.DB.GetContext(ctx, &check, query, data.CityCode); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, err
	}

	if check.CityName == data.CityName && check.Province == data.Province && check.RingArea == data.RingArea && check.Location == data.Location {
		return data, customerrors.ErrNoDataEdited
	}

	query = "UPDATE tblcity SET CityName = ?, ProvCode = ?, RingCode = ?, LocationCode = ?, LastUpBy = ?, LastUpDt = ? WHERE CityCode = ?"

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	log.Printf("Executing query: %s", query)
	_, err = tx.ExecContext(ctx, query, data.CityName, data.Province, data.RingArea, data.Location, data.UserCode, data.LastUpdateDate, data.CityCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating City: %w", err)
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	log.Printf("Executing query: %s with UserCode: %s, Code: %s, Category: %s, Date: %s", query, data.UserCode, data.CityCode, "City", data.LastUpdateDate)
	_, err = tx.ExecContext(ctx, query, data.UserCode, data.CityCode, "City", data.LastUpdateDate)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error insert to log activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblCityRepository) GetGroupCities(ctx context.Context) ([]*datagroup.DataGroup, error) {
	var data []*datagroup.DataGroup
	var cities []*tblcity.ReadTblCity
	var result []*tblcity.GroupCityByProv

	query := `SELECT 
			GROUP_CONCAT(
				CONCAT_WS(' - ', c.CityCode, c.CityName, p.ProvName, c.CreateDt, c.RingCode)
				SEPARATOR ' | '
			) as GroupedData,
			p.ProvName as ProvName 
			FROM tblprovince p
			RIGHT JOIN tblcity c
			ON p.ProvCode = c.ProvCode
			GROUP BY p.ProvName`

	log.Printf("Executing query: %s", query)

	if err := t.DB.SelectContext(ctx, &result, query); err != nil {
		log.Printf("Detailed error: %+v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, fmt.Errorf("error Fetch City: %w", err)
	}

	countCity := 0
	countProv := 0
	// Iterasi untuk setiap provinsi
	for _, group := range result {
		cities = []*tblcity.ReadTblCity{} // Reset cities untuk setiap provinsi

		// Split data yang dipisahkan dengan ' | '
		cityStrings := strings.Split(group.GroupedData, " | ")

		// Iterasi untuk setiap data kota dalam provinsi
		for i, cityStr := range cityStrings {
			// Split data yang dipisahkan dengan ' - '
			parts := strings.Split(cityStr, " - ")
			if len(parts) >= 5 {
				city := &tblcity.ReadTblCity{
					Number:     uint(i + 1),
					CityCode:   parts[0],
					CityName:   parts[1],
					Province:   parts[2],
					CreateDate: parts[3],
					RingArea:   nulldatatype.NewNullStringDataType(parts[4]),
				}
				cities = append(cities, city)
			} else {
				city := &tblcity.ReadTblCity{
					Number:     uint(i + 1),
					CityCode:   parts[0],
					CityName:   parts[1],
					Province:   parts[2],
					CreateDate: parts[3],
					RingArea:   nulldatatype.NewNullStringDataType("null"),
				}
				cities = append(cities, city)
			}
			countCity++
		}

		// Buat DataGroup untuk provinsi ini
		dataGroup := &datagroup.DataGroup{
			InGroup: group.ProvName,
			Data:    cities,
		}
		data = append(data, dataGroup)
		countProv++
	}
	return data, nil
}
