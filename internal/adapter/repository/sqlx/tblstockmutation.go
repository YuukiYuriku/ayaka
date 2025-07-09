package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	// "github.com/jmoiron/sqlx"
	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"

	// "gitlab.com/ayaka/internal/domain/tblstockamutationdtl"
	"gitlab.com/ayaka/internal/domain/tblstockmutationdtl"
	"gitlab.com/ayaka/internal/domain/tblstockmutationhdr"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	"gitlab.com/ayaka/internal/pkg/customerrors"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockMutationRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblStockMutationRepository) Fetch(ctx context.Context, doc, warehouse, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	var args []interface{}
	var endQuery []string

	countQuery := `
			SELECT
				COUNT(d.DNo)
			FROM tblstockmutationhdr h
			JOIN tblstockmutationdtl d `
	// endQuery = append(endQuery, ` h.DocDt BETWEEN ? AND ?  `)
	// args = append(args, startDate, endDate)

	if warehouse != "" {
		endQuery = append(endQuery, ` h.WhsCode = ? `)
		args = append(args, warehouse)
	}
	if doc != "" {
		doc = "%" + doc + "%"
		endQuery = append(endQuery, ` h.DocNo LIKE ? `)
		args = append(args, doc)
	}
	if batch != "" {
		batch = "%" + batch + "%"
		endQuery = append(endQuery, ` h.BatchNo LIKE ? `)
		args = append(args, batch)
	}

	if len(endQuery) > 0 {
		countQuery += " WHERE " + strings.Join(endQuery, " AND ")
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

	var data []*tblstockmutationhdr.Fetch
	query := `
		SELECT
			h.DocNo,
			h.CancelInd,
			h.DocDt,
			w.WhsName,
			d.FromTo,
			i.ItName,
			d.BatchNo,
			d.Qty,
			u.UomName
		FROM tblstockmutationhdr h
		JOIN tblstockmutationdtl d ON d.DocNo = h.DocNo
		JOIN tblitem i ON i.ItCode = d.ItCode
		JOIN tblwarehouse w ON w.WhsCode = h.WhsCode
		JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode`

	if len(endQuery) > 0 {
		query += " WHERE " + strings.Join(endQuery, " AND ")
	}
	query += ` LIMIT ? OFFSET ? `
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblstockmutationhdr.Fetch, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch stock mutation: %w", err)
	}

	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
		detail.TblDate = share.FormatDate(detail.Date)
		detail.Date = share.ToDatePicker(detail.Date)
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

func (t *TblStockMutationRepository) Detail(ctx context.Context, docNo string) (*tblstockmutationhdr.Detail, error) {
	var data tblstockmutationhdr.Detail
	query := `
		SELECT
			h.DocNo AS DocNo,
			h.DocDt AS DocDt,
			h.WhsCode AS WhsCode,
			w.WhsName AS WhsName,
			h.CancelReason AS CancelReason,
			h.CancelInd AS CancelInd,
			h.Remark AS Remark
		FROM tblstockmutationhdr h
		JOIN tblwarehouse w ON h.WhsCode = w.WhsCode
		WHERE h.DocNo = ?;
	`

	if err := t.DB.GetContext(ctx, &data, query, docNo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, fmt.Errorf("error Get header: %w", err)
	}

	var detailFrom []tblstockmutationdtl.Detail
	var detailTo []tblstockmutationdtl.Detail
	query = `
		SELECT
			d.DocNo,
			d.DNo,
			d.ItCode,
			i.ItName,
			d.BatchNo,
			d.Source,
			COALESCE(s.Stock, 0) AS Stock,
			d.Qty,
			u.UomName
		FROM tblstockmutationdtl d
		JOIN tblstockmutationhdr h ON h.DocNo = d.DocNo
		JOIN tblitem i ON i.ItCode = d.ItCode
		JOIN tbluom u ON u.UomCode = i.PurchaseUOMCode
		LEFT JOIN (
			SELECT
				ss.WhsCode,
				ss.ItCode,
				ss.BatchNo,
				SUM(ss.Qty + ss.Qty2 - ss.Qty3) AS Stock
			FROM tblstocksummary ss
			GROUP BY ss.WhsCode, ss.ItCode, ss.BatchNo
		) s ON s.WhsCode = h.WhsCode
			AND s.ItCode = d.ItCode
			AND (s.BatchNo = d.BatchNo)
		WHERE d.DocNo = ?
	`

	fromQuery := query + ` AND d.FromTo = "From";`
	if err := t.DB.SelectContext(ctx, &detailFrom, fromQuery, docNo); err != nil {
		return nil, fmt.Errorf("error Get detail from: %w", err)
	}

	toQuery := query + ` AND d.FromTo = "To";`
	if err := t.DB.SelectContext(ctx, &detailTo, toQuery, docNo); err != nil {
		return nil, fmt.Errorf("error Get detail to: %w", err)
	}

	data.FromArray = detailFrom
	data.ToArray = detailTo
	data.Date = share.ToDatePicker(data.Date)

	return &data, nil
}

func (t *TblStockMutationRepository) Create(ctx context.Context, data *tblstockmutationhdr.Create) (*tblstockmutationhdr.Create, error) {
	query := `
		INSERT INTO tblstockmutationhdr (
			DocNo,
			DocDt,
			WhsCode,
			BatchNo,
			Source,
			Remark,
			CreateBy,
			CreateDt
		) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.DocDate,
		data.WarehouseCode,
		data.BatchNo,
		data.Source,
		data.Remark,
		data.CreateBy,
		data.CreateDate,
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

	if len(data.FromArray) > 0 && len(data.ToArray) > 0 {
		// details
		query = `
			INSERT INTO tblstockmutationdtl (
				DocNo,
				DNo,
				ItCode,
				BatchNo,
				Source,
				Qty,
				FromTo,
				CreateDt,
				CreateBy
			) VALUES 
		`
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

		for _, detail := range data.FromArray {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				detail.Qty,
				"From",
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
				detail.ItCode,
				detail.BatchNo,
				0,
				0,
				detail.Qty,
				data.CreateBy,
				data.DocDate,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Stock Mutation (From)",
				data.DocNo,
				detail.DNo,
				"N",
				data.DocDate,
				data.WarehouseCode,
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				0,
				detail.Qty,
				data.Remark,
				data.CreateBy,
				data.DocDate,
			)

			// history of stock
			placeholdersHistory = append(placeholdersHistory, "(?, ?, ?, ?, ?, ?)")
			argsHistory = append(argsHistory,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				"N",
				data.CreateBy,
				data.DocDate,
			)
		}

		for _, detail := range data.ToArray {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				detail.Qty,
				"To",
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
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.Qty,
				0,
				data.CreateBy,
				data.DocDate,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Stock Mutation (To)",
				data.DocNo,
				detail.DNo,
				"N",
				data.DocDate,
				data.WarehouseCode,
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.Qty,
				0,
				data.Remark,
				data.CreateBy,
				data.DocDate,
			)

			// history of stock
			placeholdersHistory = append(placeholdersHistory, "(?, ?, ?, ?, ?, ?)")
			argsHistory = append(argsHistory,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				"N",
				data.CreateBy,
				data.DocDate,
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

func (t *TblStockMutationRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblstockmutationhdr.Detail) (*tblstockmutationhdr.Detail, error) {
	var resultDetail sql.Result
	var rowsAffectedDtl int64

	var placeholdersStockSummary, placeholdersStockSummary2, placeholdersEdit, inTuples, inTuplesSum2, inTuplesSum []string
	var args, argsStockSummary, argsStockSummary2, argsEdit, argsIn, argsInSum, argsInSum2 []interface{}

	var err error

	if len(data.FromArray) == 0 || len(data.ToArray) == 0 {
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

	query := `UPDATE tblstockmutationhdr SET
		CancelInd = ?,
		CancelReason = ?
	WHERE DocNo = ?`
	args = append(args, data.Cancel, data.CancelReason, data.DocNo)

	for _, detail := range data.FromArray {
		// min qty
		placeholdersStockSummary = append(placeholdersStockSummary, `WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty3 - ?) `)
		argsStockSummary = append(argsStockSummary, data.WarehouseCode, detail.Source, detail.ItemCode, detail.Quantity)

		// edit cancel status on history of stock and stock movement
		placeholdersEdit = append(placeholdersEdit, ` WHEN  ItCode = ? AND Source = ? AND BatchNo = ? THEN ? `)
		argsEdit = append(argsEdit, detail.ItemCode, detail.Source, detail.Batch, data.Cancel)
		
		inTuples = append(inTuples, "(?, ?, ?)")
		inTuplesSum = append(inTuplesSum, "(?, ?, ?)")
		argsIn = append(argsIn, detail.ItemCode, detail.Source, detail.Batch)
		argsInSum = append(argsInSum, data.WarehouseCode, detail.Source, detail.ItemCode)
	}

	for _, detail := range data.ToArray {
		// min qty
		placeholdersStockSummary2 = append(placeholdersStockSummary2, `WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty2 - ?) `)
		argsStockSummary2 = append(argsStockSummary2, data.WarehouseCode, detail.Source, detail.ItemCode, detail.Quantity)

		// edit cancel status on history of stock and stock movement
		placeholdersEdit = append(placeholdersEdit, ` WHEN  ItCode = ? AND Source = ? AND BatchNo = ? THEN ? `)
		argsEdit = append(argsEdit, detail.ItemCode, detail.Source, detail.Batch, data.Cancel)
		
		inTuples = append(inTuples, "(?, ?, ?)")
		argsIn = append(argsIn, detail.ItemCode, detail.Source, detail.Batch)
		inTuplesSum2 = append(inTuplesSum2, "(?, ?, ?)")
		argsInSum2 = append(argsInSum2, data.WarehouseCode, detail.Source, detail.ItemCode)
	}

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update stock mutation dtl: %+v", err)
		return nil, fmt.Errorf("error updating stock mutation dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for stock mutation dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for stock mutation dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update stock summary FROM
	query = `UPDATE tblstocksummary
		SET Qty3 = CASE
			` + strings.Join(placeholdersStockSummary, " ") + `
			ELSE Qty3
		END
		WHERE (WhsCode, Source, ItCode) IN (` + strings.Join(inTuplesSum, ",") + `)
	`
	argsStockSummary = append(argsStockSummary, argsInSum...)

	fmt.Println("Query update stock summary FROM:", query)
	fmt.Println("Args:", argsStockSummary)	

	if _, err := tx.ExecContext(ctx, query, argsStockSummary...); err != nil {
		log.Printf("Failed to update stock summary: %+v", err)
		return nil, fmt.Errorf("error updating stock summary: %w", err)
	}

	// Update stock summary TO
	query = `UPDATE tblstocksummary
		SET Qty2 = CASE
			` + strings.Join(placeholdersStockSummary2, " ") + `
			ELSE Qty2
		END
		WHERE (WhsCode, Source, ItCode) IN (` + strings.Join(inTuplesSum2, ",") + `)
	`
	argsStockSummary2 = append(argsStockSummary2, argsInSum2...)

	fmt.Println("Query update stock summary FROM:", query)
	fmt.Println("Args:", argsStockSummary2)

	if _, err := tx.ExecContext(ctx, query, argsStockSummary2...); err != nil {
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
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "StockMutation", lastUpDate); err != nil {
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

// func (t *TblStockMutationRepository) Create(ctx context.Context, data *tblstockmutationhdr.Create) (*tblstockmutationhdr.Create, error) {
// 	// Mulai transaksi
// 	tx, err := t.DB.BeginTxx(ctx, nil)
// 	if err != nil {
// 		log.Printf("Failed to start transaction: %+v", err)
// 		return nil, fmt.Errorf("error starting transaction: %w", err)
// 	}

// 	var txErr error
// 	defer func() {
// 		if txErr != nil {
// 			if rbErr := tx.Rollback(); rbErr != nil {
// 				log.Printf("Failed to rollback transaction: %+v", rbErr)
// 			}
// 		}
// 	}()

// 	var args []interface{}

// 	query := `
// 		INSERT INTO tblstockmutationhdr (
// 			DocNo,
// 			DocDt,
// 			WhsCode,
// 			BatchNo,
// 			Source,
// 			Remark,
// 			CreateBy,
// 			CreateDt
// 		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
// 	`
// 	args = append(args,
// 		data.DocNo,
// 		data.DocDate,
// 		data.WarehouseCode,
// 		data.BatchNo,
// 		data.Source,
// 		data.Remark,
// 		data.CreateBy,
// 		data.CreateDate,
// 	)

// 	// insert header
// 	_, err = tx.ExecContext(ctx, query, args...)
// 	if err != nil {
// 		txErr = fmt.Errorf("failed to insert header stock mutation: %w", err)
// 		return nil, txErr
// 	}

// 	if len(data.FromArray) > 0 && len(data.ToArray) > 0 {
// 		query = `
// 			INSERT INTO tblstockmutationdtl (
// 				DocNo,
// 				DNo,
// 				ItCode,
// 				BatchNo,
// 				Source,
// 				Qty,
// 				FromTo,
// 				CreateDt,
// 				CreateBy
// 			) VALUES 
// 		`
// 		queryInsertSum := `
// 			INSERT INTO tblstocksummary (
// 				WhsCode,
// 				Lot,
// 				Bin,
// 				Source,
// 				ItCode,
// 				BatchNo,
// 				Qty,
// 				Qty2,
// 				Qty3,
// 				Remark,
// 				CreateBy,
// 				CreateDt
// 			) VALUES 
// 		`

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

// 		var placeholders []string
// 		var whenClauses []string
// 		var movementValues []string
// 		var summaryValues []string

// 		args = args[:0]
// 		var argsClauses, argsMov, argsSummary []interface{}

// 		// set insert and value for from array
// 		for _, detail := range data.FromArray {
// 			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
// 			args = append(args,
// 				data.DocNo,
// 				detail.DNo,
// 				detail.ItCode,
// 				detail.BatchNo,
// 				detail.Source,
// 				detail.Qty,
// 				"From",
// 				data.CreateDate,
// 				data.CreateBy,
// 			)

// 			whenClauses = append(whenClauses, ` WHEN WhsCode = ? AND ItCode = ? THEN ?`)
// 			argsClauses = append(argsClauses, data.WarehouseCode, detail.ItCode, detail.Qty)

// 			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			argsMov = append(argsMov,
// 				"Stock Mutation",
// 				data.DocNo,
// 				detail.DNo,
// 				data.DocDate,
// 				data.WarehouseCode,
// 				detail.Source,
// 				detail.ItCode,
// 				data.BatchNo,
// 				detail.Qty,
// 				detail.Qty,
// 				detail.Qty,
// 				data.Remark,
// 				data.CreateBy,
// 				data.CreateDate,
// 			)
// 		}

// 		// set insert and value for to array
// 		for _, detail := range data.ToArray {
// 			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
// 			args = append(args,
// 				detail.DocNo,
// 				detail.DNo,
// 				detail.ItCode,
// 				detail.BatchNo,
// 				detail.Source,
// 				detail.Qty,
// 				"To",
// 				data.CreateDate,
// 				data.CreateBy,
// 			)

// 			summaryValues = append(summaryValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			argsSummary = append(argsSummary,
// 				data.WarehouseCode,
// 				"-",
// 				"-",
// 				detail.ItCode,
// 				detail.ItCode,
// 				detail.BatchNo,
// 				detail.Qty,
// 				0,
// 				0,
// 				data.Remark,
// 				data.CreateBy,
// 				data.CreateDate,
// 			)

// 			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
// 			argsMov = append(argsMov,
// 				"Stock Mutation",
// 				data.DocNo,
// 				detail.DNo,
// 				data.DocDate,
// 				data.WarehouseCode,
// 				detail.Source,
// 				detail.ItCode,
// 				data.BatchNo,
// 				detail.Qty,
// 				detail.Qty,
// 				detail.Qty,
// 				data.Remark,
// 				data.CreateBy,
// 				data.CreateDate,
// 			)
// 		}

// 		// insert detail
// 		query += strings.Join(placeholders, ",")
// 		_, err = tx.ExecContext(ctx, query, args...)
// 		if err != nil {
// 			txErr = fmt.Errorf("failed to insert detail stock mutation: %w", err)
// 			return nil, txErr
// 		}

// 		// insert stock sum
// 		queryInsertSum += strings.Join(summaryValues, ",") + `
// 			ON DUPLICATE KEY UPDATE
// 				Qty = Qty + VALUES(Qty),
// 				Qty2 = Qty2,
// 				Qty3 = Qty3,
// 				LastUpBy = VALUES(CreateBy),
// 				LastUpDt = VALUES(CreateDt);`
// 		_, err = tx.ExecContext(ctx, queryInsertSum, argsSummary...)
// 		if err != nil {
// 			txErr = fmt.Errorf("failed to insert stock summary: %w", err)
// 			return nil, txErr
// 		}

// 		// insert movement
// 		queryMov += strings.Join(movementValues, ",")
// 		_, err = tx.ExecContext(ctx, queryMov, argsMov...)
// 		if err != nil {
// 			txErr = fmt.Errorf("failed to insert stock movement: %w", err)
// 			return nil, txErr
// 		}

// 		// Update ke tabel stock summary
// 		querySum := `UPDATE tblstocksummary 
// 				SET
// 					Qty3 = CASE ` + strings.Join(whenClauses, " ") + `
// 					ELSE Qty3
// 					END,
// 					LastUpBy = ?,
// 					LastUpDt = ?
// 				WHERE WhsCode = ?;`
// 		argsClauses = append(argsClauses, data.CreateBy, data.CreateDate, data.WarehouseCode)

// 		_, err = tx.ExecContext(ctx, querySum, argsClauses...)
// 		if err != nil {
// 			txErr = fmt.Errorf("failed to update stock summary: %w", err)
// 			return nil, txErr
// 		}
// 	}

// 	// Commit transaksi
// 	if err := tx.Commit(); err != nil {
// 		log.Printf("Failed to commit transaction: %+v", err)
// 		return nil, fmt.Errorf("error committing transaction: %w", err)
// 	}

// 	return data, nil
// }