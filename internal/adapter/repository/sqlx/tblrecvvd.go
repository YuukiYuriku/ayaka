package sqlx

import (
	"context"
	"database/sql"
	"errors"

	// "database/sql"
	// "errors"
	"fmt"
	"log"
	"strings"

	// "github.com/jmoiron/sqlx"
	"gitlab.com/ayaka/internal/adapter/repository"
	// share "gitlab.com/ayaka/internal/domain/shared"

	// "gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblrecvvddtl"
	"gitlab.com/ayaka/internal/domain/tblrecvvdhdr"

	// "gitlab.com/ayaka/internal/domain/tblmasteritem"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDirectPurchaseReceiveRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblDirectPurchaseReceiveRepository) Fetch(ctx context.Context, doc string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"

	countQuery := "SELECT COUNT(*) FROM tblstockinitialhdr WHERE DocNo LIKE ?"

	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, searchDoc); err != nil {
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

	var data []*tblrecvvdhdr.Fetch
	query := `
	SELECT
		hdr.DocNo,
		hdr.DocDt,
		dtl.CancelInd,
		hdr.LocalDocNo,
		w.WhsName,
		hdr.VdCode,
		hdr.VdDONo,
		i.ItName,
		i.ItCodeInternal,
		i.ForeignName,
		dtl.BatchNo,
		dtl.Lot,
		dtl.Bin,
		hdr.CurCode,
		c.CurName,
		dtl.UPrice,
		dtl.QtyPurchase,
		i.PurchaseUOMCode,
		u.UomName,
		dtl.Discount,
		dtl.RoundingValue,
		( ( (dtl.UPrice * dtl.QtyPurchase) - (dtl.UPrice * dtl.QtyPurchase * (dtl.Discount / 100)) ) + dtl.RoundingValue) AS Amount
	FROM tblrecvvdhdr hdr
	JOIN tblrecvvddtl dtl ON hdr.DocNo = dtl.DocNo
	JOIN tblitem i ON i.ItCode = dtl.ItCode
	JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
	JOIN tblcurrency c ON hdr.CurCode = c.CurCode
	JOIN tblwarehouse w ON hdr.WhsCode = w.WhsCode
	WHERE hdr.DocNo LIKE ? LIMIT ? OFFSET ?
	`

	if err := t.DB.SelectContext(ctx, &data, query, searchDoc, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblrecvvdhdr.Fetch, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Direct Purchase Receive: %w", err)
	}

	// proses data
	j := offset
	for _, detaildata := range data {
		j++
		detaildata.Number = uint(j)
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

func (t *TblDirectPurchaseReceiveRepository) Detail(ctx context.Context, docNo string) (*tblrecvvdhdr.Detail, error) {
	var header tblrecvvdhdr.Detail
	query := `
	SELECT
		hdr.DocNo,
		hdr.DocDt,
		hdr.LocalDocNo,
		hdr.WhsCode,
		w.WhsName,
		hdr.VdCode,
		hdr.VdDONo,
		hdr.SiteCode,
		hdr.PtCode,
		hdr.CurCode,
		c.CurName,
		hdr.TaxCode1,
		(hdr.TaxCode1 + hdr.TaxCode2 + hdr.TaxCode3) AS TotalTax,
		hdr.DiscountAmt,
		hdr.Remark,
		hdr.ProjectCode
	FROM tblrecvvdhdr hdr
	JOIN tblwarehouse w ON hdr.WhsCode = w.WhsCode
	JOIN tblcurrency c ON hdr.CurCode = c.CurCode
	WHERE hdr.DocNo = ?`

	if err := t.DB.GetContext(ctx, &header, query, docNo); err != nil {
		return nil, fmt.Errorf("error detail header direct purchase receive: %w", err)
	}

	var details []tblrecvvddtl.Detail
	query = `
	SELECT
		dtl.DNo,
		dtl.CancelInd,
		dtl.CancelReason,
		i.ItName,
		i.ItCodeInternal,
		dtl.BatchNo,
		dtl.UPrice,
		dtl.QtyPurchase,
		u.UomName,
		dtl.Discount,
		dtl.RoundingValue
	FROM tblrecvvddtl dtl
	JOIN tblitem i ON dtl.ItCode = i.ItCode
	JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
	WHERE dtl.DocNo = ?`

	if err := t.DB.SelectContext(ctx, &details, query, docNo); err != nil {
		return nil, fmt.Errorf("error detail direct purchase receive: %w", err)
	}

	header.TotalQuantity = 0.0
	var total float32 = 0.0
	for _, detail := range details {
		header.TotalQuantity += float32(detail.Quantity)
		total += (detail.Price * detail.Quantity) - (detail.Price * detail.Quantity * (detail.Discount / 100)) + detail.Rounding
	}

	header.Detail = details
	header.GrandTotal = total

	return &header, nil
}

func (t *TblDirectPurchaseReceiveRepository) Create(ctx context.Context, data *tblrecvvdhdr.Create) (*tblrecvvdhdr.Create, error) {
	countDetail := len(data.Detail)
	var args []interface{}

	// Mulai transaksi
	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// Pastikan rollback selalu dijalankan jika terjadi error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %+v", rbErr)
			}
		}
	}()

	// Insert header
	query := `
	INSERT INTO tblrecvvdhdr (
		DocNo,
		DocDt,
		LocalDocNo,
		WhsCode,
		VdCode,
		VdDONo,
		SiteCode,
		PtCode,
		CurCode,
		TaxCode1,
		Remark,
		ProjectCode,
		POInd,
		CreateBy,
		CreateDt
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	args = append(args,
		data.DocNo,
		data.Date,
		data.LocalCode,
		data.WarehouseCode,
		data.Vendor,
		data.DO,
		data.Site,
		data.TermOfPayment,
		data.CurrencyCode,
		data.Tax,
		data.Remark,
		data.Project,
		"N",
		data.CreateBy,
		data.CreateDate,
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert header: %w", err)
	}

	if countDetail > 0 {
		query = `
		INSERT INTO tblrecvvddtl (
			DocNo,
			DNo,
			ItCode,
			BatchNo,
			Source,
			Lot,
			Bin,
			QtyPurchase,
			UPrice,
			Discount,
			RoundingValue,
			Remark,
			ExpiredDt,
			CreateBy,
			CreateDt
		) VALUES `
		var placeholders []string
		args = args[:0]

		for _, detail := range data.Detail {
			placeholders = append(placeholders, `(?, ?, ?, ?, ?, "-", "-", ?, ?, ?, ?, ?, ?, ?, ?)`)
			args = append(args,
				data.DocNo,
				detail.DNo,
				detail.ItemCode,
				detail.Batch,
				detail.LocalCode,
				detail.Quantity,
				detail.Price,
				detail.Discount,
				detail.Rounding,
				detail.Remark,
				detail.Expired,
				data.CreateBy,
				data.CreateDate,
			)
		}

		query += strings.Join(placeholders, ",")
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert details: %w", err)
		}

		// Insert ke tabel stock movement
		query = `INSERT INTO tblstockmovement (
			DocType,
			DocNo,
			DNo,
			CancelInd,
			DocDt,
			WhsCode,
			Source,
			ItCode,
			BatchNo,
			Qty,
			Qty2,
			Qty3,
			Remark,
			CreateBy,
			CreateDt
		) VALUES `
		var movementValues []string
		args = args[:0]

		for _, detail := range data.Detail {
			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			args = append(args,
				"Direct Purchase Receive",
				data.DocNo,
				detail.DNo,
				detail.Cancel,
				data.Date,
				data.WarehouseCode,
				detail.LocalCode,
				detail.ItemCode,
				detail.Batch,
				detail.Quantity,
				detail.Quantity,
				detail.Quantity,
				data.Remark,
				data.CreateBy,
				data.CreateDate,
			)
		}

		query += strings.Join(movementValues, ",")
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert stock movement: %w", err)
		}

		// Insert ke tabel stock summary
		query = `INSERT INTO tblstocksummary (
			WhsCode,
			Lot,
			Bin,
			Source,
			ItCode,
			BatchNo,
			Qty,
			Qty2,
			Qty3,
			Remark,
			CreateBy,
			CreateDt
		) VALUES `
		var summaryValues []string
		args = args[:0]

		for _, detail := range data.Detail {
			summaryValues = append(summaryValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			args = append(args,
				data.WarehouseCode,
				"-",
				"-",
				detail.ItemCode,
				detail.ItemCode,
				detail.Batch,
				detail.Quantity,
				0,
				0,
				data.Remark,
				data.CreateBy,
				data.CreateDate,
			)
		}

		query += strings.Join(summaryValues, ",") + `
			ON DUPLICATE KEY UPDATE
				Qty = Qty + VALUES(Qty),
				Qty2 = Qty2,
				Qty3 = Qty3,
				LastUpBy = VALUES(CreateBy),
				LastUpDt = VALUES(CreateDt);`
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert stock summary: %w", err)
		}
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblDirectPurchaseReceiveRepository) Update(ctx context.Context, data, oldData *tblrecvvdhdr.Detail, lastUpBy, lastUpDt string) (*tblrecvvdhdr.Detail, error) {
	count := len(data.Detail)

	if count > 0 {
		var args, argsReason, argsLog, argsMov []interface{}
		var whenClauses []string
		var placeholders []string
		var whenClausesMov []string

		status := false

		qLog := `INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES `

		for i := 0; i < count; i++ {
			if oldData.Detail[i].Cancel != booldatatype.FromBool(true) && oldData.Detail[i].Cancel != data.Detail[i].Cancel {
				whenClauses = append(whenClauses, `WHEN DNo = ? THEN ?`)
				args = append(args, data.Detail[i].DNo, data.Detail[i].Cancel)
				argsReason = append(argsReason, data.Detail[i].DNo, data.Detail[i].CancelReason)

				whenClausesMov = append(whenClausesMov, `WHEN DocNo = ? AND DNo = ? THEN ?`)
				argsMov = append(argsMov, data.DocNo, data.Detail[i].DNo, data.Detail[i].Cancel)

				placeholders = append(placeholders, (`(?, ?, ?, ?)`))
				argsLog = append(argsLog, lastUpBy, data.Detail[i].DNo, "DirectPurchaseReceiveDtl", lastUpDt)
				status = true
			}
		}

		if len(whenClauses) == 0 {
			return nil, customerrors.ErrNoDataEdited
		}

		if status {
			query := `UPDATE tblrecvvddtl
					SET 
						CancelInd = CASE ` + strings.Join(whenClauses, " ") + `
						ELSE CancelInd
						END,
						CancelReason = CASE ` + strings.Join(whenClauses, " ") + `
						ELSE CancelReason
						END,
						LastUpDt = ?,
						LastUpBy = ?
					WHERE DocNo = ?;`
			args = append(args, argsReason...)
			args = append(args, lastUpDt, lastUpBy, data.DocNo)

			queryMov := `UPDATE tblstockmovement
					SET 
						CancelInd = CASE ` + strings.Join(whenClausesMov, " ") + `
						ELSE CancelInd
						END,
						LastUpDt = ?,
						LastUpBy = ?
					WHERE DocNo = ?;`
			argsMov = append(argsMov, lastUpDt, lastUpBy, data.DocNo)

			qLog += `(?, ?, ?, ?)`
			qLog += ", " + strings.Join(placeholders, ", ")
			argsLog = append(argsLog, lastUpBy, data.DocNo, "StockInitial", lastUpDt)

			tx, err := t.DB.BeginTxx(ctx, nil)
			if err != nil {
				log.Printf("Failed to start transaction: %+v", err)
				return nil, fmt.Errorf("error starting transaction: %w", err)
			}

			// Pastikan rollback dipanggil jika transaksi tidak berhasil
			defer func() {
				if err != nil {
					// Rollback jika error terjadi
					if rbErr := tx.Rollback(); rbErr != nil {
						log.Printf("Failed to rollback transaction: %+v", rbErr)
					}
				}
			}()

			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return nil, fmt.Errorf("error executing update query: %w", err)
			}

			_, err = tx.ExecContext(ctx, queryMov, argsMov...)
			if err != nil {
				return nil, fmt.Errorf("error executing update query for tblstockmovement: %w", err)
			}

			// Eksekusi log activity
			_, err = tx.ExecContext(ctx, qLog, argsLog...)
			if err != nil {
				log.Printf("Failed to execute update query log: %v", err)
				return nil, fmt.Errorf("error insert to log activity: %w", err)
			}

			// Commit transaksi jika semua query berhasil
			if err := tx.Commit(); err != nil {
				log.Printf("Failed to commit transaction: %+v", err)
				return nil, fmt.Errorf("error committing transaction: %w", err)
			}

			return data, nil
		}
	}
	return nil, customerrors.ErrNoDataEdited
}
