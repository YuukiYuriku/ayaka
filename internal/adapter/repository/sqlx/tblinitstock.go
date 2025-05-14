package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblinitialstock"
	"gitlab.com/ayaka/internal/domain/tblinitialstockdtl"
	"gitlab.com/ayaka/internal/domain/tblmasteritem"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblInitStockRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblInitStockRepository) Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhs := "%" + warehouse + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblstockinitialhdr WHERE DocNo LIKE ? AND WhsCode LIKE ?"
	args = append(args, searchDoc, searchWhs)

	if startDate != "" && endDate != "" {
		countQuery += " AND DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

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

	var data []*tblinitialstock.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			w.WhsName,
			w.WhsCode,
			c.CurCode,
			c.CurName,
			i.ExcRate,
			i.Remark
			FROM tblstockinitialhdr i
			JOIN tblwarehouse w ON i.WhsCode = w.WhsCode
			JOIN tblcurrency c ON i.CurCode = c.CurCode
			WHERE i.DocNo LIKE ? AND i.WhsCode LIKE ?`
	args = append(args, searchDoc, searchWhs)

	if startDate != "" && endDate != "" {
		query += " AND i.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblmasteritem.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch init stock: %w", err)
	}

	// proses data
	j := offset
	docsNo := make([]string, len(data))
	for i, detail := range data {
		j++
		detail.Number = uint(j)
		detail.Date = share.FormatDate(detail.Date)
		docsNo[i] = detail.DocNo
	}

	if len(docsNo) == 0 {
		return &pagination.PaginationResponse{
			Data:         make([]*tblinitialstock.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblinitialstockdtl.Read{}
	detailQuery := `SELECT 
				d.DocNo, d.DNo, d.CancelInd, d.ItCode, d.BatchNo, d.Qty, d.UPrice,
				i.ItName, i.ItCodeInternal,
				u.UomName
			FROM tblstockinitialdtl d
			JOIN tblitem i ON d.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUomCode = u.UomCode
			WHERE d.DocNo IN (?);`

	query, args, err := sqlx.In(detailQuery, docsNo)
	if err != nil {
		return nil, fmt.Errorf("error preparing detail query: %w", err)
	}
	query = t.DB.Rebind(query)

	if err := t.DB.SelectContext(ctx, &details, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching details: %w", err)
	}

	// Kelompokkan detail berdasarkan DocNo
	detailMap := make(map[string][]tblinitialstockdtl.Read)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range data {
		h.Details = detailMap[h.DocNo]
		var count float32 = 0.0
		for _, data := range detailMap[h.DocNo] {
			count += float32(data.Quantity)
		}
		h.TotalQuantity = count
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

func (t *TblInitStockRepository) Detail(ctx context.Context, docNo string) (*tblinitialstock.Detail, error) {
	query := `SELECT
		i.WhsCode AS WhsCode,
		i.CurCode AS CurCode,
		c.CurName AS CurName,
		i.ExcRate AS ExcRate,
		i.Remark AS Remark
	FROM tblstockinitialhdr i
	JOIN tblcurrency c ON i.CurCode = c.CurCode
	WHERE i.DocNo = ?;`

	var header tblinitialstock.Detail

	// get header init stock
	if err := t.DB.GetContext(ctx, &header, query, docNo); err != nil {
		return nil, fmt.Errorf("error detail header init stock: %w", err)
	}

	query = `SELECT
		d.DNo AS DNo,
		d.CancelInd AS CancelInd,
		d.ItCode AS ItCode,
		i.ItName AS ItName,
		i.ItCodeInternal AS ItCodeInternal,
		d.BatchNo AS BatchNo,
		d.Qty AS Qty,
		u.UomName AS UomName,
		d.UPrice AS UPrice
	FROM tblstockinitialdtl d
	JOIN tblitem i ON d.ItCode = i.ItCode
	JOIN tbluom u ON i.PurchaseUomCode = u.UomCode
	WHERE d.DocNo = ?;`

	var details []tblinitialstockdtl.Read

	// get detail init stock
	if err := t.DB.SelectContext(ctx, &details, query, docNo); err != nil {
		return nil, fmt.Errorf("error detail init stock: %w", err)
	}

	var count float32 = 0.0
	for _, data := range details {
		count += float32(data.Quantity)
	}

	header.Detail = details
	header.TotalQuantity = float32(count)

	return &header, nil
}

func (t *TblInitStockRepository) Create(ctx context.Context, data *tblinitialstock.Create) (*tblinitialstock.Create, error) {
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

	// Insert ke tabel header
	query := `INSERT INTO tblstockinitialhdr (
		DocNo,
		DocDt,
		WhsCode,
		CurCode,
		ExcRate,
		Remark,
		CreateDt,
		CreateBy
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	args = []interface{}{
		data.DocNo,
		data.Date,
		data.WarehouseCode,
		data.CurrencyCode,
		data.Rate,
		data.Remark,
		data.CreateDate,
		data.CreateBy,
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert header: %w", err)
	}

	if countDetail > 0 {
		// Insert ke tabel detail
		query = `INSERT INTO tblstockinitialdtl (
			DocNo,
			DNo,
			CancelInd,
			ItCode,
			BatchNo,
			Lot,
			Qty,
			Qty2,
			Qty3,
			UPrice,
			CreateDt,
			CreateBy
		) VALUES `
		var placeholders []string
		args = args[:0]

		for _, detail := range data.Detail {
			placeholders = append(placeholders, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			args = append(args,
				data.DocNo,
				detail.DNo,
				detail.Cancel,
				detail.ItemCode,
				detail.Batch,
				"-",
				detail.Quantity,
				detail.Quantity,
				detail.Quantity,
				detail.Price,
				data.CreateDate,
				data.CreateBy,
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
				data.DocType,
				data.DocNo,
				detail.DNo,
				detail.Cancel,
				data.Date,
				data.WarehouseCode,
				detail.Source,
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

func (t *TblInitStockRepository) Update(ctx context.Context, data, oldData *tblinitialstock.Detail, lastUpBy, lastUpDt string) (*tblinitialstock.Detail, error) {
	count := len(data.Detail)

	if count > 0 {
		var args, argsLog, argsMov []interface{}
		var whenClauses []string
		var placeholders []string
		var whenClausesMov []string

		status := false

		qLog := `INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES `

		for i := 0; i < count; i++ {
			if oldData.Detail[i].Cancel != booldatatype.FromBool(true) && oldData.Detail[i].Cancel != data.Detail[i].Cancel {
				whenClauses = append(whenClauses, `WHEN DNo = ? THEN ?`)
				args = append(args, data.Detail[i].DNo, data.Detail[i].Cancel)

				whenClausesMov = append(whenClausesMov, `WHEN DocNo = ? AND DNo = ? THEN ?`)
				argsMov = append(argsMov, data.DocNo, data.Detail[i].DNo, data.Detail[i].Cancel)

				placeholders = append(placeholders, (`(?, ?, ?, ?)`))
				argsLog = append(argsLog, lastUpBy, data.Detail[i].DNo, "StockInitialDtl", lastUpDt)
				status = true
			}
		}

		if len(whenClauses) == 0 {
			return nil, customerrors.ErrNoDataEdited
		}

		query := `UPDATE tblstockinitialdtl
				SET 
					CancelInd = CASE ` + strings.Join(whenClauses, " ") + `
					ELSE CancelInd
					END,
					LastUpDt = ?,
					LastUpBy = ?
				WHERE DocNo = ?;`
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

		if status {
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
