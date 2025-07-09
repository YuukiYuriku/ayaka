package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/tblhistoryofstock"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblHistoryOfStockRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblHistoryOfStockRepository) Fetch(ctx context.Context, item, batch, source string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var args []interface{}
	
	searchItem := "%" + item + "%"
	searchBatch := "%" + batch + "%"
	searchSource := "%" + source + "%"
	
	countQuery := `SELECT COUNT(*) FROM 
		tblhistoryofstock h
		JOIN tblitem i ON h.ItCode = i.ItCode
		WHERE (i.ItName LIKE ?
		AND h.BatchNo LIKE ? AND h.source LIKE ?) AND h.CancelInd = 'N';
	`
	args = append(args, searchItem, searchBatch, searchSource)

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

	var items []*tblhistoryofstock.Read
	query := `SELECT
			h.ItCode,
			i.ItName,
			h.BatchNo,
			h.Source
		FROM tblhistoryofstock h
		JOIN tblitem i ON h.ItCode = i.ItCode
		WHERE (i.ItName LIKE ?
		AND h.BatchNo LIKE ? AND h.Source LIKE ?) AND h.CancelInd = 'N'
	`

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &items, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblhistoryofstock.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch History of Stock: %w", err)
	}

	// proses data
	j := offset
	for _, item := range items {
		j++
		item.Number = uint(j)
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