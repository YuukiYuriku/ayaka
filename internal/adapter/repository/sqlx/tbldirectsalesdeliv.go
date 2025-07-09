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
	"gitlab.com/ayaka/internal/domain/tbldirectsalesdelivery"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDirectSalesDeliveryRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblDirectSalesDeliveryRepository) Create(ctx context.Context, data *tbldirectsalesdelivery.Create) (*tbldirectsalesdelivery.Create, error) {
	query := `INSERT INTO tbldirectsalesdelivhdr 
	(
		DocNo,
		DocDt,
		WhsCode,
		CustomerName,
		Address,
		CityCode,
		PostalCode,
		Phone,
		Email,
		Mobile,
		TaxCode,
		Remark,
		CreateDt,
		CreateBy
	) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		data.WhsCode,
		data.Customer,
		data.Address,
		data.CityCode,
		data.PostalCode,
		data.Phone,
		data.Email,
		data.Mobile,
		data.TaxCode,
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
		query = `INSERT INTO tbldirectsalesdelivdtl (
			DocNo,
			DNo,
			CancelInd,
			ItCode,
			BatchNo,
			Price,
			Stock,
			Qty,
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

		// history of stock
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
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				fmt.Sprintf("%03d", i+1),
				"N",
				detail.ItCode,
				detail.BatchNo,
				detail.Price,
				detail.Stock,
				detail.Qty,
				detail.Source,
				data.CreateDt,
				data.CreateBy,
			)

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
				0,
				detail.Qty,
				data.CreateBy,
				data.Date,
			)

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsStockMovement = append(argsStockMovement,
				"Direct Sales Delivery",
				data.DocNo,
				detail.DNo,
				detail.Cancel,
				data.Date,
				data.WhsCode,
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

func (t *TblDirectSalesDeliveryRepository) Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhs := "%" + warehouse + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tbldirectsalesdelivhdr WHERE DocNo LIKE ? AND WhsCode LIKE ?"
	args = append(args, searchDoc, searchWhs)

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

	var data []*tbldirectsalesdelivery.Read
	args = []interface{}{searchDoc, searchWhs}

	query := `SELECT
				t.DocNo,
				t.DocDt,
				t.WhsCode,
				w.WhsName,
				t.CustomerName,
				t.Address,
				t.CityCode,
				t.PostalCode,
				t.Phone,
				t.Email,
				t.Mobile,
				t.TaxCode,
				t.Remark
			FROM tbldirectsalesdelivhdr t
			JOIN tblwarehouse w ON t.WhsCode = w.WhsCode
			LEFT JOIN tblcity c ON t.CityCode = c.CityCode
			LEFT JOIN tbltax tx ON t.TaxCode = tx.TaxCode
			WHERE t.DocNo LIKE ? AND t.WhsCode LIKE ?`

	if startDate != "" && endDate != "" {
		query += " AND t.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         []*tbldirectsalesdelivery.Read{},
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
			Data:         []*tbldirectsalesdelivery.Read{},
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
				d.Price,
				d.Qty,
				d.Stock,
				d.Price,
				i.ItName,
				u.UomName
			FROM tbldirectsalesdelivdtl d
			JOIN tblitem i ON d.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUomCode = u.UomCode
			WHERE d.DocNo IN (?)`

	query, args, err := sqlx.In(detailQuery, docsNo)
	if err != nil {
		return nil, fmt.Errorf("error preparing detail query: %w", err)
	}
	query = t.DB.Rebind(query)

	var details []*tbldirectsalesdelivery.Detail
	if err := t.DB.SelectContext(ctx, &details, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching details: %w", err)
	}

	// Kelompokkan detail berdasarkan DocNo
	detailMap := make(map[string][]tbldirectsalesdelivery.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Hitung TotalQuantity dan GrandTotal + Pajak
	for _, h := range data {
		h.Details = detailMap[h.DocNo]

		var subtotal float32 = 0.0
		var taxRate float32 = 0.0
		var count float32 = 0.0

		// Ambil tax rate dari tbltax
		if h.TaxCode.String != "" {
			queryTax := `SELECT TaxRate FROM tbltax WHERE TaxCode = ?`
			if err := t.DB.GetContext(ctx, &taxRate, queryTax, h.TaxCode.String); err != nil {
				return nil, fmt.Errorf("error fetching tax rate for %s: %w", h.TaxCode.String, err)
			}
		}

		for i := range h.Details {
			d := &h.Details[i]
			subtotal += d.Price * d.Qty
			count += float32(d.Qty)
		}

		taxAmount := subtotal * taxRate
		h.TotalAmount = subtotal + taxAmount
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

func (t *TblDirectSalesDeliveryRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tbldirectsalesdelivery.Read) (*tbldirectsalesdelivery.Read, error) {
	var resultDetail sql.Result
	var rowsAffectedDtl int64

	var placeholders, placeholdersStockSummary, placeholdersEdit, inTuples, inTuplesSum []string
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

	existingCancels := make(map[string]booldatatype.BoolDataType)

	querySelect := `SELECT DNo, CancelInd FROM tbldirectsalesdelivdtl WHERE DocNo = ?`
	rows, err := t.DB.QueryxContext(ctx, querySelect, data.DocNo)
	if err != nil {
		return nil, fmt.Errorf("error fetching existing cancel status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dno string
		var cancel booldatatype.BoolDataType
		if err := rows.Scan(&dno, &cancel); err != nil {
			return nil, fmt.Errorf("error scanning cancel status: %w", err)
		}
		existingCancels[dno] = cancel
	}

	for _, detail := range data.Details {
		// detail cancel
		placeholders = append(placeholders, ` WHEN DocNo = ? AND DNo = ? THEN ? `)
		args = append(args, data.DocNo, detail.DNo, detail.Cancel)

		// stock sum
		prevCancel, exists := existingCancels[detail.DNo]
		if exists && !prevCancel.ToBool() && detail.Cancel.ToBool() {
			// Jika sebelumnya false dan sekarang true, maka kurangi Qty3
			placeholdersStockSummary = append(placeholdersStockSummary,
				` WHEN WhsCode = ? AND Source = ? AND ItCode = ? THEN (Qty3 - ?) `)
			argsStockSummary = append(argsStockSummary, data.WhsCode, detail.Source, detail.ItCode, detail.Qty)
			inTuplesSum = append(inTuplesSum, "(?, ?, ?)")
			argsInSum = append(argsInSum, data.WhsCode, detail.Source, detail.ItCode)
		}

		// edit cancel status on history of stock and stock movement
		placeholdersEdit = append(placeholdersEdit, ` WHEN  ItCode = ? AND Source = ? AND BatchNo = ? THEN ? `)
		argsEdit = append(argsEdit, detail.ItCode, detail.Source, detail.BatchNo, detail.Cancel)

		inTuples = append(inTuples, "(?, ?, ?)")
		argsIn = append(argsIn, detail.ItCode, detail.Source, detail.BatchNo)
	}

	query := `UPDATE tbldirectsalesdelivdtl
		SET CancelInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE CancelInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update direct sales delivery dtl: %+v", err)
		return nil, fmt.Errorf("error updating direct sales delivery dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for direct sales delivery dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for direct sales delivery dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update stock summary
	query = `UPDATE tblstocksummary
		SET Qty3 = CASE
			` + strings.Join(placeholdersStockSummary, " ") + `
			ELSE Qty3
		END
		WHERE (WhsCode, Source, ItCode) IN (` + strings.Join(inTuplesSum, ",") + `)
	`
	argsStockSummary = append(argsStockSummary, argsInSum...)

	fmt.Println("query sum: ", query)
	fmt.Println("args stock sum: ", argsStockSummary)

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
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "DirectSalesDelivery", lastUpDate); err != nil {
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
