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
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tbltransferbetweenwhs"

	// share "gitlab.com/ayaka/internal/domain/shared"
	// "gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
	// "gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblTransferItemBetweenWhsRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblTransferItemBetweenWhsRepository) GetMaterial(ctx context.Context, itemName, batch, warehouseFrom, warehouseTo string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var search []interface{}
	var endquery []string

	fmt.Println("Warehouse From: ", warehouseFrom)
	fmt.Println("Warehouse To: ", warehouseTo)

	countQuery := `SELECT COUNT(*) FROM (
		SELECT 1
		FROM tblmaterialtransferdtl s
		JOIN tblmaterialtransferhdr h ON s.DocNo = h.DocNo
		JOIN tblitem i ON s.ItCode = i.ItCode
		WHERE h.WhsCodeFrom = ? AND h.WhsCodeTo = ?`
	search = append(search, warehouseFrom, warehouseTo)

	if itemName != "" {
		endquery = append(endquery, ` i.ItName LIKE ? `)
		itemName = "%" + itemName + "%"
		search = append(search, itemName)
	}
	if batch != "" {
		endquery = append(endquery, ` s.BatchNo LIKE ? `)
		batch = "%" + batch + "%"
		search = append(search, batch)
	}

	if len(endquery) > 0 {
		countQuery += " AND " + strings.Join(endquery, " AND ")
	}

	countQuery += `
		GROUP BY 
			s.DocNo,
			s.DNo,
			s.ItCode,
			i.ItName,
			s.BatchNo,
			s.Qty,
			h.WhsCodeFrom,
			h.WhsCodeTo
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

	var data []*tbltransferbetweenwhs.GetMaterial
	query := `SELECT 
			t.DocNo AS DocNoMaterialTransfer,
			h.DocDt,
			t.ItCode,
			i.ItName,
			t.BatchNo,
			(t.Qty - COALESCE(SUM(r.QtyActual), 0)) AS QtyRemaining,
			u.UomName
		FROM tblmaterialtransferdtl t
		JOIN tblmaterialtransferhdr h ON h.DocNo = t.DocNo
		JOIN tblitem i ON i.ItCode = t.ItCode
		JOIN tblUom u on i.PurchaseUomCode = u.UomCode
		LEFT JOIN tblmaterialreceivedtl r 
			ON r.DocNoMaterialTransfer = t.DocNo
			AND r.ItCode = t.ItCode
			AND (r.BatchNo = t.BatchNo OR (r.BatchNo IS NULL AND t.BatchNo IS NULL))
		WHERE h.WhsCodeFrom = ? AND h.WhsCodeTo = ? AND t.CancelInd = 'N'
 	`

	if len(endquery) > 0 {
		query += " AND " + strings.Join(endquery, " AND ")
	}

	query += `
			GROUP BY 
				t.DocNo,
				t.DNo,
				t.ItCode,
				i.ItName,
				t.BatchNo,
				t.Qty,
				h.WhsCodeFrom,
				h.WhsCodeTo
			ORDER BY 
				t.DocNo, t.DNo
			LIMIT ? OFFSET ?`
	search = append(search, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, search...); err != nil {
		log.Printf("Error executing query: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tbltransferbetweenwhs.GetMaterial, 0),
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
	var filtered []*tbltransferbetweenwhs.GetMaterial
	for _, detail := range data {
		if detail.QtyRemaining != 0 {
			j++
			detail.Number = uint(j)
			detail.Date = share.FormatDate(detail.Date)
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

func (t *TblTransferItemBetweenWhsRepository) Fetch(ctx context.Context, item, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchItem := "%" + item + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) FROM tbltransferbetweenwhs t
		JOIN tblitem i ON t.ItCode = i.ItCode
		WHERE i.ItName LIKE ? AND t.CancelInd = 'N'
	`
	var endQuery string
	args = append(args, searchItem)

	if warehouseFrom != "" {
		endQuery += " AND t.WhsFrom = ? "
		args = append(args, warehouseFrom)
	}

	if warehouseTo != "" {
		endQuery += " AND t.WhsTo = ? "
		args = append(args, warehouseTo)
	}

	if startDate != "" && endDate != "" {
		endQuery += " AND DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	countQuery += endQuery
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

	var data []*tbltransferbetweenwhs.Read

	query := `SELECT
			t.DocNo,
			t.DocDt,
			wf.WhsName AS WhsFrom,
			wt.WhsName AS WhsTo,
			i.ItName,
			t.BatchNo,
			t.Qty,
			u.UomName,
			t.Remark
		FROM tbltransferbetweenwhs t
		JOIN tblwarehouse wf
			ON t.WhsFrom = wf.WhsCode
		JOIN tblwarehouse wt
			ON t.WhsTo = wt.WhsCode
		JOIN tblitem i
			ON t.ItCode = i.ItCode
		JOIN tbluom u
			ON i.PurchaseUOMCode = u.UomCode
		WHERE i.ItName LIKE ?
	`
	endQuery += ` LIMIT ? OFFSET ? `
	args = append(args, param.PageSize, offset)

	query += endQuery
	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tbltransferbetweenwhs.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch material transfer: %w", err)
	}

	// proses data
	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
		detail.Date = share.FormatDate(detail.Date)
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         data,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}

	return response, nil
}
