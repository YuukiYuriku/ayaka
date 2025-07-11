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
	"gitlab.com/ayaka/internal/domain/tbldirectmaterialreceive"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDirectMaterialReceiveRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblDirectMaterialReceiveRepository) Create(ctx context.Context, data *tbldirectmaterialreceive.Create) (*tbldirectmaterialreceive.Create, error) {
	query := `INSERT INTO tbldirectmaterialreceivehdr 
	(
		DocNo,
		DocDt,
		WhsCodeFrom,
		WhsCodeTo,
		Remark,
		CreateDt,
		CreateBy
	) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		data.WhsCodeFrom,
		data.WhsCodeTo,
		data.Remark,
		data.CreateDt,
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
		query = `INSERT INTO tbldirectmaterialreceivedtl (
			DocNo,
			DNo,
			CancelInd,
			ItCode,
			BatchNo,
			SenderStock,
			Qty,
			Source,
			Remark,
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

		// order report
		queryOrder := `INSERT INTO tbltransferbetweenwhs (
			DocNo,
			DocDt,
			CancelInd,
			WhsFrom,
			WhsTo,
			ItCode,
			BatchNo,
			Qty,
			Remark,
			CreateBy,
			CreateDt
		) VALUES `
		var placeholdersOrder []string
		var argsOrder []interface{}

		for i, detail := range data.Details {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				fmt.Sprintf("%03d", i+1),
				"N",
				detail.ItCode,
				detail.BatchNo,
				detail.SenderStock,
				detail.Qty,
				detail.Source,
				detail.Remark,
				data.CreateDt,
				data.CreateBy,
			)

			//////////////////////////// TO WAREHOUSE
			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockSummary = append(argsStockSummary,
				data.WhsCodeTo,
				"-",
				"-",
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.Qty,
				0,
				data.CreateBy,
				data.Date,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Direct Material Receive",
				data.DocNo,
				detail.DNo,
				"N",
				data.Date,
				data.WhsCodeTo,
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.Qty,
				0,
				data.Remark,
				data.CreateBy,
				data.Date,
			)

			//////////////////////////// FROM WAREHOUSE
			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockSummary = append(argsStockSummary,
				data.WhsCodeFrom,
				"-",
				"-",
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				0,
				detail.Qty,
				data.CreateBy,
				data.Date,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Direct Material Receive",
				data.DocNo,
				detail.DNo,
				"N",
				data.Date,
				data.WhsCodeFrom,
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				0,
				detail.Qty,
				data.Remark,
				data.CreateBy,
				data.Date,
			)

			// history of stock
			placeholdersHistory = append(placeholdersHistory, "(?, ?, ?, ?, ?, ?)")
			argsHistory = append(argsHistory,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				"N",
				data.CreateBy,
				data.Date,
			)

			// transfer between whs report
			placeholdersOrder = append(placeholdersOrder, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsOrder = append(argsOrder,
				data.DocNo,
				data.Date,
				detail.Cancel,
				data.WhsCodeFrom,
				data.WhsCodeTo,
				detail.ItCode,
				detail.BatchNo,
				detail.Qty,
				detail.Remark,
				data.CreateBy,
				data.CreateDt,
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

		// insert order
		queryOrder += strings.Join(placeholdersOrder, ",") + ";"
		if _, err = tx.ExecContext(ctx, queryOrder, argsOrder...); err != nil {
			log.Printf("Error insert Order of stock: %+v", err)
			return nil, fmt.Errorf("error Insert Order of Stock: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblDirectMaterialReceiveRepository) Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhs := "%" + warehouseFrom + "%"
	searchToWhs := "%" + warehouseTo + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tbldirectmaterialreceivehdr WHERE DocNo LIKE ? AND WhsCodeFrom LIKE ? AND WhsCodeTo LIKE ?"
	args = append(args, searchDoc, searchWhs, searchToWhs)

	if startDate != "" && endDate != "" {
		countQuery += " AND DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, args...); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	var totalPages, offset int
	if param != nil {
		totalPages, offset = pagination.CountPagination(param, totalRecords)
	} else {
		param = &pagination.PaginationParam{
			PageSize: totalRecords,
			Page:     1,
		}
		totalPages = 1
		offset = 0
	}

	var data []*tbldirectmaterialreceive.Read
	args = []interface{}{searchDoc, searchWhs, searchToWhs}

	query := `SELECT
				t.DocNo,
				t.DocDt,
				t.WhsCodeFrom,
				w.WhsName AS WhsNameFrom,
				t.WhsCodeTo,
				w.WhsName AS WhsNameTo,
				t.Remark
			FROM tbldirectmaterialreceivehdr t
			JOIN tblwarehouse w ON t.WhsCodeFrom = w.WhsCode
			JOIN tblwarehouse w2 ON t.WhsCodeTo = w2.WhsCode
			WHERE t.DocNo LIKE ? AND t.WhsCodeFrom LIKE ? AND t.WhsCodeTo LIKE ?`

	if startDate != "" && endDate != "" {
		query += " AND t.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         []*tbldirectmaterialreceive.Read{},
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch direct material receive: %w", err)
	}

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
			Data:         []*tbldirectmaterialreceive.Read{},
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	detailQuery := `SELECT 
				d.DocNo,
				d.DNo,
				d.CancelInd,
				d.ItCode,
				d.BatchNo,
				d.Source,
				d.SenderStock,
				d.Qty,
				i.ItName,
				u.UomName
			FROM tbldirectmaterialreceivedtl d
			JOIN tblitem i ON d.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUomCode = u.UomCode
			WHERE d.DocNo IN (?)`

	query, args, err := sqlx.In(detailQuery, docsNo)
	if err != nil {
		return nil, fmt.Errorf("error preparing detail query: %w", err)
	}
	query = t.DB.Rebind(query)

	var details []*tbldirectmaterialreceive.Detail
	if err := t.DB.SelectContext(ctx, &details, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching details: %w", err)
	}

	// Kelompokkan detail berdasarkan DocNo
	detailMap := make(map[string][]tbldirectmaterialreceive.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Hitung TotalQuantity dan GrandTotal + Pajak
	for _, h := range data {
		h.Details = detailMap[h.DocNo]
		var count float32 = 0.0

		for i := range h.Details {
			d := &h.Details[i]
			count += float32(d.Qty)
		}

		h.TotalQuantity = count
	}

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

func (t *TblDirectMaterialReceiveRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tbldirectmaterialreceive.Read) (*tbldirectmaterialreceive.Read, error) {
	if len(data.Details) == 0 {
		return data, nil
	}

	var (
		err                                        error
		resultDetail                               sql.Result
		rowsAffectedDtl                            int64
		placeholders, placeholdersStockSummary2    []string
		placeholdersStockSummary, placeholdersEdit []string
		placeholdersTransfer, placeholdersMov      []string
		inTuplesSummaryTo, inTuplesSummaryFrom     []string
		inTuplesMovHist, inTupleMov                []string
		args, argsStockSummary, argsStockSummary2  []interface{}
		argsEdit, argsTransfer, argsMov            []interface{}
		argsInSumTo, argsInSumFrom                 []interface{}
		argsInHistMov, argsTuplesMov               []interface{}
	)

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			log.Printf("Transaction rollback due to error: %+v", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %+v", rbErr)
			}
		}
	}()

	for _, detail := range data.Details {
		placeholders = append(placeholders, " WHEN DocNo = ? AND DNo = ? THEN ? ")
		args = append(args, data.DocNo, detail.DNo, detail.Cancel)

		// Qty2: FROM WhsCodeTo (barang keluar dari transfer tujuan)
		placeholdersStockSummary = append(placeholdersStockSummary, " WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty2 - ?) ")
		argsStockSummary = append(argsStockSummary, data.WhsCodeTo, detail.Source, detail.ItCode, detail.Qty)
		inTuplesSummaryTo = append(inTuplesSummaryTo, "(?, ?, ?)")
		argsInSumTo = append(argsInSumTo, data.WhsCodeTo, detail.Source, detail.ItCode)

		// Qty3: FROM WhsCodeFrom (barang masuk ke transfer asal)
		placeholdersStockSummary2 = append(placeholdersStockSummary2, " WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty3 - ?) ")
		argsStockSummary2 = append(argsStockSummary2, data.WhsCodeFrom, detail.Source, detail.ItCode, detail.Qty)
		inTuplesSummaryFrom = append(inTuplesSummaryFrom, "(?, ?, ?)")
		argsInSumFrom = append(argsInSumFrom, data.WhsCodeFrom, detail.Source, detail.ItCode)

		// History (tanpa WhsCode)
		placeholdersEdit = append(placeholdersEdit, " WHEN ItCode = ? AND Source = ? AND BatchNo = ? THEN ? ")
		argsEdit = append(argsEdit, detail.ItCode, detail.Source, detail.BatchNo, detail.Cancel)
		inTuplesMovHist = append(inTuplesMovHist, "(?, ?, ?)")
		argsInHistMov = append(argsInHistMov, detail.ItCode, detail.Source, detail.BatchNo)

		// Stock Movement
		placeholdersMov = append(placeholdersMov, " WHEN WhsCode = ? AND ItCode = ? AND Source = ? AND BatchNo = ? THEN ? ")
		argsMov = append(argsMov, data.WhsCodeFrom, detail.ItCode, detail.Source, detail.BatchNo, detail.Cancel)
		inTupleMov = append(inTupleMov, "(?, ?, ?, ?)")
		argsTuplesMov = append(argsTuplesMov, data.WhsCodeFrom, detail.ItCode, detail.Source, detail.BatchNo)

		placeholdersMov = append(placeholdersMov, " WHEN WhsCode = ? AND ItCode = ? AND Source = ? AND BatchNo = ? THEN ? ")
		argsMov = append(argsMov, data.WhsCodeTo, detail.ItCode, detail.Source, detail.BatchNo, detail.Cancel)
		inTupleMov = append(inTupleMov, "(?, ?, ?, ?)")
		argsTuplesMov = append(argsTuplesMov, data.WhsCodeTo, detail.ItCode, detail.Source, detail.BatchNo)

		// Transfer antar gudang
		placeholdersTransfer = append(placeholdersTransfer, " WHEN DocNo = ? AND ItCode = ? AND BatchNo = ? THEN ? ")
		argsTransfer = append(argsTransfer, data.DocNo, detail.ItCode, detail.BatchNo, detail.Cancel)
	}

	// Update detail
	query := `UPDATE tbldirectmaterialreceivedtl
		SET CancelInd = CASE ` + strings.Join(placeholders, " ") + ` ELSE CancelInd END
		WHERE DocNo = ?`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("error updating direct material rcv dtl: %w", err)
	}
	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		return nil, fmt.Errorf("error getting rows affected for direct material rcv dtl: %w", err)
	}
	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update Qty2
	query = `UPDATE tblstocksummary SET Qty2 = CASE ` + strings.Join(placeholdersStockSummary, " ") + ` ELSE Qty2 END
		WHERE (WhsCode, Source, ItCode) IN (` + strings.Join(inTuplesSummaryTo, ",") + `)`
	argsStockSummary = append(argsStockSummary, argsInSumTo...)
	if _, err := tx.ExecContext(ctx, query, argsStockSummary...); err != nil {
		return nil, fmt.Errorf("error updating stock summary Qty2: %w", err)
	}

	// Update Qty3
	query = `UPDATE tblstocksummary SET Qty3 = CASE ` + strings.Join(placeholdersStockSummary2, " ") + ` ELSE Qty3 END
		WHERE (WhsCode, Source, ItCode) IN (` + strings.Join(inTuplesSummaryFrom, ",") + `)`
	argsStockSummary2 = append(argsStockSummary2, argsInSumFrom...)
	if _, err := tx.ExecContext(ctx, query, argsStockSummary2...); err != nil {
		return nil, fmt.Errorf("error updating stock summary Qty3: %w", err)
	}

	// Update history of stock
	// Build the correct number of "(?, ?, ?)" placeholders for the IN clause
	// numTuples := len(argsInHistMov) / 4 * 2 // Each detail adds 2 tuples (from and to)
	// inPlaceholders := make([]string, 0, numTuples)
	// for i := 0; i < numTuples; i++ {
	// 	inPlaceholders = append(inPlaceholders, "(?, ?, ?)")
	// }
	query = `UPDATE tblhistoryofstock SET CancelInd = CASE ` + strings.Join(placeholdersEdit, " ") + ` ELSE CancelInd END
		WHERE (ItCode, Source, BatchNo) IN (` + strings.Join(inTuplesMovHist, ",") + `)`
	argsEdit = append(argsEdit, argsInHistMov...)

	fmt.Println("query detail: ", query)
	fmt.Println("args detail: ", argsEdit)
	if _, err := tx.ExecContext(ctx, query, argsEdit...); err != nil {
		return nil, fmt.Errorf("error updating history of stock: %w", err)
	}

	// Update stock movement
	query = `UPDATE tblstockmovement SET CancelInd = CASE ` + strings.Join(placeholdersMov, " ") + ` ELSE CancelInd END
		WHERE (WhsCode, ItCode, Source, BatchNo) IN (` + strings.Join(inTupleMov, ",") + `)`
	
	argsMov = append(argsMov, argsTuplesMov...)
	if _, err := tx.ExecContext(ctx, query, argsMov...); err != nil {
		return nil, fmt.Errorf("error updating stock movement: %w", err)
	}

	// Update transfer antar gudang
	query = `UPDATE tbltransferbetweenwhs SET CancelInd = CASE ` + strings.Join(placeholdersTransfer, " ") + ` ELSE CancelInd END
		WHERE DocNo = ?`
	argsTransfer = append(argsTransfer, data.DocNo)
	if _, err := tx.ExecContext(ctx, query, argsTransfer...); err != nil {
		return nil, fmt.Errorf("error updating transfer: %w", err)
	}

	// Insert log
	query = `INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)`
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "DirectMaterialReceive", lastUpDate); err != nil {
		return nil, fmt.Errorf("error inserting log activity: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}
