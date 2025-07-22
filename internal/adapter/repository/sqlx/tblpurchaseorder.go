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
	"gitlab.com/ayaka/internal/domain/tblpurchaseorder"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblPurchaseOrderRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblPurchaseOrderRepository) Create(ctx context.Context, data *tblpurchaseorder.Create) (*tblpurchaseorder.Create, error) {
	query := `INSERT INTO tblpurchaseorderhdr
	(
		DocNo,
		DocDt,
		Status,
		VendorCode,
		ContactPersonDNo,
		Remark,
		TaxCode,
		CreateBy,
		CreateDt
	) VALUES `
	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		"Outstanding",
		data.VendorCode,
		data.ContactPersonDNo,
		data.Remark,
		data.TaxCode,
		data.CreateBy,
		data.CreateDt,
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
		query = `INSERT INTO tblpurchaseorderdtl
		(
			DocNo,
			DNo,
			CancelInd,
			SuccessInd,
			PurchaseOrderReqDocNo,
			PurchaseOrderReqDNo,
			ItCode,
			Qty,
			Total,
			Remark,
			CreateBy,
			CreateDt
		) VALUES `
		placeholders = placeholders[:0]
		args = args[:0]

		queryUpdatePurchaseOrderReqDtl := `UPDATE tblpurchaseorderreqdtl
			SET SuccessInd = CASE 
		`
		var whensPurchaseOrderReq []string
		var wheresPurchaseOrderReq []string
		var argsPurchaseOrderReq, argsInPurchaseOrderReq []interface{}

		for _, detail := range data.Details {
			detail.Total = detail.Qty * detail.UPrice
			// details
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				"N",
				"N",
				detail.PurchaseOrderReqDocNo,
				detail.PurchaseOrderReqDNo,
				detail.ItCode,
				detail.Qty,
				detail.Total,
				detail.Remark,
				data.CreateBy,
				data.CreateDt,
			)

			// purchase order req
			whensPurchaseOrderReq = append(whensPurchaseOrderReq, "WHEN DocNo = ? AND DNo = ? THEN 'Y'")
			argsPurchaseOrderReq = append(argsPurchaseOrderReq, detail.PurchaseOrderReqDocNo, detail.PurchaseOrderReqDNo)
			// tuples purchase order req
			wheresPurchaseOrderReq = append(wheresPurchaseOrderReq, "(?, ?)")
			argsInPurchaseOrderReq = append(argsInPurchaseOrderReq, detail.PurchaseOrderReqDocNo, detail.PurchaseOrderReqDNo)
		}

		// insert detail
		query += strings.Join(placeholders, ",") + ";"

		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Error insert detail: %+v", err)
			return nil, fmt.Errorf("error Insert Detail: %w", err)
		}

		// update purchase order req
		queryUpdatePurchaseOrderReqDtl += strings.Join(whensPurchaseOrderReq, "\n") + "\nELSE SuccessInd END\n"
		queryUpdatePurchaseOrderReqDtl += "WHERE (DocNo, DNo) IN (" + strings.Join(wheresPurchaseOrderReq, ", ") + ")"

		argsPurchaseOrderReq = append(argsPurchaseOrderReq, argsInPurchaseOrderReq...)
		if _, err = tx.ExecContext(ctx, queryUpdatePurchaseOrderReqDtl, argsPurchaseOrderReq...); err != nil {
			log.Printf("Error update purchase order req detail batch: %+v", err)
			return nil, fmt.Errorf("error update purchase order req detail: %w", err)
		}

	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblPurchaseOrderRepository) Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblpurchaseorderhdr WHERE DocNo LIKE ?"
	args = append(args, searchDoc)

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

	var data []*tblpurchaseorder.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.Status,
			i.VendorCode,
			v.VendorName,
			i.ContactPersonDNo AS DNo,
			i.TaxCode,
			i.Remark
			FROM tblpurchaseorderhdr i
			JOIN tblvendorhdr v ON i.VendorCode = v.VendorCode
			LEFT JOIN tblcontactvendordtl cv ON v.VendorCode = cv.VendorCode
			LEFT JOIN tbltax t ON i.TaxCode = t.TaxCode
			WHERE i.DocNo LIKE ? OR v.VendorName LIKE ?
			GROUP BY i.DocNo`
	args = append(args, searchDoc, searchDoc)

	if startDate != "" && endDate != "" {
		query += " AND i.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblpurchaseorder.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch purchase order: %w", err)
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
			Data:         make([]*tblpurchaseorder.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblpurchaseorder.Detail{}
	detailQuery := `SELECT
				d.DocNo,
				d.DNo,
				d.CancelInd,
				d.SuccessInd,
				d.PurchaseOrderReqDocNo,
				d.PurchaseOrderReqDNo,
				d.ItCode,
				i.ItName,
				d.Qty,
				c.CurName,
				vqd.Price,
				d.Total,
				vqh.DeliveryType,
				d.Remark
		FROM tblpurchaseorderdtl d
		JOIN tblitem i ON d.ItCode = i.ItCode
		JOIN tblpurchaseorderreqdtl por ON			
			d.PurchaseOrderReqDocNo = por.DocNo
			AND d.PurchaseOrderReqDNo = por.DNo
		JOIN tblvendorquotationhdr vqh ON
			por.VendorQTDocNo = vqh.DocNo
		JOIN tblvendorquotationdtl vqd ON
			por.VendorQTDocNo = vqd.DocNo
			AND por.VendorQTDNo = vqd.DNo
		JOIN tblcurrency c ON
			vqh.CurCode = c.CurCode
		WHERE d.DocNo IN (?);
	`
	query, args, err := sqlx.In(detailQuery, docsNo)
	if err != nil {
		return nil, fmt.Errorf("error preparing detail query: %w", err)
	}
	query = t.DB.Rebind(query)

	if err := t.DB.SelectContext(ctx, &details, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching details: %w", err)
	}

	// Kelompokkan detail berdasarkan DocNo
	detailMap := make(map[string][]tblpurchaseorder.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range data {
		h.Details = detailMap[h.DocNo]

		var subtotal float32 = 0.0
		var taxRate float32 = 0.0
		// Ambil tax rate dari tbltax
		if h.TaxCode.String != "" {
			queryTax := `SELECT TaxRate FROM tbltax WHERE TaxCode = ?`
			if err := t.DB.GetContext(ctx, &taxRate, queryTax, h.TaxCode.String); err != nil {
				return nil, fmt.Errorf("error fetching tax rate for %s: %w", h.TaxCode.String, err)
			}
		}
		h.TaxRate = taxRate

		for i := range h.Details {
			d := &h.Details[i]
			subtotal += d.UPrice * d.Qty
		}
		taxAmount := subtotal * taxRate
		h.GrandTotal = subtotal + taxAmount
		h.TotalTax = taxAmount
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

func (t *TblPurchaseOrderRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblpurchaseorder.Read) (*tblpurchaseorder.Read, error) {
	var resultDetail sql.Result
	var rowsAffectedDtl int64

	var placeholders []string
	var args []interface{}
	var argsWhere []interface{}

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

	query := `UPDATE tblpurchaseorderdtl
			SET CancelInd = CASE 
		`
	var wheresPurchaseOrder []string

	// material request
	var whensMatReq, whensOpenIndMatReq []string
	var wheresMatReq []string
	var argsMatReq, argsOpenIndMatReq, argsInMatReq []interface{}

	for _, detail := range data.Details {
		// detail cancel
		placeholders = append(placeholders, ` WHEN DocNo = ? AND DNo = ? THEN ? `)
		args = append(args, data.DocNo, detail.DNo, detail.CancelInd)
		wheresPurchaseOrder = append(wheresPurchaseOrder, "(?, ?)")
		argsWhere = append(argsWhere, data.DocNo, detail.DNo)

		// material req
		whensMatReq = append(whensMatReq, "WHEN por.DocNo = ? AND por.DNo = ? THEN ?")
		argsMatReq = append(argsMatReq, detail.PurchaseOrderReqDocNo, detail.PurchaseOrderReqDNo, booldatatype.FromBool(!detail.CancelInd.ToBool()))
		whensOpenIndMatReq = append(whensOpenIndMatReq, "WHEN por.DocNo = ? AND por.DNo = ? THEN 'Y'")
		argsOpenIndMatReq = append(argsOpenIndMatReq, detail.PurchaseOrderReqDocNo, detail.PurchaseOrderReqDNo)
		// tuples material req
		wheresMatReq = append(wheresMatReq, "(?, ?)")
		argsInMatReq = append(argsInMatReq, detail.PurchaseOrderReqDocNo, detail.PurchaseOrderReqDNo)
	}

	// update detail purchase order
	query += strings.Join(placeholders, "\n") + "\nELSE CancelInd END\n"
	query += "WHERE (DocNo, DNo) IN (" + strings.Join(wheresPurchaseOrder, ", ") + ")"

	args = append(args, argsWhere...)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update purchase order dtl: %+v", err)
		return nil, fmt.Errorf("error updating purchase order dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for purchase order dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for purchase order dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		fmt.Println("ga diubah")
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// update material req
	queryUpdateMatReqDtl := `UPDATE tblmaterialrequestdtl mr
			JOIN tblpurchaseorderreqdtl por ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			SET 
				mr.SuccessInd = CASE 
					` + strings.Join(whensMatReq, "\n") + `
					ELSE mr.SuccessInd 
				END,
				mr.OpenInd = CASE
					` + strings.Join(whensOpenIndMatReq, "\n") + `
					ELSE mr.OpenInd 
				END
			WHERE (por.DocNo, por.DNo) IN (` + strings.Join(wheresMatReq, ", ") + `)
		`
	fullArgsMatReq := []interface{}{}
	fullArgsMatReq = append(fullArgsMatReq, argsMatReq...)        // untuk CASE SuccessInd
	fullArgsMatReq = append(fullArgsMatReq, argsOpenIndMatReq...) // untuk CASE OpenInd
	fullArgsMatReq = append(fullArgsMatReq, argsInMatReq...)      // untuk WHERE IN

	if _, err = tx.ExecContext(ctx, queryUpdateMatReqDtl, fullArgsMatReq...); err != nil {
		log.Printf("Error update material request detail batch: %+v", err)
		return nil, fmt.Errorf("error update material request detail: %w", err)
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "PurchaseOrder", lastUpDate); err != nil {
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

func (t *TblPurchaseOrderRepository) GetPurchaseOrder(ctx context.Context, doc, vendor, item string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchVendor := "%" + vendor + "%"
	searchItem := "%" + item + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) AS total FROM (
			SELECT 
			d.DocNo,
			d.DNo,
			h.DocDt,
			d.ItCode,
			i.ItName,
			d.Qty - COALESCE((
				SELECT SUM(r.PurchaseQty)
				FROM tblpurchasematerialreceivedtl r
				WHERE r.PurchaseOrderDocNo = d.DocNo
				AND r.PurchaseOrderDNo = d.DNo
				AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
			), 0) AS OutstandingQty,
			u.UomName
		FROM 
			tblpurchaseorderdtl d
		JOIN 
			tblpurchaseorderhdr h ON d.DocNo = h.DocNo
		JOIN 
			tblitem i ON d.ItCode = i.ItCode
		JOIN 
			tbluom u ON i.PurchaseUomCode = u.UomCode
		WHERE 
			h.VendorCode LIKE ?
			AND (d.CancelInd != 'Y') AND i.ItName LIKE ?
			AND h.DocNo LIKE ?
		HAVING 
			OutstandingQty > 0
	) AS grouped`

	args = append(args, searchVendor, searchItem, searchDoc)

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

	var data []*tblpurchaseorder.GetPurchaseOrder
	args = args[:0]

	query := `SELECT 
		d.DocNo,
		d.DNo,
		h.DocDt,
		d.ItCode,
		i.ItName,
		d.Qty - COALESCE((
			SELECT SUM(r.PurchaseQty)
			FROM tblpurchasematerialreceivedtl r
			WHERE r.PurchaseOrderDocNo = d.DocNo
			AND r.PurchaseOrderDNo = d.DNo
			AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
		), 0) AS OutstandingQty,
		u.UomName
	FROM 
		tblpurchaseorderdtl d
	JOIN 
		tblpurchaseorderhdr h ON d.DocNo = h.DocNo
	JOIN 
		tblitem i ON d.ItCode = i.ItCode
	JOIN 
		tbluom u ON i.PurchaseUomCode = u.UomCode
	WHERE 
		h.VendorCode LIKE ?
		AND (d.CancelInd != 'Y') AND i.ItName LIKE ?
		AND h.DocNo LIKE ?
	HAVING 
		OutstandingQty > 0`

	args = append(args, searchVendor, searchItem, searchDoc)

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblpurchaseorder.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch purchase order: %w", err)
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

func (t *TblPurchaseOrderRepository) OutstandingPO(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) AS total FROM (
			SELECT 
			d.DocNo,
			d.DNo,
			h.DocDt,
			d.ItCode,
			i.ItName,
			d.Qty - COALESCE((
				SELECT SUM(r.PurchaseQty)
				FROM tblpurchasematerialreceivedtl r
				WHERE r.PurchaseOrderDocNo = d.DocNo
				AND r.PurchaseOrderDNo = d.DNo
				AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
			), 0) AS OutstandingQty,
			u.UomName
		FROM 
			tblpurchaseorderdtl d
		JOIN 
			tblpurchaseorderhdr h ON d.DocNo = h.DocNo
        JOIN tblpurchaseorderreqdtl por
        	ON d.PurchaseOrderReqDocNo = por.DocNo
            AND d.PurchaseOrderReqDNo = por.DNo
        JOIN tblmaterialrequestdtl mr
        	ON por.MaterialReqDocNo = mr.DocNo
            AND por.MaterialReqDNo = mr.DNo
        JOIN tblmaterialrequesthdr mrh
        	ON mr.DocNo = mrh.DocNo
        JOIN tblvendorquotationhdr vq
        	ON por.VendorQTDocNo = vq.DocNo
		JOIN 
			tblitem i ON d.ItCode = i.ItCode
		JOIN 
			tbluom u ON i.PurchaseUomCode = u.UomCode
		JOIN tblvendorhdr v
			ON h.VendorCode = v.VendorCode
		JOIN tblcurrency c
			ON vq.CurCode = c.CurCode
		WHERE 
			(d.CancelInd != 'Y') AND (i.ItName LIKE ?
			OR h.DocNo LIKE ? OR v.VendorName LIKE ?)
		`
	args = append(args, searchDoc, searchDoc, searchDoc)

	if startDate != "" && endDate != "" {
		countQuery += " AND DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}
	countQuery += ` HAVING 
			OutstandingQty > 0
	) AS grouped`

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

	var data []*tblpurchaseorder.OutstandingPO
	args = args[:0]

	query := `SELECT 
			d.DocNo,
			h.Status,
			v.VendorName,
			mrh.Department,
			i.ItName,
			d.Qty AS PurchaseQty,
			d.Qty - COALESCE((
				SELECT SUM(r.PurchaseQty)
				FROM tblpurchasematerialreceivedtl r
				WHERE r.PurchaseOrderDocNo = d.DocNo
				AND r.PurchaseOrderDNo = d.DNo
				AND (r.CancelInd IS NULL OR r.CancelInd != 'Y')
			), 0) AS OutstandingQty,
			u.UomName,
			c.CurName,
			(d.Total / d.Qty) AS Price,
			d.Total
		FROM 
			tblpurchaseorderdtl d
		JOIN 
			tblpurchaseorderhdr h ON d.DocNo = h.DocNo
        JOIN tblpurchaseorderreqdtl por
        	ON d.PurchaseOrderReqDocNo = por.DocNo
            AND d.PurchaseOrderReqDNo = por.DNo
        JOIN tblmaterialrequestdtl mr
        	ON por.MaterialReqDocNo = mr.DocNo
            AND por.MaterialReqDNo = mr.DNo
        JOIN tblmaterialrequesthdr mrh
        	ON mr.DocNo = mrh.DocNo
        JOIN tblvendorquotationhdr vq
        	ON por.VendorQTDocNo = vq.DocNo
		JOIN 
			tblitem i ON d.ItCode = i.ItCode
		JOIN 
			tbluom u ON i.PurchaseUomCode = u.UomCode
		JOIN tblvendorhdr v
			ON h.VendorCode = v.VendorCode
		JOIN tblcurrency c
			ON vq.CurCode = c.CurCode
		WHERE 
			(d.CancelInd != 'Y') AND (i.ItName LIKE ?
			OR h.DocNo LIKE ? OR v.VendorName LIKE ?)`
	args = append(args, searchDoc, searchDoc, searchDoc)

	if startDate != "" && endDate != "" {
		query += " AND i.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += ` HAVING 
			OutstandingQty > 0 LIMIT ? OFFSET ?`
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblpurchaseorder.OutstandingPO, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch purchase order: %w", err)
	}

	// proses data
	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
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
