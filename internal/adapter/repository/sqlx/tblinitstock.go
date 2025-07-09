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
		detail.TblDate = share.FormatDate(detail.Date)
		detail.Date = share.ToDatePicker(detail.Date)
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
				i.ItName, i.ItCodeInternal, d.Source,
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
		for i := range detailMap[h.DocNo] {
			detailMap[h.DocNo][i].Disabled = detailMap[h.DocNo][i].Cancel.ToBool()
			count += float32(detailMap[h.DocNo][i].Quantity)
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
		d.UPrice AS UPrice,
		d.Source AS Source
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
	query := `INSERT INTO tblstockinitialhdr 
	(
		DocNo,
		DocDt,
		WhsCode,
		CurCode,
		ExcRate,
		Remark,
		CreateDt,
		CreateBy
	) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		data.WarehouseCode,
		data.CurrencyCode,
		data.Rate,
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

	if len(data.Detail) > 0 {
		// detail query
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
			Source,
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

		for i, detail := range data.Detail {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				fmt.Sprintf("%03d", i+1),
				"N",
				detail.ItemCode,
				detail.Batch,
				"-",
				detail.Quantity,
				detail.Quantity,
				detail.Quantity,
				detail.Price,
				detail.Source,
				data.CreateDate,
				data.CreateBy,
			)

			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockSummary = append(argsStockSummary,
				data.WarehouseCode,
				"-",
				"-",
				detail.Source,
				detail.ItemCode,
				detail.Batch,
				detail.Quantity,
				0,
				0,
				data.CreateBy,
				data.Date,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Initial Stock",
				data.DocNo,
				detail.DNo,
				detail.Cancel,
				data.Date,
				data.WarehouseCode,
				detail.Source,
				detail.ItemCode,
				detail.Batch,
				detail.Quantity,
				0,
				0,
				data.Remark,
				data.CreateBy,
				data.Date,
			)

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

func (t *TblInitStockRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblinitialstock.Detail) (*tblinitialstock.Detail, error) {
	var resultDetail sql.Result
	var rowsAffectedDtl int64

	var placeholders, placeholdersStockSummary, placeholdersEdit, inTuples []string
	var args, argsStockSummary, argsEdit, argsIn, argsInSum []interface{}

	var err error

	if len(data.Detail) == 0 {
		return data, nil
	}

	// transaction begin
	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// Pastikan rollback dipanggil jika transaksi tidak berhasil
	defer func() {
		if err != nil {
			log.Printf("Transaction rollback due to error: %+v", err)
			// Rollback jika error terjadi
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %+v", rbErr)
			}
		}
	}()

	for _, detail := range data.Detail {
		// detail cancel
		placeholders = append(placeholders, ` WHEN DocNo = ? AND DNo = ? THEN ? `)
		args = append(args, data.DocNo, detail.DNo, detail.Cancel)

		// min qty
		placeholdersStockSummary = append(placeholdersStockSummary, ` WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty - ?) `)
		argsStockSummary = append(argsStockSummary, data.WarehouseCode, detail.Source, detail.ItemCode, detail.Quantity)

		// edit cancel status on history of stock and stock movement
		placeholdersEdit = append(placeholdersEdit, ` WHEN  ItCode = ? AND Source = ? AND BatchNo = ? THEN ? `)
		argsEdit = append(argsEdit, detail.ItemCode, detail.Source, detail.Batch, detail.Cancel)
		
		inTuples = append(inTuples, "(?, ?, ?)")
		argsIn = append(argsIn, detail.ItemCode, detail.Source, detail.Batch)
		argsInSum = append(argsInSum, data.WarehouseCode, detail.Source, detail.ItemCode)
	}

	query := `UPDATE tblstockinitialdtl
		SET CancelInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE CancelInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update stock initial dtl: %+v", err)
		return nil, fmt.Errorf("error updating stock initial dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for stock initial dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for stock initial dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update stock summary
	query = `UPDATE tblstocksummary
		SET Qty = CASE
			` + strings.Join(placeholdersStockSummary, " ") + `
			ELSE Qty
		END
		WHERE (WhsCode, Source, ItCode) IN (` + strings.Join(inTuples, ",") + `)
	`
	argsStockSummary = append(argsStockSummary, argsInSum...)

	if _, err := tx.ExecContext(ctx, query, argsStockSummary...); err != nil {
		log.Printf("Failed to update stock summary: %+v", err)
		return nil, fmt.Errorf("error updating stock summary: %w", err)
	}

	// Update history of stock
	query = `UPDATE tblhistoryofstock
		SET CancelInd = CASE
			` + strings.Join(placeholdersEdit, " ") + `
			ELSE CancelInd
		END
		WHERE (ItCode, Source, BatchNo) IN (` + strings.Join(inTuples, ",") + `)
	`

	fmt.Println("Query history of stock: ", query)
	argsEdit = append(argsEdit, argsIn...)
	fmt.Println("args history of stock: ", argsEdit)
	if _, err := tx.ExecContext(ctx, query, argsEdit...); err != nil {
		log.Printf("Failed to update history of stock: %+v", err)
		return nil, fmt.Errorf("error updating history of stock: %w", err)
	}

	// Update stock movement
	query = `UPDATE tblstockmovement
		SET CancelInd = CASE
			` + strings.Join(placeholdersEdit, " ") + `
			ELSE CancelInd
		END
		WHERE (ItCode, Source, BatchNo) IN (` + strings.Join(inTuples, ",") + `)
	`
	fmt.Println("Query stock sum: ", query)
	fmt.Println("args stock sum: ", argsEdit)
	if _, err := tx.ExecContext(ctx, query, argsEdit...); err != nil {
		log.Printf("Failed to update stock movement: %+v", err)
		return nil, fmt.Errorf("error updating stock movement: %w", err)
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "StockInitial", lastUpDate); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error inserting log activity: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}



// func (t *TblInitStockRepository) Create(ctx context.Context, data *tblinitialstock.Create) (*tblinitialstock.Create, error) {
// 	countDetail := len(data.Detail)
// 	var args []interface{}

// 	// Mulai transaksi
// 	tx, err := t.DB.BeginTxx(ctx, nil)
// 	if err != nil {
// 		log.Printf("Failed to start transaction: %+v", err)
// 		return nil, fmt.Errorf("error starting transaction: %w", err)
// 	}

// 	// Pastikan rollback selalu dijalankan jika terjadi error
// 	defer func() {
// 		if err != nil {
// 			if rbErr := tx.Rollback(); rbErr != nil {
// 				log.Printf("Failed to rollback transaction: %+v", rbErr)
// 			}
// 		}
// 	}()

// 	// Insert ke tabel header
// 	query := `INSERT INTO tblstockinitialhdr (
// 		DocNo,
// 		DocDt,
// 		WhsCode,
// 		CurCode,
// 		ExcRate,
// 		Remark,
// 		CreateDt,
// 		CreateBy
// 	) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
// 	args = []interface{}{
// 		data.DocNo,
// 		data.Date,
// 		data.WarehouseCode,
// 		data.CurrencyCode,
// 		data.Rate,
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
// 		query = `INSERT INTO tblstockinitialdtl (
// 			DocNo,
// 			DNo,
// 			CancelInd,
// 			ItCode,
// 			BatchNo,
// 			Lot,
// 			Qty,
// 			Qty2,
// 			Qty3,
// 			UPrice,
// 			CreateDt,
// 			CreateBy
// 		) VALUES `
// 		var placeholders []string
// 		args = args[:0]

// 		for _, detail := range data.Detail {
// 			placeholders = append(placeholders, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			args = append(args,
// 				data.DocNo,
// 				detail.DNo,
// 				detail.Cancel,
// 				detail.ItemCode,
// 				detail.Batch,
// 				"-",
// 				detail.Quantity,
// 				detail.Quantity,
// 				detail.Quantity,
// 				detail.Price,
// 				data.CreateDate,
// 				data.CreateBy,
// 			)
// 		}

// 		query += strings.Join(placeholders, ",")
// 		_, err = tx.ExecContext(ctx, query, args...)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to insert details: %w", err)
// 		}

// 		// Insert ke tabel stock movement
// 		query = `INSERT INTO tblstockmovement (
// 			DocType,
// 			DocNo,
// 			DNo,
// 			CancelInd,
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
// 		var movementValues []string
// 		args = args[:0]

// 		for _, detail := range data.Detail {
// 			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			args = append(args,
// 				data.DocType,
// 				data.DocNo,
// 				detail.DNo,
// 				detail.Cancel,
// 				data.Date,
// 				data.WarehouseCode,
// 				detail.Source,
// 				detail.ItemCode,
// 				detail.Batch,
// 				detail.Quantity,
// 				detail.Quantity,
// 				detail.Quantity,
// 				data.Remark,
// 				data.CreateBy,
// 				data.CreateDate,
// 			)
// 		}

// 		query += strings.Join(movementValues, ",")
// 		_, err = tx.ExecContext(ctx, query, args...)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to insert stock movement: %w", err)
// 		}

// 		// Insert ke tabel stock summary
// 		query = `INSERT INTO tblstocksummary (
// 			WhsCode,
// 			Lot,
// 			Bin,
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
// 		var summaryValues []string
// 		args = args[:0]

// 		for _, detail := range data.Detail {
// 			summaryValues = append(summaryValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			args = append(args,
// 				data.WarehouseCode,
// 				"-",
// 				"-",
// 				detail.ItemCode,
// 				detail.ItemCode,
// 				detail.Batch,
// 				detail.Quantity,
// 				0,
// 				0,
// 				data.Remark,
// 				data.CreateBy,
// 				data.CreateDate,
// 			)
// 		}

// 		query += strings.Join(summaryValues, ",") + `
// 			ON DUPLICATE KEY UPDATE
// 				Qty = Qty + VALUES(Qty),
// 				Qty2 = Qty2,
// 				Qty3 = Qty3,
// 				LastUpBy = VALUES(CreateBy),
// 				LastUpDt = VALUES(CreateDt);`
// 		_, err = tx.ExecContext(ctx, query, args...)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to insert stock summary: %w", err)
// 		}
// 	}

// 	// Commit transaksi
// 	if err := tx.Commit(); err != nil {
// 		log.Printf("Failed to commit transaction: %+v", err)
// 		return nil, fmt.Errorf("error committing transaction: %w", err)
// 	}

// 	return data, nil
// }

// func (t *TblInitStockRepository) Update(ctx context.Context, data, oldData *tblinitialstock.Detail, lastUpBy, lastUpDt string) (*tblinitialstock.Detail, error) {
// 	count := len(data.Detail)

// 	if count > 0 {
// 		var args, argsLog, argsMov []interface{}
// 		var whenClauses []string
// 		var placeholders []string
// 		var whenClausesMov []string

// 		status := false

// 		qLog := `INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES `

// 		for i := 0; i < count; i++ {
// 			if oldData.Detail[i].Cancel != booldatatype.FromBool(true) && oldData.Detail[i].Cancel != data.Detail[i].Cancel {
// 				whenClauses = append(whenClauses, `WHEN DNo = ? THEN ?`)
// 				args = append(args, data.Detail[i].DNo, data.Detail[i].Cancel)

// 				whenClausesMov = append(whenClausesMov, `WHEN DocNo = ? AND DNo = ? THEN ?`)
// 				argsMov = append(argsMov, data.DocNo, data.Detail[i].DNo, data.Detail[i].Cancel)

// 				placeholders = append(placeholders, (`(?, ?, ?, ?)`))
// 				argsLog = append(argsLog, lastUpBy, data.Detail[i].DNo, "StockInitialDtl", lastUpDt)
// 				status = true
// 			}
// 		}

// 		if len(whenClauses) == 0 {
// 			return nil, customerrors.ErrNoDataEdited
// 		}

// 		query := `UPDATE tblstockinitialdtl
// 				SET 
// 					CancelInd = CASE ` + strings.Join(whenClauses, " ") + `
// 					ELSE CancelInd
// 					END,
// 					LastUpDt = ?,
// 					LastUpBy = ?
// 				WHERE DocNo = ?;`
// 		args = append(args, lastUpDt, lastUpBy, data.DocNo)

// 		queryMov := `UPDATE tblstockmovement
// 				SET 
// 					CancelInd = CASE ` + strings.Join(whenClausesMov, " ") + `
// 					ELSE CancelInd
// 					END,
// 					LastUpDt = ?,
// 					LastUpBy = ?
// 				WHERE DocNo = ?;`
// 		argsMov = append(argsMov, lastUpDt, lastUpBy, data.DocNo)

// 		if status {
// 			qLog += `(?, ?, ?, ?)`
// 			qLog += ", " + strings.Join(placeholders, ", ")
// 			argsLog = append(argsLog, lastUpBy, data.DocNo, "StockInitial", lastUpDt)

// 			tx, err := t.DB.BeginTxx(ctx, nil)
// 			if err != nil {
// 				log.Printf("Failed to start transaction: %+v", err)
// 				return nil, fmt.Errorf("error starting transaction: %w", err)
// 			}

// 			// Pastikan rollback dipanggil jika transaksi tidak berhasil
// 			defer func() {
// 				if err != nil {
// 					// Rollback jika error terjadi
// 					if rbErr := tx.Rollback(); rbErr != nil {
// 						log.Printf("Failed to rollback transaction: %+v", rbErr)
// 					}
// 				}
// 			}()

// 			_, err = tx.ExecContext(ctx, query, args...)
// 			if err != nil {
// 				return nil, fmt.Errorf("error executing update query: %w", err)
// 			}

// 			_, err = tx.ExecContext(ctx, queryMov, argsMov...)
// 			if err != nil {
// 				return nil, fmt.Errorf("error executing update query for tblstockmovement: %w", err)
// 			}

// 			// Eksekusi log activity
// 			_, err = tx.ExecContext(ctx, qLog, argsLog...)
// 			if err != nil {
// 				log.Printf("Failed to execute update query log: %v", err)
// 				return nil, fmt.Errorf("error insert to log activity: %w", err)
// 			}

// 			// Commit transaksi jika semua query berhasil
// 			if err := tx.Commit(); err != nil {
// 				log.Printf("Failed to commit transaction: %+v", err)
// 				return nil, fmt.Errorf("error committing transaction: %w", err)
// 			}

// 			return data, nil
// 		}
// 	}
// 	return nil, customerrors.ErrNoDataEdited
// }
