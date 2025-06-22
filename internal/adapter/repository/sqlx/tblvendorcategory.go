package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblvendorcategory"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorCategoryRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblVendorCategoryRepository) Fetch(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblvendorcategory WHERE VendorCatCode LIKE ? OR VendorCatName LIKE ?"
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

	var items []*tblvendorcategory.Read
	query := `SELECT VendorCatCode,
				VendorCatName,
				CreateDt
				FROM tblvendorcategory
				WHERE VendorCatCode LIKE ? OR VendorCatName LIKE ?
				LIMIT ? OFFSET ?`

	if err := t.DB.SelectContext(ctx, &items, query, search, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblvendorcategory.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Vendor Category: %w", err)
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

func (t *TblVendorCategoryRepository) Create(ctx context.Context, data *tblvendorcategory.Create) (*tblvendorcategory.Create, error) {
	query := `INSERT INTO tblvendorcategory
				(
					VendorCatCode,
					VendorCatName,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.VendorCategoryCode,
		data.VendorCategoryName,
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Vendor Category: %w", err)
	}

	return data, nil
}

func (t *TblVendorCategoryRepository) Update(ctx context.Context, data *tblvendorcategory.Update) (*tblvendorcategory.Update, error) {
	query := `UPDATE tblvendorcategory SET
			VendorCatName = ?
			WHERE VendorCatCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	result, err := tx.ExecContext(ctx, query,
		data.VendorCategoryName,
		data.VendorCategoryCode,
	)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating Vendor Category: %w", err)
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

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.VendorCategoryCode, "VendorCategory", data.LastUpdateDate)

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