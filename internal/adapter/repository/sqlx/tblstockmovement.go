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
	"gitlab.com/ayaka/internal/domain/tblstockmovement"

	share "gitlab.com/ayaka/internal/domain/shared"
	// "gitlab.com/ayaka/internal/domain/shared/nulldatatype"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
	// "gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockMovementRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblStockMovementRepository) Fetch(
	ctx context.Context,
	warehouse []string,
	dateRangeStart, dateRangeEnd string,
	docType, itemCategory, itemName, batch string,
	param *pagination.PaginationParam,
) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var search []interface{}
	var endquery []string

	countQuery := `SELECT COUNT(*) 
				FROM tblstockmovement s
				JOIN tblitem i ON s.ItCode = i.ItCode WHERE s.CancelInd = 'N'`

	// Add date range logic
	if dateRangeStart != "" || dateRangeEnd != "" {
		if dateRangeStart != "" && dateRangeEnd != "" {
			endquery = append(endquery, `s.DocDt BETWEEN ? AND ?`)
			search = append(search, dateRangeStart, dateRangeEnd)
		} else if dateRangeStart != "" && dateRangeEnd == "" {
			endquery = append(endquery, `s.DocDt >= ?`)
			search = append(search, dateRangeStart)
		} else if dateRangeStart == "" && dateRangeEnd != "" {
			endquery = append(endquery, `s.DocDt <= ?`)
			search = append(search, dateRangeEnd)
		}
	}
	if docType != "" {
		endquery = append(endquery, `s.DocType LIKE ?`)
		docType = "%" + docType + "%"
		search = append(search, docType)
	}
	if itemCategory != "" {
		endquery = append(endquery, `i.ItCtCode = ?`)
		search = append(search, itemCategory)
	}
	if itemName != "" {
		endquery = append(endquery, `i.ItName LIKE ?`)
		itemName = "%" + itemName + "%"
		search = append(search, itemName)
	}
	if batch != "" {
		batch = "%" + batch + "%"
		endquery = append(endquery, `s.BatchNo LIKE ?`)
		search = append(search, batch)
	}
	if len(warehouse) > 0 {
		endquery = append(endquery, "s.WhsCode IN (?"+strings.Repeat(",?", len(warehouse)-1)+")")
		for _, detailWhs := range warehouse {
			search = append(search, detailWhs)
		}
	}

	if len(endquery) > 0 {
		countQuery += " AND " + strings.Join(endquery, " AND ")
	}

	fmt.Println("query count: ", countQuery)
	fmt.Println("search: ", search)

	// Count total records
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search...); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	// Pagination logic
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

	var data []*tblstockmovement.Fetch
	query := `
		SELECT
			s.DocType,
			s.DocNo,
			s.Source,
			s.DocDt,
			w.WhsName,
			s.ItCode,
			i.ItName,
			i.Specification,
			u.UomName,
			s.BatchNo,
			(s.Qty + s.Qty2 - s.Qty3) AS Qty,
			s.Remark,
			s.CreateBy,
			s.CreateDt
		FROM tblstockmovement s
		JOIN tblwarehouse w ON s.WhsCode = w.WhsCode
		JOIN tblitem i ON s.ItCode = i.ItCode
		JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode WHERE s.CancelInd = 'N'`

	if len(endquery) > 0 {
		query += " AND " + strings.Join(endquery, " AND ")
	}

	query += " LIMIT ? OFFSET ?"
	search = append(search, param.PageSize, offset)

	fmt.Println("query: ", query)
	fmt.Println("search: ", search)

	if err := t.DB.SelectContext(ctx, &data, query, search...); err != nil {
		log.Printf("Error executing query: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblstockmovement.Fetch, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch stock movement: %w", err)
	}

	// Number the rows for response
	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
		if detail.DocType == "Stock Mutation" {
			detail.DocType = fmt.Sprintf("%s (%s)", detail.DocType, detail.FromTo.String)
		}
		detail.DocDt = share.FormatDate(detail.DocDt)
	}

	// Prepare response
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

// func (t *TblStockMovementRepository) Fetch(
// 	ctx context.Context,
// 	warehouse []string,
// 	dateRangeStart, dateRangeEnd string,
// 	docType, itemCategory, itemName, batch string,
// 	param *pagination.PaginationParam,
// ) (*pagination.PaginationResponse, error) {
// 	var totalRecords int
// 	var search []interface{}
// 	var endquery []string

// 	countQuery := `SELECT COUNT(*) 
// 				FROM tblstockmovement s
// 				JOIN tblitem i ON s.ItCode = i.ItCode`

// 	// Add date range logic
// 	if dateRangeStart != "" || dateRangeEnd != "" {
// 		if dateRangeStart != "" && dateRangeEnd != "" {
// 			endquery = append(endquery, `s.DocDt BETWEEN ? AND ?`)
// 			search = append(search, dateRangeStart, dateRangeEnd)
// 		} else if dateRangeStart != "" && dateRangeEnd == "" {
// 			endquery = append(endquery, `s.DocDt >= ?`)
// 			search = append(search, dateRangeStart)
// 		} else if dateRangeStart == "" && dateRangeEnd != "" {
// 			endquery = append(endquery, `s.DocDt <= ?`)
// 			search = append(search, dateRangeEnd)
// 		}
// 	}
// 	if docType != "" {
// 		endquery = append(endquery, `s.DocType = ?`)
// 		search = append(search, docType)
// 	}
// 	if itemCategory != "" {
// 		endquery = append(endquery, `i.ItCtCode = ?`)
// 		search = append(search, itemCategory)
// 	}
// 	if itemName != "" {
// 		endquery = append(endquery, `i.ItName LIKE ?`)
// 		itemName = "%" + itemName + "%"
// 		search = append(search, itemName)
// 	}
// 	if batch != "" {
// 		batch = "%" + batch + "%"
// 		endquery = append(endquery, `s.BatchNo LIKE ?`)
// 		search = append(search, batch)
// 	}
// 	if len(warehouse) > 0 {
// 		endquery = append(endquery, "s.WhsCode IN (?"+strings.Repeat(",?", len(warehouse)-1)+")")
// 		for _, detailWhs := range warehouse {
// 			search = append(search, detailWhs)
// 		}
// 	}

// 	if len(endquery) > 0 {
// 		countQuery += " WHERE " + strings.Join(endquery, " AND ")
// 	}

// 	// Count total records
// 	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search...); err != nil {
// 		return nil, fmt.Errorf("error counting records: %w", err)
// 	}

// 	// Pagination logic
// 	var totalPages int
// 	var offset int

// 	if param != nil {
// 		totalPages, offset = pagination.CountPagination(param, totalRecords)

// 		log.Printf("Calculated values - Total Records: %d, Total Pages: %d, Offset: %d",
// 			totalRecords, totalPages, offset)
// 	} else {
// 		param = &pagination.PaginationParam{
// 			PageSize: totalRecords,
// 			Page:     1,
// 		}
// 		totalPages = 1
// 		offset = 0
// 	}

// 	var data []*tblstockmovement.Fetch
// 	query := `
// 		SELECT
// 			s.DocType,
// 			d.FromTo,
// 			s.DocNo,
// 			s.DNo,
// 			s.CancelInd,
// 			s.Source,
// 			s.Source2,
// 			s.DocDt,
// 			s.WhsCode,
// 			w.WhsName,
// 			s.Lot,
// 			s.Bin,
// 			s.ItCode,
// 			i.ItName,
// 			i.Specification,
// 			u.UomName,
// 			s.PropCode,
// 			s.BatchNo,
// 			s.Qty,
// 			s.Qty2,
// 			s.Qty3,
// 			s.MovingAvgCurCode,
// 			s.MovingAvgPrice,
// 			s.Remark,
// 			s.CreateBy,
// 			s.CreateDt,
// 			s.LastUpBy,
// 			s.LastUpDt
// 		FROM tblstockmovement s
// 		LEFT JOIN tblstockmutationdtl d ON d.DocNo = s.DocNo
// 		JOIN tblwarehouse w ON s.WhsCode = w.WhsCode
// 		JOIN tblitem i ON s.ItCode = i.ItCode
// 		JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode`

// 	if len(endquery) > 0 {
// 		query += " WHERE " + strings.Join(endquery, " AND ")
// 	}
// 	query += " LIMIT ? OFFSET ?"
// 	search = append(search, param.PageSize, offset)

// 	if err := t.DB.SelectContext(ctx, &data, query, search...); err != nil {
// 		log.Printf("Error executing query: %v", err)
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return &pagination.PaginationResponse{
// 				Data:         make([]*tblstockmovement.Fetch, 0),
// 				TotalRecords: 0,
// 				TotalPages:   0,
// 				CurrentPage:  param.Page,
// 				PageSize:     param.PageSize,
// 				HasNext:      false,
// 				HasPrevious:  false,
// 			}, nil
// 		}
// 		return nil, fmt.Errorf("error Fetch stock movement: %w", err)
// 	}

// 	// Number the rows for response
// 	j := offset
// 	for _, detail := range data {
// 		j++
// 		detail.Number = uint(j)
// 		if detail.DocType == "Stock Mutation" {
// 			detail.DocType = fmt.Sprintf("%s (%s)", detail.DocType, detail.FromTo.String)
// 		}
// 		detail.DocDt = share.FormatDate(detail.DocDt)
// 	}

// 	// Prepare response
// 	response := &pagination.PaginationResponse{
// 		Data:         data,
// 		TotalRecords: totalRecords,
// 		TotalPages:   totalPages,
// 		CurrentPage:  param.Page,
// 		PageSize:     param.PageSize,
// 		HasNext:      param.Page < totalPages,
// 		HasPrevious:  param.Page > 1,
// 	}

// 	return response, nil
// }
