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
	"gitlab.com/ayaka/internal/domain/tblstockadjustmentdtl"
	"gitlab.com/ayaka/internal/domain/tblstockadjustmenthdr"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	// "gitlab.com/ayaka/internal/pkg/customerrors"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockAdjustRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblStockAdjustRepository) Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhs := "%" + warehouse + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblstockadjustmenthdr WHERE DocNo LIKE ? AND WhsCode LIKE ?"
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

	var data []*tblstockadjustmenthdr.Detail
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.WhsCode,
			i.Remark,
			w.WhsName
			FROM tblstockadjustmenthdr i
			JOIN tblwarehouse w ON i.WhsCode = w.WhsCode
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
				Data:         make([]*tblstockadjustmenthdr.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch stock adjust: %w", err)
	}

	// proses data
	j := offset
	docsNo := make([]string, len(data))
	for i, detail := range data {
		j++
		detail.Number = uint(j)
		detail.TblDate = share.FormatDate(detail.Date)
		detail.Date = share.ToDatePicker(detail.Date)
		docsNo[i] = detail.DocNo
	}

	if len(docsNo) == 0 {
		return &pagination.PaginationResponse{
			Data:         make([]*tblstockadjustmenthdr.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblstockadjustmentdtl.Detail{}
	detailQuery := `SELECT
				d.DocNo, 
				d.DNo,
				i.ItName,
				d.BatchNo,
				d.Qty,
				d.QtyActual,
				(d.QtyActual - d.Qty) AS Balance,
				u.UomName,
				i.Specification
			FROM tblstockadjustmentdtl d
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
	detailMap := make(map[string][]tblstockadjustmentdtl.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range data {
		h.Details = detailMap[h.DocNo]
		var count float32 = 0.0
		for _, data := range detailMap[h.DocNo] {
			count += float32(data.Balance)
		}
		h.TotalBalance = count
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

func (t *TblStockAdjustRepository) Detail(ctx context.Context, docNo string) (*tblstockadjustmenthdr.Detail, error) {
	query := `SELECT
		a.DocNo,
		a.DocDt,
		a.WhsCode,
		w.WhsName,
		a.Remark
	FROM tblstockadjustmenthdr a
	JOIN tblwarehouse w ON a.WhsCode = w.WhsCode
	WHERE a.DocNo = ?;`

	var header tblstockadjustmenthdr.Detail

	// get header stock adjustment
	if err := t.DB.GetContext(ctx, &header, query, docNo); err != nil {
		return nil, fmt.Errorf("error detail header stock adjustment: %w", err)
	}

	query = `SELECT
		d.DNo,
		i.ItName,
		d.BatchNo,
		d.Qty,
		d.QtyActual,
		(d.QtyActual - d.Qty) AS Balance,
		u.UomName,
		i.Spesification
	FROM tblstockadjustmentdtl d
	JOIN tblitem i ON d.ItCode = i.ItCode
	JOIN tbluom u ON i.PurchaseUomCode = u.UomCode
	WHERE d.DocNo = ?;`

	var details []tblstockadjustmentdtl.Detail

	// get detail stock adjust
	if err := t.DB.SelectContext(ctx, &details, query, docNo); err != nil {
		return nil, fmt.Errorf("error detail stock adjust: %w", err)
	}

	header.Details = details

	return &header, nil
}

func (t *TblStockAdjustRepository) Create(ctx context.Context, data *tblstockadjustmenthdr.Create) (*tblstockadjustmenthdr.Create, error) {
	query := `INSERT INTO tblstockadjustmenthdr 
	(
		DocNo,
		DocDt,
		WhsCode,
		Remark,
		CreateDt,
		CreateBy
	) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		data.WarehouseCode,
		data.Remark,
		data.CreateDate,
		data.CreateBy,
	)

	// transaction begin
	var err error
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

	// insert header
	query += strings.Join(placeholders, "")
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Error insert header: %+v", err)
		return nil, fmt.Errorf("error Insert Header: %w", err)
	}

	if len(data.Details) > 0 {
		// detail query
		query = `INSERT INTO tblstockadjustmentdtl (
			DocNo,
			DNo,
			ItCode,
			BatchNo,
			Lot,
			Bin,
			Qty,
			Qty2,
			Qty3,
			QtyActual,
			QtyActual2,
			QtyActual3,
			CreateDt,
			CreateBy
		) VALUES `

		placeholders = placeholders[:0]
		args = args[:0]

		// stock summary query
		queryStockSummary := `INSERT INTO tblstocksummary (
			WhsCode,
			Lot,
			Bin,
			Source,
			ItCode,
			BatchNo,
			Qty,
			Qty2,
			Qty3,
			CreateBy,
			CreateDt
		) VALUES `
		var placeholdersStockSummary []string
		var argsStockSummary []interface{}

		// stock movement query
		queryStockMovement := `INSERT INTO tblstockmovement (
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
		var placeholdersStockMovement []string
		var argsStockMovement []interface{}

		// stock history of stock
		queryHistory := `INSERT INTO tblhistoryofstock (
			ItCode,
			BatchNo,
			Source,
			CancelInd,
			CreateBy,
			CreateDt
		) VALUES `
		var placeholdersHistory []string
		var argsHistory []interface{}

		for i, detail := range data.Details {
			balance := detail.StockActual - detail.StockSystem

			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				fmt.Sprintf("%03d", i+1),
				detail.ItemCode,
				detail.Batch,
				"-",
				"-",
				detail.StockSystem,
				0,
				0,
				detail.StockActual,
				0,
				0,
				data.CreateDate,
				data.CreateBy,
			)

			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			if balance > 0 {
				argsStockSummary = append(argsStockSummary,
					data.WarehouseCode,
					"-",
					"-",
					detail.Source,
					detail.ItemCode,
					detail.Batch,
					0,
					balance,
					0,
					data.CreateBy,
					data.Date,
				)
				argsStockMovement = append(argsStockMovement,
					"Stock Adjustment",
					data.DocNo,
					detail.DNo,
					"N",
					data.Date,
					data.WarehouseCode,
					detail.Source,
					detail.ItemCode,
					detail.Batch,
					0,
					balance,
					0,
					data.Remark,
					data.CreateBy,
					data.Date,
				)
			} else {
				argsStockSummary = append(argsStockSummary,
					data.WarehouseCode,
					"-",
					"-",
					detail.Source,
					detail.ItemCode,
					detail.Batch,
					0,
					0,
					(0 - balance),
					data.CreateBy,
					data.Date,
				)
				argsStockMovement = append(argsStockMovement,
					"Stock Adjustment",
					data.DocNo,
					detail.DNo,
					"N",
					data.Date,
					data.WarehouseCode,
					detail.Source,
					detail.ItemCode,
					detail.Batch,
					0,
					0,
					(0 - balance),
					data.Remark,
					data.CreateBy,
					data.Date,
				)
			}

			// history of stock
			placeholdersHistory = append(placeholdersHistory, "(?, ?, ?, ?, ?, ?)")
			argsHistory = append(argsHistory,
				detail.ItemCode,
				detail.Batch,
				detail.Source,
				"N",
				data.CreateBy,
				data.Date,
			)
		}

		// insert detail
		query += strings.Join(placeholders, ",") + ";"
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Error insert detail: %+v", err)
			return nil, fmt.Errorf("error Insert Detail: %w", err)
		}

		// insert stock summary
		queryStockSummary += strings.Join(placeholdersStockSummary, ",") + `;`
		if _, err = tx.ExecContext(ctx, queryStockSummary, argsStockSummary...); err != nil {
			log.Printf("Error insert stock summary: %+v", err)
			return nil, fmt.Errorf("error Insert Stock Summary: %w", err)
		}

		// insert stock movement
		queryStockMovement += strings.Join(placeholdersStockMovement, ",") + ";"
		if _, err = tx.ExecContext(ctx, queryStockMovement, argsStockMovement...); err != nil {
			log.Printf("Error insert stock movement: %+v", err)
			return nil, fmt.Errorf("error Insert Stock Movement: %w", err)
		}

		// insert history
		queryHistory += strings.Join(placeholdersHistory, ",") + ";"
		if _, err = tx.ExecContext(ctx, queryHistory, argsHistory...); err != nil {
			log.Printf("Error insert history of stock: %+v", err)
			return nil, fmt.Errorf("error Insert History of Stock: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}


// func (t *TblStockAdjustRepository) Create(ctx context.Context, data *tblstockadjustmenthdr.Create) (*tblstockadjustmenthdr.Create, error) {
// 	// Mulai transaksi
// 	tx, err := t.DB.BeginTxx(ctx, nil)
// 	if err != nil {
// 		log.Printf("Failed to start transaction: %+v", err)
// 		return nil, fmt.Errorf("error starting transaction: %w", err)
// 	}

// 	// Pastikan rollback selalu dijalankan jika terjadi error
// 	var txErr error
// 	defer func() {
// 		if txErr != nil {
// 			if rbErr := tx.Rollback(); rbErr != nil {
// 				log.Printf("Failed to rollback transaction: %+v", rbErr)
// 			}
// 		}
// 	}()

// 	countDetail := len(data.Details)
// 	var args []interface{}

// 	query := `INSERT INTO tblstockadjustmenthdr (
// 		DocNo,
// 		DocDt,
// 		WhsCode,
// 		Remark,
// 		CreateDt,
// 		CreateBy
// 	) VALUES (?, ?, ?, ?, ?, ?);`
// 	args = []interface{}{
// 		data.DocNo,
// 		data.Date,
// 		data.WarehouseCode,
// 		data.Remark,
// 		data.CreateDate,
// 		data.CreateBy,
// 	}

// 	_, err = tx.ExecContext(ctx, query, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to insert header: %w", err)
// 	}

// 	if countDetail > 0 {
// 		// Insert ke tabel detail
// 		query = `INSERT INTO tblstockadjustmentdtl (
// 			DocNo,
// 			DNo,
// 			ItCode,
// 			BatchNo,
// 			Lot,
// 			Bin,
// 			Qty,
// 			Qty2,
// 			Qty3,
// 			QtyActual,
// 			QtyActual2,
// 			QtyActual3,
// 			CreateDt,
// 			CreateBy
// 		) VALUES `
// 		var placeholders []string
// 		var whenClauses []string
// 		var movementValues []string
// 		args = args[:0]
// 		var argsClauses, argsMov []interface{}

// 		for _, detail := range data.Details {
// 			placeholders = append(placeholders, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			args = append(args,
// 				data.DocNo,
// 				detail.DNo,
// 				detail.ItemCode,
// 				detail.Batch,
// 				"-",
// 				"-",
// 				detail.StockSystem,
// 				detail.StockSystem,
// 				detail.StockSystem,
// 				detail.StockActual,
// 				detail.StockActual,
// 				detail.StockActual,
// 				data.CreateDate,
// 				data.CreateBy,
// 			)

// 			whenClauses = append(whenClauses, `WHEN WhsCode = ? AND ItCode = ? THEN ?`)
// 			argsClauses = append(argsClauses, data.WarehouseCode, detail.ItemCode, detail.StockActual)

// 			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			argsMov = append(argsMov,
// 				"Stock Adjustment",
// 				data.DocNo,
// 				detail.DNo,
// 				data.Date,
// 				data.WarehouseCode,
// 				detail.Source,
// 				detail.ItemCode,
// 				detail.Batch,
// 				detail.StockActual,
// 				detail.StockActual,
// 				detail.StockActual,
// 				data.Remark,
// 				data.CreateBy,
// 				data.CreateDate,
// 			)
// 		}

// 		query += strings.Join(placeholders, ",")
// 		_, err = tx.ExecContext(ctx, query, args...)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to insert details: %w", err)
// 		}

// 		// Update ke tabel stock summary
// 		querySum := `UPDATE tblstocksummary 
// 				SET
// 					Qty2 = CASE ` + strings.Join(whenClauses, " ") + `
// 					ELSE Qty2
// 					END,
// 					LastUpBy = ?,
// 					LastUpDt = ?
// 				WHERE WhsCode = ?;`
// 		argsClauses = append(argsClauses, data.CreateBy, data.CreateDate, data.WarehouseCode)

// 		// Insert ke tabel stock movement
// 		queryMov := `INSERT INTO tblstockmovement (
// 			DocType,
// 			DocNo,
// 			DNo,
// 			DocDt,
// 			WhsCode,
// 			Source,
// 			ItCode,
// 			BatchNo,
// 			Qty,
// 			Qty2,
// 			Qty3,
// 			Remark,
// 			CreateBy,
// 			CreateDt
// 		) VALUES `

// 		_, err = tx.ExecContext(ctx, querySum, argsClauses...)
// 		if err != nil {
// 			return nil, fmt.Errorf("error executing update query stock summary: %w", err)
// 		}

// 		queryMov += strings.Join(movementValues, ",")
// 		_, err = tx.ExecContext(ctx, queryMov, argsMov...)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to insert stock movement: %w", err)
// 		}
// 	}

// 	// Commit transaksi
// 	if err := tx.Commit(); err != nil {
// 		log.Printf("Failed to commit transaction: %+v", err)
// 		return nil, fmt.Errorf("error committing transaction: %w", err)
// 	}

// 	return data, nil
// }