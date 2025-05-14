package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblcurrency"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCurrencyRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblCurrencyRepository) Fetch(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblcurrency WHERE CurName LIKE ? OR CurCode LIKE ?"
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search, search); err != nil {
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

	var data []*tblcurrency.Fetch
	query := "SELECT CurCode, CurName, CreateDt FROM tblcurrency WHERE CurName LIKE ? OR CurCode LIKE ? LIMIT ? OFFSET ?"

	if err := t.DB.SelectContext(ctx, &data, query, search, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblcurrency.Fetch, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Currency: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tblcurrency.Fetch, len(data))
	for i, detaildata := range data {
		j++
		result[i] = &tblcurrency.Fetch{
			Number:       uint(j),
			CurrencyCode: detaildata.CurrencyCode,
			CurrencyName: detaildata.CurrencyName,
			CreateDate:   share.FormatDate(detaildata.CreateDate),
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

func (t *TblCurrencyRepository) Create(ctx context.Context, data *tblcurrency.Create) (*tblcurrency.Create, error) {
	query := "INSERT INTO tblcurrency (CurCode, CurName, CreateBy, CreateDt) VALUES (?, ?, ?, ?)"

	_, err := t.DB.ExecContext(ctx, query, data.CurrencyCode, data.CurrencyName, data.CreateBy, data.CreateDate)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Currency: %w", err)
	}

	return data, nil
}

func (t *TblCurrencyRepository) Update(ctx context.Context, data *tblcurrency.Update) (*tblcurrency.Update, error) {
	query := "SELECT CurName FROM tblcurrency WHERE CUrCode = ?"
	var check tblcurrency.Fetch

	if err := t.DB.GetContext(ctx, &check, query, data.CurrencyCode); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, err
	}

	if check.CurrencyName == data.CurrencyName {
		return data, customerrors.ErrNoDataEdited
	}

	query = "UPDATE tblcurrency SET CurName = ?, LastUpBy = ?, LastUpDt = ? WHERE CurCode = ?"

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, data.CurrencyName, data.UserCode, data.LastUpdateDate, data.CurrencyCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating currency: %w", err)
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.UserCode, data.CurrencyCode, "Currency", data.LastUpdateDate)

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
