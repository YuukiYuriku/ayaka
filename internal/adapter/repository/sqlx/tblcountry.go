package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblcountry"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCountryRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblCountryRepository) FetchCountries(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblcountry WHERE CntName LIKE ?"
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

	var countries []*tblcountry.Readtblcountry
	query := "SELECT CntCode, CntName, CreateDt FROM tblcountry WHERE CntName LIKE ? LIMIT ? OFFSET ?"

	if err := t.DB.SelectContext(ctx, &countries, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblcountry.Readtblcountry, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Country: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tblcountry.Readtblcountry, len(countries))
	for i, country := range countries {
		j++
		result[i] = &tblcountry.Readtblcountry{
			Number:      uint(j),
			CountryCode: country.CountryCode,
			CountryName: country.CountryName,
			CreateDate:  share.FormatDate(country.CreateDate),
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

func (t *TblCountryRepository) Create(ctx context.Context, data *tblcountry.Createtblcountry) (*tblcountry.Createtblcountry, error) {
	query := "INSERT INTO tblCountry (CntCode, CntName, CreateBy, CreateDt) VALUES (?, ?, ?, ?)"

	log.Printf("Executing query: %s with CntCode: %s, CntName: %s, CreateBy: %s, CreateDt: %s", query, data.CountryCode, data.CountryName, data.CreateBy, data.CreateDate)
	_, err := t.DB.ExecContext(ctx, query, data.CountryCode, data.CountryName, data.CreateBy, data.CreateDate)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Country: %w", err)
	}

	return data, nil
}

func (t *TblCountryRepository) Update(ctx context.Context, data *tblcountry.Updatetblcountry) (*tblcountry.Updatetblcountry, error) {
	query := "SELECT CntName FROM tblcountry WHERE CntCode = ?"
	var check tblcountry.Readtblcountry

	if err := t.DB.GetContext(ctx, &check, query, data.CountryCode); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, err
	}

	if check.CountryName == data.CountryName {
		return data, customerrors.ErrNoDataEdited
	}

	query = "UPDATE tblcountry SET CntName = ?, LastUpBy = ?, LastUpDt = ? WHERE CntCode = ?"

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	log.Printf("Executing query: %s with CntCode: %s and CntName: %s by %s", query, data.CountryCode, data.CountryName, data.UserCode)
	_, err = tx.ExecContext(ctx, query, data.CountryName, data.UserCode, data.LastUpdateDate, data.CountryCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating country: %w", err)
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	log.Printf("Executing query: %s with UserCode: %s, Code: %s, Category: %s, Date: %s", query, data.UserCode, data.CountryCode, "Country", data.LastUpdateDate)
	_, err = tx.ExecContext(ctx, query, data.UserCode, data.CountryCode, "Country", data.LastUpdateDate)

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
