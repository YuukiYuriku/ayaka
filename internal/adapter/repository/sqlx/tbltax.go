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
	"gitlab.com/ayaka/internal/domain/tbltax"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblTaxRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblTaxRepository) Fetch(ctx context.Context, name, category string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchName := "%" + name + "%"

	if strings.TrimSpace(category) != "" {
		countQuery := "SELECT COUNT(*) FROM tbltax WHERE (TaxCode LIKE ? OR TaxName LIKE ?) AND TaxGroupCode = ?"
		if err := t.DB.GetContext(ctx, &totalRecords, countQuery, searchName, searchName, category); err != nil {
			return nil, fmt.Errorf("error counting records: %w", err)
		}
	} else {
		countQuery := "SELECT COUNT(*) FROM tbltax WHERE (TaxCode LIKE ? OR TaxName LIKE ?)"
		if err := t.DB.GetContext(ctx, &totalRecords, countQuery, searchName, searchName); err != nil {
			return nil, fmt.Errorf("error counting records: %w", err)
		}
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

	var items []*tbltax.Read
	var query string

	if strings.TrimSpace(category) != "" {
		query = `SELECT t.TaxCode,
				t.TaxName,
				t.TaxGroupCode,
				t.TaxRate,
				tg.TaxGroupName,
				t.CreateDt
				FROM tbltax t
				JOIN tbltaxgroup tg ON t.TaxGroupCode = tg.TaxGroupCode
				WHERE (t.TaxCode LIKE ? OR t.TaxName LIKE ?)
				AND i.TaxGroupCode = ?
				LIMIT ? OFFSET ?`
	} else {
		query = `SELECT t.TaxCode,
				t.TaxName,
				t.TaxGroupCode,
				t.TaxRate,
				tg.TaxGroupName,
				t.CreateDt
				FROM tbltax t
				JOIN tbltaxgroup tg ON t.TaxGroupCode = tg.TaxGroupCode
				WHERE (t.TaxCode LIKE ? OR t.TaxName LIKE ?)
				LIMIT ? OFFSET ?`
	}

	if strings.TrimSpace(category) != "" {
		if err := t.DB.SelectContext(ctx, &items, query, searchName, searchName, category, param.PageSize, offset); err != nil {
			if errors.Is(err, sql.ErrNoRows) || items == nil {
				return &pagination.PaginationResponse{
					Data:         make([]*tbltax.Read, 0),
					TotalRecords: 0,
					TotalPages:   0,
					CurrentPage:  param.Page,
					PageSize:     param.PageSize,
					HasNext:      false,
					HasPrevious:  false,
				}, nil
			}
			return nil, fmt.Errorf("error Fetch tax: %w", err)
		}
	} else {
		if err := t.DB.SelectContext(ctx, &items, query, searchName, searchName, param.PageSize, offset); err != nil {
			if errors.Is(err, sql.ErrNoRows) || items == nil {
				fmt.Println("error fetch: ", err.Error())
				return &pagination.PaginationResponse{
					Data:         make([]*tbltax.Read, 0),
					TotalRecords: 0,
					TotalPages:   0,
					CurrentPage:  param.Page,
					PageSize:     param.PageSize,
					HasNext:      false,
					HasPrevious:  false,
				}, nil
			}
			return nil, fmt.Errorf("error Fetch tax: %w", err)
		}
	}

	// proses data
	j := offset
	for _, item := range items {
		j++
		item.Number = uint(j)
		item.CreateDate = share.FormatDate(item.CreateDate)
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         items,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}

	return response, nil
}

func (t *TblTaxRepository) Create(ctx context.Context, data *tbltax.Create) (*tbltax.Create, error) {
	query := `INSERT INTO tbltax
				(
					TaxCode,
					TaxName,
					TaxRate,
					TaxGroupCode,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.TaxCode,
		data.TaxName,
		data.TaxRate,
		data.TaxGroupCode,
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create tax: %w", err)
	}

	return data, nil
}

func (t *TblTaxRepository) Update(ctx context.Context, data *tbltax.Update) (*tbltax.Update, error) {
	query := `SELECT TaxName,
			TaxRate,
			TaxGroupCode
			FROM tbltax
			WHERE TaxCode = ? LIMIT 1`

	var temp tbltax.Read

	err := t.DB.GetContext(ctx, &temp, query, data.TaxCode)
	if err != nil {
		return nil, fmt.Errorf("error get prev tax: %w", err)
	}

	fmt.Printf("\ntax name: %s - %s", data.TaxName, temp.TaxName)
	fmt.Printf("\ntax rate: %f - %f", data.TaxRate, temp.TaxRate)
	fmt.Printf("\ntax grup: %s - %s", data.TaxGroupCode, temp.TaxGroupCode)
	if data.TaxName == temp.TaxName &&
		fmt.Sprintf("%.6f", data.TaxRate) == fmt.Sprintf("%.6f", temp.TaxRate) &&
		data.TaxGroupCode == temp.TaxGroupCode {
		return nil, customerrors.ErrNoDataEdited
	}

	query = `UPDATE tbltax SET
			TaxName = ?,
			TaxRate = ?,
			TaxGroupCode = ?,
			LastUpBy = ?,
			LastUpDt = ?
			WHERE TaxCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, query,
		data.TaxName,
		data.TaxRate,
		data.TaxGroupCode,
		data.LastUpdateBy,
		data.LastUpdateDate,
		data.TaxCode,
	)

	if err != nil {
		fmt.Println("masuk: ", err.Error())
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return data, customerrors.ErrNoDataEdited
		}
		return nil, fmt.Errorf("error updating tax: %w", err)
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.TaxCode, "Tax", data.LastUpdateDate)

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