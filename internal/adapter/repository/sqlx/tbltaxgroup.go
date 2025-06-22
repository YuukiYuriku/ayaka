package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tbltaxgroup"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblTaxGroupRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblTaxGroupRepository) Fetch(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tbltaxgroup WHERE TaxGroupCode LIKE ? OR TaxGroupName LIKE ?"
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

	var taxGroup []*tbltaxgroup.ReadTblTaxGroup
	query := `SELECT TaxGroupCode,
				TaxGroupName,
				CreateDt
				FROM tbltaxgroup
				WHERE TaxGroupCode LIKE ? OR TaxGroupName LIKE ?
				LIMIT ? OFFSET ?`

	if err := t.DB.SelectContext(ctx, &taxGroup, query, search, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tbltaxgroup.ReadTblTaxGroup, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Tax Group: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tbltaxgroup.ReadTblTaxGroup, len(taxGroup))
	for i, item := range taxGroup {
		j++
		result[i] = &tbltaxgroup.ReadTblTaxGroup{
			Number:       uint(j),
			TaxGroupCode: item.TaxGroupCode,
			TaxGroupName: item.TaxGroupName,
			CreateDate:   share.FormatDate(item.CreateDate),
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

func (t *TblTaxGroupRepository) Create(ctx context.Context, data *tbltaxgroup.Create) (*tbltaxgroup.Create, error) {
	query := `INSERT INTO tbltaxgroup
				(
					TaxGroupCode,
					TaxGroupName,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.TaxGroupCode,
		data.TaxGroupName,
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Tax Group: %w", err)
	}

	return data, nil
}

func (t *TblTaxGroupRepository) Update(ctx context.Context, data *tbltaxgroup.Update) (*tbltaxgroup.Update, error) {
	query := `UPDATE tbltaxgroup SET
			TaxGroupName = ?,
			LastUpDt = ?,
			LastUpBy = ?
			WHERE TaxGroupCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	result, err := tx.ExecContext(ctx, query,
		data.TaxGroupName,
		data.LastUpdateDate,
		data.LastUpdateBy,
		data.TaxGroupCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return data, customerrors.ErrNoDataEdited
		}
		return nil, fmt.Errorf("error updating tax group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to check row affected: %+v", err)
		return nil, fmt.Errorf("error check row affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, customerrors.ErrNoDataEdited
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.TaxGroupCode, "TaxGroup", data.LastUpdateDate)

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