package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/tblcoa"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCoaRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblCoaRepository) FetchCoa(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblcoa WHERE AcDesc LIKE ?"
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

	var coas []*tblcoa.ReadTblCoa
	query := "SELECT AcNo, AcDesc FROM tblcoa WHERE AcDesc LIKE ? LIMIT ? OFFSET ?"

	if err := t.DB.SelectContext(ctx, &coas, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblcoa.ReadTblCoa, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch coa: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tblcoa.ReadTblCoa, len(coas))
	for i, coa := range coas {
		j++
		result[i] = &tblcoa.ReadTblCoa{
			Number:        uint(j),
			AccountNumber: coa.AccountNumber,
			Description:   coa.Description,
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
