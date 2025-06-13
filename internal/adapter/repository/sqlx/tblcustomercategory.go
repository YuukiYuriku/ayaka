package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblcustomercategory"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCustomerCategoryRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblCustomerCategoryRepository) Fetch(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblcustomercategory WHERE CustCatCode LIKE ? OR CustCatName LIKE ?"
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

	var custCategories []*tblcustomercategory.Read
	query := `SELECT CustCatCode,
				CustCatName,
				CreateDt
				FROM tblcustomercategory
				WHERE CustCatCode LIKE ? OR CustCatName LIKE ?
				LIMIT ? OFFSET ?`

	if err := t.DB.SelectContext(ctx, &custCategories, query, search, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblcustomercategory.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Customer Category: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tblcustomercategory.Read, len(custCategories))
	for i, item := range custCategories {
		j++
		result[i] = &tblcustomercategory.Read{
			Number:               uint(j),
			CustomerCategoryCode: item.CustomerCategoryCode,
			CustomerCategoryName: item.CustomerCategoryName,
			CreateDate:           share.FormatDate(item.CreateDate),
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

func (t *TblCustomerCategoryRepository) Create(ctx context.Context, data *tblcustomercategory.Create) (*tblcustomercategory.Create, error) {
	query := `INSERT INTO tblcustomercategory
				(
					CustCatCode,
					CustCatName,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.CustomerCategoryCode,
		data.CustomerCategoryName,
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Customer Category: %w", err)
	}

	return data, nil
}

func (t *TblCustomerCategoryRepository) Update(ctx context.Context, data *tblcustomercategory.Update) (*tblcustomercategory.Update, error) {
	query := `UPDATE tblcustomercategory SET
			CustCatName = ?
			WHERE CustCatCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	result, err := tx.ExecContext(ctx, query,
		data.CustomerCategoryName,
		data.CustomerCategoryCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating item's category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to check row affected: %+v", err)
		return nil, fmt.Errorf("error check row affected: %w", err)
	}

	if rowsAffected == 0  {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, customerrors.ErrNoDataEdited
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.CustomerCategoryCode, "CustomerCategory", data.LastUpdateDate)

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