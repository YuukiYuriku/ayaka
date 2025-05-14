package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tbluom"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblUomRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblUomRepository) FetchUom(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tbluom WHERE UomName LIKE ?"
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

	var uoms []*tbluom.ReadTblUom
	query := "SELECT UomCode, UomName, CreateDt FROM tbluom WHERE UomName LIKE ? LIMIT ? OFFSET ?"

	if err := t.DB.SelectContext(ctx, &uoms, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tbluom.ReadTblUom, 0),
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
	result := make([]*tbluom.ReadTblUom, len(uoms))
	for i, uom := range uoms {
		j++
		result[i] = &tbluom.ReadTblUom{
			Number:     uint(j),
			UomCode:    uom.UomCode,
			UomName:    uom.UomName,
			CreateDate: share.FormatDate(uom.CreateDate),
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

func (t *TblUomRepository) Create(ctx context.Context, data *tbluom.CreateTblUom) (*tbluom.CreateTblUom, error) {
	query := "INSERT INTO tbluom (UomCode, UomName, CreateBy, CreateDt) VALUES (?, ?, ?, ?)"

	_, err := t.DB.ExecContext(ctx, query, data.UomCode, data.UomName, data.CreateBy, data.CreateDate)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Country: %w", err)
	}

	return data, nil
}

func (t *TblUomRepository) Update(ctx context.Context, data *tbluom.UpdateTblUom) (*tbluom.UpdateTblUom, error) {
	query := "SELECT UomName FROM tbluom WHERE UomCode = ?"
	var check tbluom.ReadTblUom

	if err := t.DB.GetContext(ctx, &check, query, data.UomCode); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, err
	}

	if check.UomName == data.UomName {
		return data, customerrors.ErrNoDataEdited
	}

	query = "UPDATE tbluom SET UomName = ?, LastUpBy = ?, LastUpDt = ? WHERE UomCode = ?"

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, query, data.UomName, data.UserCode, data.LastUpdateDate, data.UomCode)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error updating uom: %w", err)
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.UserCode, data.UomCode, "Uom", data.LastUpdateDate)

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
