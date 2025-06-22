package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblvendorrating"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorRatingRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblVendorRatingRepository) Fetch(ctx context.Context, name, active string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var args []interface{}

	search := "%" + name + "%"
	countQuery := "SELECT COUNT(*) FROM tblvendorrating WHERE IndicatorCode LIKE ?"
	args = append(args, search)

	if active != "" {
		countQuery += " AND ActiveInd = ?"
		args = append(args, active)
	}

	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, args...); err != nil {
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

	var items []*tblvendorrating.Read
	query := `SELECT IndicatorCode,
				Description,
				ActiveInd,
				CreateDt
				FROM tblvendorrating
				WHERE IndicatorCode LIKE ?
				`

	if active != "" {
		query += " AND ActiveInd = ?"
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &items, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblvendorrating.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Vendor Rating: %w", err)
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

func (t *TblVendorRatingRepository) Create(ctx context.Context, data *tblvendorrating.Create) (*tblvendorrating.Create, error) {
	query := `INSERT INTO tblvendorrating
				(
					IndicatorCode,
					Description,
					ActiveInd,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.IndicatorCode,
		data.Description,
		data.Active,
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Vendor Rating: %w", err)
	}

	return data, nil
}

func (t *TblVendorRatingRepository) Update(ctx context.Context, data *tblvendorrating.Update) (*tblvendorrating.Update, error) {
	query := `UPDATE tblvendorrating SET
			Description = ?,
			ActiveInd = ?
			WHERE IndicatorCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	result, err := tx.ExecContext(ctx, query,
		data.Description,
		data.Active,
		data.IndicatorCode,
	)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating Vendor Rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to check row affected: %+v", err)
		return nil, fmt.Errorf("error check row affected: %w", err)
	}

	if rowsAffected == 0 {
		fmt.Println("rows affected: ", rowsAffected)
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, customerrors.ErrNoDataEdited
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.IndicatorCode, "VendorRating", data.LastUpdateDate)

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
