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
	sharedfunc "gitlab.com/ayaka/internal/domain/shared/sharedFunc"
	"gitlab.com/ayaka/internal/domain/tblpurchasematerialreceive"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblPurchaseMaterialReceiveRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblPurchaseMaterialReceiveRepository) Create(ctx context.Context, data *tblpurchasematerialreceive.Create) (*tblpurchasematerialreceive.Create, error) {
	query := `INSERT INTO tblpurchasematerialreceivehdr 
	(
		DocNo,
		DocDt,
		WhsCode,
		VendorCode,
		SiteCode,
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
		data.WhsCode,
		data.VendorCode,
		data.SiteCode,
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
		query = `INSERT INTO tblpurchasematerialreceivedtl (
			DocNo,
			DNo,
			CancelInd,
			PurchaseOrderDocNo,
			PurchaseOrderDNo,
			ItCode,
			BatchNo,
			Source,
			OutstandingQty,
			PurchaseQty,
			InventoryQty,
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
		queryOrder := `INSERT INTO tblorderreport (
			DocNo,
			VendorCode,
			ItCode,
			Source,
			BatchNo,
			Qty,
			Price,
			TaxRate,
			DocDt
		) VALUES `
		var placeholdersOrder []string
		var argsOrder []interface{}

		// purchase order detail
		queryUpdatePurchaseOrderDtl := `UPDATE tblpurchaseorderdtl
			SET SuccessInd = CASE 
		`
		var whensPurchaseOrderDtl []string
		var wheresPurchaseOrderDtl []string
		var argsPurchaseOrderDtl, argsInPurchaseOrderDtl []interface{}

		// material request
		var whensOpenIndMatReq []string
		var wheresMatReq []string
		var argsOpenIndMatReq, argsInMatReq []interface{}

		for i, detail := range data.Details {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				fmt.Sprintf("%03d", i+1),
				"N",
				detail.PurchaseOrderDocNo,
				detail.PurchaseOrderDNo,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				detail.OutstandingQty,
				detail.PurchaseQty,
				0,
				detail.Remark,
				data.CreateDt,
				data.CreateBy,
			)

			// purchase order
			whensPurchaseOrderDtl = append(whensPurchaseOrderDtl, "WHEN DocNo = ? AND DNo = ? THEN 'Y'")
			argsPurchaseOrderDtl = append(argsPurchaseOrderDtl, detail.PurchaseOrderDocNo, detail.PurchaseOrderDNo)
			// tuples purchase order
			wheresPurchaseOrderDtl = append(wheresPurchaseOrderDtl, "(?, ?)")
			argsInPurchaseOrderDtl = append(argsInPurchaseOrderDtl, detail.PurchaseOrderDocNo, detail.PurchaseOrderDNo)

			whensOpenIndMatReq = append(whensOpenIndMatReq, "WHEN po.DocNo = ? AND po.DNo = ? THEN 'Y'")
			argsOpenIndMatReq = append(argsOpenIndMatReq, detail.PurchaseOrderDocNo, detail.PurchaseOrderDNo)
			// tuples material req
			wheresMatReq = append(wheresMatReq, "(?, ?)")
			argsInMatReq = append(argsInMatReq, detail.PurchaseOrderDocNo, detail.PurchaseOrderDNo)

			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockSummary = append(argsStockSummary,
				data.WhsCode,
				"-",
				"-",
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.PurchaseQty,
				0,
				data.CreateBy,
				data.Date,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Purchase Material Receive",
				data.DocNo,
				detail.DNo,
				detail.CancelInd,
				data.Date,
				data.WhsCode,
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.PurchaseQty,
				0,
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

			// order report
			placeholdersOrder = append(placeholdersOrder, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsOrder = append(argsOrder,
				data.DocNo,
				data.VendorCode,
				detail.ItCode,
				detail.Source,
				detail.BatchNo,
				detail.PurchaseQty,
				0,
				0,
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

		// update purchase order detail
		queryUpdatePurchaseOrderDtl += strings.Join(whensPurchaseOrderDtl, "\n") + "\nELSE CancelInd END\n"
		queryUpdatePurchaseOrderDtl += "WHERE (DocNo, Dno) IN (" + strings.Join(wheresPurchaseOrderDtl, ", ") + ")"
		argsPurchaseOrderDtl = append(argsPurchaseOrderDtl, argsInPurchaseOrderDtl...)

		log.Printf("Query update Purchase Order dtl: %s", queryUpdatePurchaseOrderDtl)
		log.Printf("Args: %+v", argsPurchaseOrderDtl)
		if _, err = tx.ExecContext(ctx, queryUpdatePurchaseOrderDtl, argsPurchaseOrderDtl...); err != nil {
			log.Printf("Error update purchase order detail detail batch: %+v", err)
			return nil, fmt.Errorf("error update purchase order detail detail: %w", err)
		}

		// update material req
		queryUpdateMatReqDtl := `UPDATE tblmaterialrequestdtl mr
			JOIN tblpurchaseorderreqdtl por ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			JOIN tblpurchaseorderdtl po ON po.PurchaseOrderReqDocNo = por.DocNo AND po.PurchaseOrderReqDNo = por.DNo
			SET 
				mr.OpenInd = CASE
					` + strings.Join(whensOpenIndMatReq, "\n") + `
					ELSE mr.OpenInd 
				END
			WHERE (po.DocNo, po.DNo) IN (` + strings.Join(wheresMatReq, ", ") + `)
		`
		fullArgsMatReq := []interface{}{}
		fullArgsMatReq = append(fullArgsMatReq, argsOpenIndMatReq...) // untuk CASE OpenInd
		fullArgsMatReq = append(fullArgsMatReq, argsInMatReq...)      // untuk WHERE IN

		fmt.Println("query mat req: ", queryUpdateMatReqDtl)
		fmt.Println("args: ", fullArgsMatReq)
		if _, err = tx.ExecContext(ctx, queryUpdateMatReqDtl, fullArgsMatReq...); err != nil {
			log.Printf("Error update material request detail batch: %+v", err)
			return nil, fmt.Errorf("error update material request detail: %w", err)
		}

		// insert order
		queryOrder += strings.Join(placeholdersOrder, ",") + ";"
		if _, err = tx.ExecContext(ctx, queryOrder, argsOrder...); err != nil {
			log.Printf("Error insert Order of stock: %+v", err)
			return nil, fmt.Errorf("error Insert Order of Stock: %w", err)
		}

		// update purchase order header
		queryUpdateHeader := `UPDATE tblpurchaseorderhdr h
			SET Status = CASE 
			-- Semua Qty diterima dan tidak dibatalkan
			WHEN (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
			) = (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
				AND d.Qty = (
					SELECT COALESCE(SUM(r.PurchaseQty), 0)
					FROM tblpurchasematerialreceivedtl r
					WHERE r.PurchaseOrderDocNo = d.DocNo
					AND r.PurchaseOrderDNo = d.DNo
					AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
				)
			) AND (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
			) > 0
			THEN 'Success'

			-- Sebagian diterima → Partial
			WHEN EXISTS (
				SELECT 1 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo 
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
				AND (
					SELECT COALESCE(SUM(r.PurchaseQty), 0)
					FROM tblpurchasematerialreceivedtl r
					WHERE r.PurchaseOrderDocNo = d.DocNo
					AND r.PurchaseOrderDNo = d.DNo
					AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
				) > 0
			)
			THEN 'Partial'

			-- Semua PMR dicancel atau belum ada penerimaan
			WHEN (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
				AND (
					SELECT COALESCE(SUM(r.PurchaseQty), 0)
					FROM tblpurchasematerialreceivedtl r
					WHERE r.PurchaseOrderDocNo = d.DocNo
					AND r.PurchaseOrderDNo = d.DNo
					AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
				) = 0
			) = (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
			)
			THEN 'Outstanding'

			ELSE h.Status
			END
			WHERE h.DocNo IN (?);
		`

		docNos := make([]string, 0)
		for _, d := range data.Details {
			if d.PurchaseOrderDocNo != "" {
				docNos = append(docNos, d.PurchaseOrderDocNo)
			}
		}
		docNos = sharedfunc.UniqueStringSlice(docNos) // hilangkan duplikat

		// gunakan sqlx.In agar support slice pada IN clause
		queryUpdateHeader, argsHeader, err := sqlx.In(queryUpdateHeader, docNos)
		if err != nil {
			log.Printf("sqlx.In error: %+v", err)
			return nil, fmt.Errorf("failed to build IN clause: %w", err)
		}

		queryUpdateHeader = tx.Rebind(queryUpdateHeader)

		if _, err := tx.ExecContext(ctx, queryUpdateHeader, argsHeader...); err != nil {
			log.Printf("Error batch update purchase order header status: %+v", err)
			return nil, fmt.Errorf("error batch update header PO status: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblPurchaseMaterialReceiveRepository) Fetch(ctx context.Context, doc, warehouse, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhs := "%" + warehouse + "%"
	searchVendor := "%" + vendor + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*)
		FROM tblpurchasematerialreceivehdr 
		WHERE DocNo LIKE ? AND WhsCode LIKE ? AND VendorCode LIKE ?`
	args = append(args, searchDoc, searchWhs, searchVendor)

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

	var data []*tblpurchasematerialreceive.Read
	args = []interface{}{searchDoc, searchWhs, searchVendor}

	query := `SELECT
				t.DocNo,
				t.DocDt,
				t.WhsCode,
				w.WhsName,
				t.VendorCode,
				v.VendorName,
				t.SiteCode,
				s.SiteName,
				t.Remark
			FROM tblpurchasematerialreceivehdr t
			JOIN tblwarehouse w ON t.WhsCode = w.WhsCode
			JOIN tblvendorhdr v ON t.VendorCode = v.VendorCode
			LEFT JOIN tblsite s ON t.SiteCode = s.SiteCode
			WHERE t.DocNo LIKE ? AND t.WhsCode LIKE ? AND t.VendorCode LIKE ?`

	if startDate != "" && endDate != "" {
		query += " AND t.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         []*tblpurchasematerialreceive.Read{},
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch direct purchase receive: %w", err)
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
			Data:         []*tblpurchasematerialreceive.Read{},
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	log.Println("DocNos:", docsNo)

	detailQuery := `SELECT 
				d.DocNo,
				d.DNo,
				d.CancelInd,
				d.PurchaseOrderDocNo,
				d.PurchaseOrderDNo,
				d.ItCode,
				i.ItName,
				u.UomName,
				d.BatchNo,
				d.Source,
				d.OutstandingQty,
				d.PurchaseQty,
				d.Remark
			FROM tblpurchasematerialreceivedtl d
			LEFT JOIN tblitem i ON d.ItCode = i.ItCode
			LEFT JOIN tbluom u ON i.PurchaseUomCode = u.UomCode
			WHERE d.DocNo IN (?)`

	query, args, err := sqlx.In(detailQuery, docsNo)
	if err != nil {
		return nil, fmt.Errorf("error preparing detail query: %w", err)
	}
	query = t.DB.Rebind(query)

	var details []*tblpurchasematerialreceive.Detail
	if err := t.DB.SelectContext(ctx, &details, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching details: %w", err)
	}

	log.Printf("Detail count: %d\n", len(details))

	// Kelompokkan detail berdasarkan DocNo
	detailMap := make(map[string][]tblpurchasematerialreceive.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	for _, d := range data {
		if dtl, ok := detailMap[d.DocNo]; ok {
			d.Details = dtl
		}
	}
	log.Printf("DetailMap: %+v\n", detailMap)
	for _, d := range data {
		log.Printf("DocNo: %s, Injected %d details\n", d.DocNo, len(detailMap[d.DocNo]))
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

func (t *TblPurchaseMaterialReceiveRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblpurchasematerialreceive.Read) (*tblpurchasematerialreceive.Read, error) {
	var resultDetail sql.Result
	var rowsAffectedDtl int64

	var placeholders, placeholdersStockSummary, placeholdersEdit, inTuples []string
	var args, argsStockSummary, argsEdit, argsIn, argsInSum []interface{}

	var err error

	if len(data.Details) == 0 {
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

	for _, detail := range data.Details {
		// detail cancel
		placeholders = append(placeholders, ` WHEN DocNo = ? AND DNo = ? THEN ? `)
		args = append(args, data.DocNo, detail.DNo, detail.CancelInd)

		// min qty
		placeholdersStockSummary = append(placeholdersStockSummary, ` WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty2 - ?) `)
		argsStockSummary = append(argsStockSummary, data.WhsCode, detail.Source, detail.ItCode, detail.PurchaseQty)

		// edit cancel status on history of stock and stock movement and order report
		placeholdersEdit = append(placeholdersEdit, ` WHEN  ItCode = ? AND Source = ? AND BatchNo = ? THEN ? `)
		argsEdit = append(argsEdit, detail.ItCode, detail.Source, detail.BatchNo, detail.CancelInd)

		inTuples = append(inTuples, "(?, ?, ?)")
		argsIn = append(argsIn, detail.ItCode, detail.Source, detail.BatchNo)
		argsInSum = append(argsInSum, data.WhsCode, detail.Source, detail.ItCode)
	}

	query := `UPDATE tblpurchasematerialreceivedtl
		SET CancelInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE CancelInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update direct purchase rcv dtl: %+v", err)
		return nil, fmt.Errorf("error updating direct purchase rcv dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for direct purchase rcv dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for direct purchase rcv dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update stock summary
	query = `UPDATE tblstocksummary
		SET Qty2 = CASE
			` + strings.Join(placeholdersStockSummary, " ") + `
			ELSE Qty2
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

	// Update order report
	query = `UPDATE tblorderreport
		SET CancelInd = CASE
			` + strings.Join(placeholdersEdit, " ") + `
			ELSE CancelInd
		END
		WHERE (ItCode, Source, BatchNo) IN (` + strings.Join(inTuples, ",") + `)
	`
	fmt.Println("Query order report: ", query)
	fmt.Println("args order report: ", argsEdit)
	if _, err := tx.ExecContext(ctx, query, argsEdit...); err != nil {
		log.Printf("Failed to update order report: %+v", err)
		return nil, fmt.Errorf("error updating order report: %w", err)
	}

	// update purchase order header
	queryUpdateHeader := `UPDATE tblpurchaseorderhdr h
			SET Status = CASE
			-- ✅ Semua quantity sudah diterima dari PMR (tidak dibatalkan)
			WHEN (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo 
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
			) = (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
				AND d.Qty = (
					SELECT COALESCE(SUM(r.PurchaseQty), 0)
					FROM tblpurchasematerialreceivedtl r
					WHERE r.PurchaseOrderDocNo = d.DocNo
					AND r.PurchaseOrderDNo = d.DNo
					AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
				)
			)
			THEN 'Success'

			-- ❗ Jika total qty = 0 (semua PMR dicancel atau belum input) → Outstanding
			WHEN (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
				AND (
					SELECT COALESCE(SUM(r.PurchaseQty), 0)
					FROM tblpurchasematerialreceivedtl r
					WHERE r.PurchaseOrderDocNo = d.DocNo
					AND r.PurchaseOrderDNo = d.DNo
					AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
				) = 0
			) = (
				SELECT COUNT(*) 
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
			)
			THEN 'Outstanding'

			-- 🔄 Jika sebagian sudah diterima → Partial
			WHEN EXISTS (
				SELECT 1
				FROM tblpurchaseorderdtl d
				WHERE d.DocNo = h.DocNo
				AND (d.CancelInd IS NULL OR d.CancelInd != 'Y')
				AND (
					SELECT COALESCE(SUM(r.PurchaseQty), 0)
					FROM tblpurchasematerialreceivedtl r
					WHERE r.PurchaseOrderDocNo = d.DocNo
					AND r.PurchaseOrderDNo = d.DNo
					AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
				) > 0
			)
			THEN 'Partial'

			ELSE h.Status
			END
			WHERE h.DocNo IN (?);
		`
	docNos := make([]string, 0)
	for _, d := range data.Details {
		if d.PurchaseOrderDocNo != "" {
			docNos = append(docNos, d.PurchaseOrderDocNo)
		}
	}
	docNos = sharedfunc.UniqueStringSlice(docNos) // hilangkan duplikat

	// gunakan sqlx.In agar support slice pada IN clause
	queryUpdateHeader, argsHeader, err := sqlx.In(queryUpdateHeader, docNos)
	if err != nil {
		log.Printf("sqlx.In error: %+v", err)
		return nil, fmt.Errorf("failed to build IN clause: %w", err)
	}

	queryUpdateHeader = tx.Rebind(queryUpdateHeader)

	if _, err := tx.ExecContext(ctx, queryUpdateHeader, argsHeader...); err != nil {
		log.Printf("Error batch update purchase order header status: %+v", err)
		return nil, fmt.Errorf("error batch update header PO status: %w", err)
	}

	// --- Update SuccessInd di purchaseorderdtl ---
	// Ambil pasangan (PurchaseOrderDocNo, PurchaseOrderDNo) dari data.Details
	var poDetailTuples [][2]string
	for _, d := range data.Details {
		if d.PurchaseOrderDocNo != "" && d.PurchaseOrderDNo != "" {
			poDetailTuples = append(poDetailTuples, [2]string{d.PurchaseOrderDocNo, d.PurchaseOrderDNo})
		}

	}
	poDetailTuples = sharedfunc.UniqueTupleSlice(poDetailTuples) // pastikan tidak duplikat

	// Siapkan query
	inTuples = inTuples[:0] // reset inTuples
	var argsTuple []interface{}
	for _, pair := range poDetailTuples {
		inTuples = append(inTuples, "(?, ?)")
		argsTuple = append(argsTuple, pair[0], pair[1])
	}

	queryUpdatePODetail := fmt.Sprintf(`
	UPDATE tblpurchaseorderdtl d
	SET SuccessInd = CASE 
		-- ❌ Semua PMR dicancel → SuccessInd = 'N'
		WHEN (
			SELECT COUNT(*)
			FROM tblpurchasematerialreceivedtl r
			WHERE r.PurchaseOrderDocNo = d.DocNo
			  AND r.PurchaseOrderDNo = d.DNo
			  AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
		) = 0 THEN 'N'

		-- ✅ Masih ada PMR aktif → SuccessInd = 'Y'
		ELSE 'Y'
	END
	WHERE (d.DocNo, d.DNo) IN (%s)
`, strings.Join(inTuples, ", "))

	if _, err := tx.ExecContext(ctx, queryUpdatePODetail, argsTuple...); err != nil {
		log.Printf("Error update SuccessInd in tblpurchaseorderdtl: %+v", err)
		return nil, fmt.Errorf("error update PO detail SuccessInd: %w", err)
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "DirectPurchaseReceive", lastUpDate); err != nil {
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

func (t *TblPurchaseMaterialReceiveRepository) Reporting(ctx context.Context, doc, warehouse, vendor, item, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhs := "%" + warehouse + "%"
	searchVendor := "%" + vendor + "%"
	searchItem := "%" + item + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) FROM (
			SELECT
				pmrd.DocNo, pmrd.DNo
			FROM tblpurchasematerialreceivehdr pmrh
			JOIN tblpurchasematerialreceivedtl pmrd
				ON pmrh.DocNo = pmrd.DocNo
			JOIN tblvendorhdr v
				ON pmrh.VendorCode = v.VendorCode
			JOIN tblitem i
				ON pmrd.ItCode = i.ItCode
			WHERE pmrh.DocNo LIKE ? AND pmrh.WhsCode LIKE ? AND v.VendorName LIKE ? AND i.ItName LIKE ? AND pmrd.CancelInd = 'N'
			`
	args = append(args, searchDoc, searchWhs, searchVendor, searchItem)

	if startDate != "" && endDate != "" {
		countQuery += " AND DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	countQuery += ` GROUP BY pmrd.DocNo, pmrd.DNo
		) AS grouped`

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

	var data []*tblpurchasematerialreceive.Reporting

	query := `SELECT
				pmrd.DocNo,
				pmrh.DocDt,
				w.WhsName,
				pmrd.PurchaseOrderDocNo AS PODoc,
				v.VendorName,
				i.ItName,
				pmrd.BatchNo,
				pmrd.PurchaseQTY AS Qty,
				u.UomName,
				pmrh.Remark AS DocumentRemark,
				pmrd.Remark AS ItemRemark
			FROM tblpurchasematerialreceivehdr pmrh
			JOIN tblpurchasematerialreceivedtl pmrd
				ON pmrh.DocNo = pmrd.DocNo
			JOIN tblvendorhdr v
				ON pmrh.VendorCode = v.VendorCode
			JOIN tblitem i
				ON pmrd.ItCode = i.ItCode
			JOIN tblwarehouse w
				ON pmrh.WhsCode = w.WhsCode
			JOIN tbluom u
			ON i.PurchaseUOMCode = u.UomCode
			WHERE pmrh.DocNo LIKE ? AND pmrh.WhsCode LIKE ? AND v.VendorName LIKE ? AND i.ItName LIKE ? AND pmrd.CancelInd = 'N'
			`

	if startDate != "" && endDate != "" {
		query += " AND t.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " GROUP BY pmrd.DocNo, pmrd.DNo LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         []*tblpurchasematerialreceive.Read{},
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch direct purchase receive: %w", err)
	}

	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
		detail.Date = share.FormatDate(detail.Date)
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
