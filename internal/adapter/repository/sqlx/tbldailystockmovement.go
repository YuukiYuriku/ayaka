package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	// "strings"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/tbldailystockmovement"

	// "gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
	// "gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDailyStockMovementRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblDailyStockMovementRepository) Fetch(ctx context.Context, warehouse, date, itemName, itemCategoryName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var search []interface{}
	var endquery []string

	// Query Count
	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM tblstockmovement s
			JOIN tblitem i ON s.ItCode = i.ItCode
			JOIN tblitemcategory c ON i.ItCtCode = c.ItCtCode
	`

	// Filtering
	if date != "" {
		endquery = append(endquery, `s.DocDt <= ?`)
		search = append(search, date)
	}
	if itemName != "" {
		endquery = append(endquery, `i.ItName LIKE ?`)
		search = append(search, "%"+itemName+"%")
	}
	if itemCategoryName != "" {
		endquery = append(endquery, `c.ItCtName LIKE ?`)
		search = append(search, "%"+itemCategoryName+"%")
	}
	if warehouse != "" {
		endquery = append(endquery, `s.WhsCode = ?`)
		search = append(search, warehouse)
	}

	if len(endquery) > 0 {
		countQuery += " WHERE " + strings.Join(endquery, " AND ") + " AND s.CancelInd = 'N'"
	}

	countQuery += `
			GROUP BY s.ItCode, s.WhsCode, i.ItName
		) AS grouped`

	// Hitung total record
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search...); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	// Hitung pagination
	var totalPages, offset int
	if param != nil {
		totalPages, offset = pagination.CountPagination(param, totalRecords)
	} else {
		param = &pagination.PaginationParam{PageSize: totalRecords, Page: 1}
		totalPages = 1
		offset = 0
	}

	// Query Data
	var data []*tbldailystockmovement.Read
	query := `
		SELECT
			s.ItCode,
			i.ItName,
			i.Specification,
			c.ItCtName,
			u.UomName,

			SUM(CASE WHEN s.DocType = 'Initial Stock' AND s.DocDt = ? AND s.CancelInd = 'N' THEN s.Qty ELSE 0 END) AS Qty,
			SUM(CASE WHEN s.DocDt = ? AND s.CancelInd = 'N' THEN s.Qty2 ELSE 0 END) AS Qty2,
			SUM(CASE WHEN s.DocDt = ? AND s.CancelInd = 'N' THEN 0 - s.Qty3 ELSE 0 END) AS Qty3,

			(
				SELECT 
					SUM(COALESCE(s2.Qty, 0) + COALESCE(s2.Qty2, 0) - COALESCE(s2.Qty3, 0))
				FROM tblstockmovement s2
				WHERE s2.ItCode = s.ItCode
					AND s2.DocDt <= ?
					AND s2.CancelInd = 'N'
			) AS ReakStock

		FROM tblstockmovement s
		JOIN tblitem i ON s.ItCode = i.ItCode
		JOIN tblitemcategory c ON i.ItCtCode = c.ItCtCode
		JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
	`

	// Tambahkan filter WHERE
	if len(endquery) > 0 {
		query += " WHERE " + strings.Join(endquery, " AND ") + " AND s.CancelInd = 'N'"
	}

	query += ` GROUP BY s.ItCode, i.ItName
		LIMIT ? OFFSET ?`

	// Urutan argumen: 3x date untuk Qty, Qty2, Qty3 dan 3x date untuk RealStock, lalu filter
	searchArgs := []interface{}{date, date, date, time.Now().Format("20060102")}
	searchArgs = append(searchArgs, search...)
	searchArgs = append(searchArgs, param.PageSize, offset)

	fmt.Println("query: ", query)
	fmt.Println("args: ", searchArgs)
	// Eksekusi query
	if err := t.DB.SelectContext(ctx, &data, query, searchArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         []*tbldailystockmovement.Read{},
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error fetching stock data: %w", err)
	}

	// Tambahkan nomor urut
	j := offset
	for _, d := range data {
		j++
		d.Number = uint(j)
		d.Total = d.Init + d.In + d.Out
	}

	// Response
	return &pagination.PaginationResponse{
		Data:         data,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}, nil
}
