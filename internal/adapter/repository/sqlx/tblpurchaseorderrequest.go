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
	sharedfunc "gitlab.com/ayaka/internal/domain/shared/sharedFunc"
	"gitlab.com/ayaka/internal/domain/tblpurchaseorderrequest"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblPurchaseOrderRequestRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblPurchaseOrderRequestRepository) Create(ctx context.Context, data *tblpurchaseorderrequest.Create) (*tblpurchaseorderrequest.Create, error) {
	query := `INSERT INTO tblpurchaseorderreqhdr
	(
		DocNo,
		DocDt,
		Remark,
		CreateBy,
		CreateDt
	) VALUES `
	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		data.Remark,
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
		query = `INSERT INTO tblpurchaseorderreqdtl
		(
			DocNo,
			DNo,
			CancelInd,
			SuccessInd,
			MaterialReqDocNo,
			MaterialReqDNo,
			ItCode,
			Qty,
			Total,
			VendorCode,
			VendorQTDocNo,
			VendorQTDNo,
			CreateBy,
			CreateDt
		) VALUES `
		placeholders = placeholders[:0]
		args = args[:0]

		// material request
		var whensMatReq, whensOpenIndMatReq []string
		var wheresMatReq []string
		var argsMatReq, argsOpenIndMatReq, argsInMatReq []interface{}

		queryUpdateVendDtl := `UPDATE tblvendorquotationdtl
			SET UsedInd = CASE 
		`
		var whensVendDtl []string
		var wheresVendDtl []string
		var argsVendDtl, argsInVendDtl []interface{}
		vendorQuoteDocNos := make([]string, 0)

		for _, detail := range data.Details {
			detail.Total = detail.Qty * detail.ActualPrice
			// details
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				"N",
				"N",
				detail.MaterialReqDocNo,
				detail.MaterialReqDNo,
				detail.ItCode,
				detail.Qty,
				(detail.Total),
				detail.VendorCode,
				detail.VendorQTDocNo,
				detail.DNo,
				data.CreateBy,
				data.CreateDt,
			)

			// material req
			whensMatReq = append(whensMatReq, "WHEN DocNo = ? AND DNo = ? THEN 'Y'")
			argsMatReq = append(argsMatReq, detail.MaterialReqDocNo, detail.MaterialReqDNo)
			whensOpenIndMatReq = append(whensOpenIndMatReq, "WHEN DocNo = ? AND DNo = ? THEN 'N'")
			argsOpenIndMatReq = append(argsOpenIndMatReq, detail.MaterialReqDocNo, detail.MaterialReqDNo)
			// tuples material req
			wheresMatReq = append(wheresMatReq, "(?, ?)")
			argsInMatReq = append(argsInMatReq, detail.MaterialReqDocNo, detail.MaterialReqDNo)

			// vendor quotation
			whensVendDtl = append(whensVendDtl, "WHEN DocNo = ? AND DNo = ? THEN 'Y'")
			argsVendDtl = append(argsVendDtl, detail.VendorQTDocNo, detail.VendorQTDNo)
			// tuples vendor quotation
			wheresVendDtl = append(wheresVendDtl, "(?, ?)")
			argsInVendDtl = append(argsInVendDtl, detail.VendorQTDocNo, detail.VendorQTDNo)
			// save docno to update status
			vendorQuoteDocNos = append(vendorQuoteDocNos, detail.VendorQTDocNo)
		}

		// insert detail
		query += strings.Join(placeholders, ",") + ";"

		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Error insert detail: %+v", err)
			return nil, fmt.Errorf("error Insert Detail: %w", err)
		}

		// update material req
		queryUpdateMatReqDtl := `UPDATE tblmaterialrequestdtl
			SET 
				SuccessInd = CASE 
					` + strings.Join(whensMatReq, "\n") + `
					ELSE SuccessInd 
				END,
				OpenInd = CASE
					` + strings.Join(whensOpenIndMatReq, "\n") + `
					ELSE OpenInd 
				END
			WHERE (DocNo, DNo) IN (` + strings.Join(wheresMatReq, ", ") + `)
		`
		fullArgsMatReq := []interface{}{}
		fullArgsMatReq = append(fullArgsMatReq, argsMatReq...)        // untuk CASE SuccessInd
		fullArgsMatReq = append(fullArgsMatReq, argsOpenIndMatReq...) // untuk CASE OpenInd
		fullArgsMatReq = append(fullArgsMatReq, argsInMatReq...)      // untuk WHERE IN

		if _, err = tx.ExecContext(ctx, queryUpdateMatReqDtl, fullArgsMatReq...); err != nil {
			log.Printf("Error update material request detail batch: %+v", err)
			return nil, fmt.Errorf("error update material request detail: %w", err)
		}
 
		// update vendor quotation
		queryUpdateVendDtl += strings.Join(whensVendDtl, "\n") + "\nELSE UsedInd END\n"
		queryUpdateVendDtl += "WHERE (DocNo, Dno) IN (" + strings.Join(wheresVendDtl, ", ") + ")"
		argsVendDtl = append(argsVendDtl, argsInVendDtl...)

		log.Printf("Query update vendor dtl: %s", queryUpdateVendDtl)
		log.Printf("Args: %+v", argsVendDtl)
		if _, err = tx.ExecContext(ctx, queryUpdateVendDtl, argsVendDtl...); err != nil {
			log.Printf("Error update vendor quotation detail batch: %+v", err)
			return nil, fmt.Errorf("error update vendor quotation detail: %w", err)
		}

		// update header vendor quotation if all details already accept
		// Hilangkan duplikat DocNo vendor quotation
		vendorQuoteDocNos = sharedfunc.UniqueStringSlice(vendorQuoteDocNos)

		// Update status di tblvendorquotationhdr jika semua detail sudah sukses
		queryUpdateVendorHeader := `
			UPDATE tblvendorquotationhdr h
			SET Status = CASE 
				WHEN (
					SELECT COUNT(*) FROM tblvendorquotationdtl 
					WHERE DocNo = h.DocNo
				) = (
					SELECT COUNT(*) FROM tblvendorquotationdtl 
					WHERE DocNo = h.DocNo AND UsedInd = 'Y'
				) AND (
					SELECT COUNT(*) FROM tblvendorquotationdtl 
					WHERE DocNo = h.DocNo
				) > 0
				THEN 'Used'
				ELSE Status
			END
			WHERE h.DocNo IN (?)
		`

		// Gunakan sqlx.In dan Rebind
		queryUpdateVendorHeader, argsHeader, err := sqlx.In(queryUpdateVendorHeader, vendorQuoteDocNos)
		if err != nil {
			log.Printf("sqlx.In error vendor header: %+v", err)
			return nil, fmt.Errorf("failed to build IN clause for vendor header: %w", err)
		}
		queryUpdateVendorHeader = tx.Rebind(queryUpdateVendorHeader)

		if _, err = tx.ExecContext(ctx, queryUpdateVendorHeader, argsHeader...); err != nil {
			log.Printf("Error batch update vendorquotationhdr status: %+v", err)
			return nil, fmt.Errorf("error batch update vendor quotation header status: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblPurchaseOrderRequestRepository) Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblpurchaseorderreqhdr WHERE DocNo LIKE ?"
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

	var data []*tblpurchaseorderrequest.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.Remark
			FROM tblpurchaseorderreqhdr i
			WHERE i.DocNo LIKE ?`
	args = append(args, searchDoc)

	if startDate != "" && endDate != "" {
		query += " AND i.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblpurchaseorderrequest.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch purchase order request: %w", err)
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
			Data:         make([]*tblpurchaseorderrequest.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblpurchaseorderrequest.Detail{}
	detailQuery := `SELECT
				d.DocNo,
				d.DNo,
				d.CancelInd,
				d.SuccessInd,
				d.MaterialReqDocNo,
				d.MaterialReqDNo,
				d.ItCode,
				i.ItName,
				d.Qty,
				d.Total,
				m.UsageDt,
				d.VendorQTDocNo,
				d.VendorQTDNo,
				d.VendorCode,
				v.VendorName,
				c2.CurName AS ActualCurName,
				vq.Price,
				vqh.TermOfPayment,
				vqh.DeliveryType,
				d.Remark,
				m.EstimatedPrice,
				mrh.Department
		FROM tblpurchaseorderreqdtl d
		JOIN tblitem i ON d.ItCode = i.ItCode
		JOIN tblvendorhdr v ON d.VendorCode = v.VendorCode
		JOIN tblmaterialrequestdtl m
			ON d.MaterialReqDocNo = m.DocNo
			AND d.MaterialReqDNo = m.DNo
		JOIN tblmaterialrequesthdr mrh
			ON m.DocNo = mrh.DocNo
		JOIN tblvendorquotationdtl vq
			ON d.VendorQTDocNo = vq.DocNo
			AND d.VendorQTDNo = vq.DNo
		JOIN tblvendorquotationhdr vqh ON vqh.DocNo = vq.DocNo
		JOIN tblcurrency c2 ON vqh.CurCode = c2.CurCode
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
	detailMap := make(map[string][]tblpurchaseorderrequest.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range data {
		h.Details = detailMap[h.DocNo]
		for i := range h.Details {
			h.Details[i].UsageDt = share.FormatDate(h.Details[i].UsageDt)
		}
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

func (t *TblPurchaseOrderRequestRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblpurchaseorderrequest.Read) (*tblpurchaseorderrequest.Read, error) {
	var resultDetail sql.Result
	var rowsAffectedDtl int64

	var placeholders []string
	var args []interface{}

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

	// material request
	var whensMatReq, whensOpenIndMatReq []string
	var wheresMatReq []string
	var argsMatReq, argsOpenIndMatReq, argsInMatReq []interface{}

	queryUpdateVendDtl := `UPDATE tblvendorquotationdtl
			SET UsedInd = CASE 
		`
	var whensVendDtl []string
	var wheresVendDtl []string
	var argsVendDtl, argsInVendDtl []interface{}
	vendorQuoteDocNos := make([]string, 0)

	for _, detail := range data.Details {
		// detail cancel
		placeholders = append(placeholders, ` WHEN DocNo = ? AND DNo = ? THEN ? `)
		args = append(args, data.DocNo, detail.DNo, detail.CancelInd)

		// material req
		whensMatReq = append(whensMatReq, "WHEN DocNo = ? AND DNo = ? THEN ?")
		argsMatReq = append(argsMatReq, detail.MaterialReqDocNo, detail.MaterialReqDNo, booldatatype.FromBool(!detail.CancelInd.ToBool()))
		whensOpenIndMatReq = append(whensOpenIndMatReq, "WHEN DocNo = ? AND DNo = ? THEN 'Y'")
		argsOpenIndMatReq = append(argsOpenIndMatReq, detail.MaterialReqDocNo, detail.MaterialReqDNo)
		// tuples material req
		wheresMatReq = append(wheresMatReq, "(?, ?)")
		argsInMatReq = append(argsInMatReq, detail.MaterialReqDocNo, detail.MaterialReqDNo)

		// vendor quotation
		whensVendDtl = append(whensVendDtl, "WHEN DocNo = ? AND DNo = ? THEN 'N'")
		argsVendDtl = append(argsVendDtl, detail.VendorQTDocNo, detail.VendorQTDNo)
		// tuples vendor quotation
		wheresVendDtl = append(wheresVendDtl, "(?, ?)")
		argsInVendDtl = append(argsInVendDtl, detail.VendorQTDocNo, detail.VendorQTDNo)
		// save docno to update status
		vendorQuoteDocNos = append(vendorQuoteDocNos, detail.VendorQTDocNo)
	}
	query := `UPDATE tblpurchaseorderreqdtl
		SET CancelInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE CancelInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)
	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update purchase order request dtl: %+v", err)
		return nil, fmt.Errorf("error updating purchase order request dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for purchase order request dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for purchase order request dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// update material req
	queryUpdateMatReqDtl := `UPDATE tblmaterialrequestdtl
			SET 
				SuccessInd = CASE 
					` + strings.Join(whensMatReq, "\n") + `
					ELSE SuccessInd 
				END,
				OpenInd = CASE
					` + strings.Join(whensOpenIndMatReq, "\n") + `
					ELSE OpenInd 
				END
			WHERE (DocNo, DNo) IN (` + strings.Join(wheresMatReq, ", ") + `)
		`
	fullArgsMatReq := []interface{}{}
	fullArgsMatReq = append(fullArgsMatReq, argsMatReq...)        // untuk CASE SuccessInd
	fullArgsMatReq = append(fullArgsMatReq, argsOpenIndMatReq...) // untuk CASE OpenInd
	fullArgsMatReq = append(fullArgsMatReq, argsInMatReq...)      // untuk WHERE IN

	if _, err = tx.ExecContext(ctx, queryUpdateMatReqDtl, fullArgsMatReq...); err != nil {
		log.Printf("Error update material request detail batch: %+v", err)
		return nil, fmt.Errorf("error update material request detail: %w", err)
	}

	// update vendor quotation
	queryUpdateVendDtl += strings.Join(whensVendDtl, "\n") + "\nELSE UsedInd END\n"
	queryUpdateVendDtl += "WHERE (DocNo, Dno) IN (" + strings.Join(wheresVendDtl, ", ") + ")"
	argsVendDtl = append(argsVendDtl, argsInVendDtl...)

	log.Printf("Query update vendor dtl: %s", queryUpdateVendDtl)
	log.Printf("Args: %+v", argsVendDtl)
	if _, err = tx.ExecContext(ctx, queryUpdateVendDtl, argsVendDtl...); err != nil {
		log.Printf("Error update vendor quotation detail batch: %+v", err)
		return nil, fmt.Errorf("error update vendor quotation detail: %w", err)
	}

	// update header vendor quotation if all details already accept
	// Hilangkan duplikat DocNo vendor quotation
	vendorQuoteDocNos = sharedfunc.UniqueStringSlice(vendorQuoteDocNos)

	// Update status di tblvendorquotationhdr jika semua detail sudah sukses
	queryUpdateVendorHeader := `
			UPDATE tblvendorquotationhdr h
			SET Status = CASE 
				WHEN (
					SELECT COUNT(*) FROM tblvendorquotationdtl 
					WHERE DocNo = h.DocNo
				) = (
					SELECT COUNT(*) FROM tblvendorquotationdtl 
					WHERE DocNo = h.DocNo AND UsedInd = 'Y'
				) AND (
					SELECT COUNT(*) FROM tblvendorquotationdtl 
					WHERE DocNo = h.DocNo
				) > 0
				THEN 'Used'
				ELSE Status
			END
			WHERE h.DocNo IN (?)
		`

	// Gunakan sqlx.In dan Rebind
	queryUpdateVendorHeader, argsHeader, err := sqlx.In(queryUpdateVendorHeader, vendorQuoteDocNos)
	if err != nil {
		log.Printf("sqlx.In error vendor header: %+v", err)
		return nil, fmt.Errorf("failed to build IN clause for vendor header: %w", err)
	}
	queryUpdateVendorHeader = tx.Rebind(queryUpdateVendorHeader)

	if _, err = tx.ExecContext(ctx, queryUpdateVendorHeader, argsHeader...); err != nil {
		log.Printf("Error batch update vendorquotationhdr status: %+v", err)
		return nil, fmt.Errorf("error batch update vendor quotation header status: %w", err)
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "PurchaseOrderRequest", lastUpDate); err != nil {
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

func (t *TblPurchaseOrderRequestRepository) GetVendorQuotation(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchItem := "%" + itemName + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) AS total FROM (
				SELECT
					d.DocNo,
					d.DNo
				FROM tblpurchaseorderreqdtl d
				JOIN tblitem i ON d.ItCode = i.ItCode
				JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
				JOIN tblvendorquotationhdr vqh
					ON d.VendorQTDocNo = vqh.DocNo
				JOIN tblvendorquotationdtl vqd
					ON d.VendorQTDocNo = vqd.DocNo
					AND d.VendorQTDNo = vqd.DNo
				JOIN tblcurrency c ON vqh.CurCode = c.CurCode
				WHERE d.VendorCode LIKE ? AND i.ItName LIKE ? AND (d.CancelInd = 'N' AND d.SuccessInd = 'N')
			) AS grouped`
	args = append(args, searchDoc, searchItem)

	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, args...); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	var totalPages int
	var offset int

	fmt.Println("Total Records:", totalRecords)

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

	var data []*tblpurchaseorderrequest.GetPurchaseOrderRequest
	args = args[:0]

	query := `SELECT 
				d.DocNo,
				d.DNo,
				d.ItCode,
				i.ItName,
				d.Qty,
				u.UomName,
				c.CurName,
				vqd.Price,
				vqh.DeliveryType
			FROM tblpurchaseorderreqdtl d
			JOIN tblitem i ON d.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
			JOIN tblvendorquotationhdr vqh
				ON d.VendorQTDocNo = vqh.DocNo
			JOIN tblvendorquotationdtl vqd
				ON d.VendorQTDocNo = vqd.DocNo
				AND d.VendorQTDNo = vqd.DNo
			JOIN tblcurrency c ON vqh.CurCode = c.CurCode
			WHERE d.VendorCode LIKE ? AND i.ItName LIKE ? AND (d.CancelInd = 'N' AND d.SuccessInd = 'N')`
	args = append(args, searchDoc, searchItem)

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblpurchaseorderrequest.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch purchase order request: %w", err)
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
