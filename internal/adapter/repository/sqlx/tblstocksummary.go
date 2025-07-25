package sqlx

import (
	"context"
	"fmt"
	"log"
	"strings"

	"database/sql"
	"errors"

	// "strings"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/tblstocksummary"

	// share "gitlab.com/ayaka/internal/domain/shared"
	// "gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
	// "gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockSummaryRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblStockSummaryRepository) Fetch(ctx context.Context, warehouse []string, date, itemCatCode, itemCode, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var search []interface{}
	var endquery []string

	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM tblstocksummary s
			JOIN tblitem i ON s.ItCode = i.ItCode`

	if date != "" {
		endquery = append(endquery, `s.CreateDt <= ?`)
		search = append(search, date)
	}
	if itemCode != "" {
		endquery = append(endquery, `i.ItCode = ?`)
		search = append(search, itemCode)
	}
	if itemName != "" {
		endquery = append(endquery, `i.ItName LIKE ?`)
		itemName = "%" + itemName + "%"
		search = append(search, itemName)
	}
	if itemCatCode != "" {
		endquery = append(endquery, `i.ItCtCode = ?`)
		search = append(search, itemCatCode)
	}

	if len(warehouse) > 0 {
		endquery = append(endquery, "s.WhsCode IN (?"+strings.Repeat(",?", len(warehouse)-1)+")")
		for _, detailWhs := range warehouse {
			search = append(search, detailWhs)
		}
	}

	if len(endquery) > 0 {
		countQuery += " WHERE " + strings.Join(endquery, " AND ")
	}

	countQuery += `
			GROUP BY s.WhsCode, s.ItCode
			HAVING SUM(Qty + Qty2 - Qty3) != 0
		) AS grouped`

	fmt.Println("Query count: ", countQuery)
	fmt.Println("args: ", search)
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search...); err != nil {
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

	var data []*tblstocksummary.Fetch
	query := `
		SELECT
			w.WhsName AS WhsName,
			s.ItCode AS ItCode,
			i.ItCodeInternal AS ItCodeInternal,
			i.ItName AS ItName,
			c.ItCtName as ItCtName,
			i.ActInd AS ActInd,
			SUM(Qty + Qty2 - Qty3) AS Stock,
			u.UomName AS UomName
		FROM tblstocksummary s
		JOIN tblwarehouse w ON s.WhsCode = w.WhsCode
		JOIN tblitem i ON s.ItCode = i.ItCode
		JOIN tblitemcategory c ON c.ItCtCode = i.ItCtCode
		JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode`

	if len(endquery) > 0 {
		query += " WHERE " + strings.Join(endquery, " AND ")
	}

	query += ` GROUP BY s.WhsCode, s.ItCode, w.WhsName
		HAVING SUM(Qty + Qty2 - Qty3) != 0
		LIMIT ? OFFSET ?`
	search = append(search, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, search...); err != nil {
		log.Printf("Error executing query: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblstocksummary.Fetch, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch stock summary: %w", err)
	}

	j := offset
	var filtered []*tblstocksummary.Fetch
	for _, detail := range data {
		if detail.Quantity != 0 {
			j++
			detail.Number = uint(j)
			filtered = append(filtered, detail)
		}
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         filtered,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}

	return response, nil
}

func (t *TblStockSummaryRepository) GetItem(ctx context.Context, itemName, itemCatCode, batch, warehouse string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var search []interface{}
	var endquery []string

	countQuery := `SELECT COUNT(*) FROM (
		SELECT 1
		FROM tblstocksummary s
		JOIN tblitem i ON s.ItCode = i.ItCode
		WHERE s.WhsCode = ?`
	search = append(search, warehouse)

	if itemName != "" {
		endquery = append(endquery, ` i.ItName LIKE ? `)
		itemName = "%" + itemName + "%"
		search = append(search, itemName)
	}
	if itemCatCode != "" {
		endquery = append(endquery, ` i.ItCtCode = ? `)
		search = append(search, itemCatCode)
	}
	if batch != "" {
		endquery = append(endquery, ` s.BatchNo LIKE ? `)
		batch = "%" + batch + "%"
		search = append(search, batch)
	}

	if len(endquery) > 0 {
		countQuery += " AND " + strings.Join(endquery, " AND ")
	}
	countQuery += " AND i.ActInd = 'Y'"

	countQuery += `
		GROUP BY i.ItName, s.BatchNo
	) AS grouped`

	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search...); err != nil {
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

	var data []*tblstocksummary.GetItem
	query := `
		SELECT
			s.ItCode,
			i.ItName,
			s.BatchNo,
			SUM(Qty + Qty2 - Qty3) AS Stock,
			u.UomName
		FROM tblstocksummary s
		JOIN tblitem i ON s.ItCode = i.ItCode
		JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
		WHERE s.WhsCode = ? `

	if len(endquery) > 0 {
		query += " AND " + strings.Join(endquery, " AND ")
	}
	query += " AND i.ActInd = 'Y'"

	query += `
			GROUP BY s.ItCode, s.BatchNo
			LIMIT ? OFFSET ?`
	search = append(search, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, search...); err != nil {
		log.Printf("Error executing query: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblstocksummary.GetItem, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch stock summary: %w", err)
	}

	j := offset
	var filtered []*tblstocksummary.GetItem
	for _, detail := range data {
		if detail.Stock != 0 {
			j++
			detail.Number = uint(j)
			filtered = append(filtered, detail)
		}
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         filtered,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}

	return response, nil
}
