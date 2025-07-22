package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	"gitlab.com/ayaka/internal/domain/tblorderreport"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblOrderReportRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblOrderReportRepository) ByVendor(ctx context.Context, date string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var args []interface{}

	countQuery := `SELECT COUNT(*) FROM 
		(
			WITH VendorOrders AS (
				SELECT 
					o.VendorCode,
					v.VendorName,
					COUNT(DISTINCT o.DocNo) AS OrderFreq,

					-- Gunakan rata-rata dari PO jika tersedia, fallback ke harga order
					ROUND(
					COALESCE(
						AVG(pord.Total / NULLIF(pord.Qty, 0)),
						AVG(o.Price)
					),
					2
					) AS AveragePrice,

					-- Total nilai order (prioritaskan data PO, fallback ke harga order)
					SUM(
						o.Qty * COALESCE(pord.Total / NULLIF(pord.Qty, 0), o.Price)
					) AS TotalOrderAmount

					FROM tblorderreport o
					LEFT JOIN tblpurchasematerialreceivedtl pmrd
						ON pmrd.DocNo = o.DocNo
						AND pmrd.ItCode = o.ItCode
					LEFT JOIN tblpurchaseorderdtl pord
						ON pmrd.PurchaseOrderDocNo = pord.DocNo
						AND pmrd.PurchaseOrderDNo = pord.DNo
					JOIN tblvendorhdr v ON v.VendorCode = o.VendorCode
					WHERE o.CancelInd = 'N'
						AND LEFT(o.DocDt, 6) = ?
					GROUP BY o.VendorCode, v.VendorName
				),

				Total AS (
				SELECT 
					SUM(TotalOrderAmount) AS TotalAllOrderAmount
				FROM VendorOrders
				)

				SELECT 
				vo.VendorName,
				vo.OrderFreq,
				vo.AveragePrice,
				vo.TotalOrderAmount,
				ROUND(
					vo.TotalOrderAmount / NULLIF(t.TotalAllOrderAmount, 0) * 100,
					2
				) AS TotalOrderPercent
				FROM VendorOrders vo
				CROSS JOIN Total t
				ORDER BY vo.TotalOrderAmount DESC
		) AS grouped
	`
	args = append(args, date)

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

	var items []*tblorderreport.OrderReportByVendor
	query := `WITH VendorOrders AS (
					SELECT 
						o.VendorCode,
						v.VendorName,
						COUNT(DISTINCT o.DocNo) AS OrderFreq,

						-- Gunakan rata-rata dari PO jika tersedia, fallback ke harga order
						ROUND(
							COALESCE(
								AVG(pord.Total / NULLIF(pord.Qty, 0)),
								AVG(o.Price)
							),
						2
						) AS AveragePrice,

						-- Total nilai order (prioritaskan data PO, fallback ke harga order)
						SUM(
							o.Qty * COALESCE(pord.Total / NULLIF(pord.Qty, 0), o.Price)
						) AS TotalOrderAmount

					FROM tblorderreport o
					LEFT JOIN tblpurchasematerialreceivedtl pmrd
						ON pmrd.DocNo = o.DocNo
						AND pmrd.ItCode = o.ItCode
					LEFT JOIN tblpurchaseorderdtl pord
						ON pmrd.PurchaseOrderDocNo = pord.DocNo
						AND pmrd.PurchaseOrderDNo = pord.DNo
					JOIN tblvendorhdr v ON v.VendorCode = o.VendorCode
					WHERE o.CancelInd = 'N'
						AND LEFT(o.DocDt, 6) = ?
					GROUP BY o.VendorCode, v.VendorName
				),

				Total AS (
					SELECT 
						SUM(TotalOrderAmount) AS TotalAllOrderAmount
					FROM VendorOrders
				)

				SELECT 
				vo.VendorName,
				vo.OrderFreq,
				vo.AveragePrice,
				vo.TotalOrderAmount,
				ROUND(
					vo.TotalOrderAmount / NULLIF(t.TotalAllOrderAmount, 0) * 100,
					2
				) AS TotalOrderPercent
				FROM VendorOrders vo
				CROSS JOIN Total t
				ORDER BY vo.TotalOrderAmount DESC
	`

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &items, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblorderreport.OrderReportByVendor, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch order report by vendor: %w", err)
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
