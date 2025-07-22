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
	"gitlab.com/ayaka/internal/domain/tblmaterialrequest"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblMaterialRequestRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblMaterialRequestRepository) Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblmaterialrequesthdr WHERE DocNo LIKE ?"
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

	var data []*tblmaterialrequest.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.SiteCode,
			i.Department,
			i.Remark
			FROM tblmaterialrequesthdr i
			LEFT JOIN tblsite s ON i.SiteCode = s.SiteCode
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
				Data:         make([]*tblmaterialrequest.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch material request: %w", err)
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
			Data:         make([]*tblmaterialrequest.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblmaterialrequest.Detail{}
	detailQuery := `SELECT 
				d.DocNo, 
				d.DNo,
				d.CancelInd,
				d.SuccessInd,
				d.ItCode,
				i.ItName,
				d.Qty,
				u.UomName,
				d.UsageDt,
				d.CurCode,
				d.EstimatedPrice,
				d.Remark,
				d.Duration,
				d.DurationUom
			FROM tblmaterialrequestdtl d
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
	detailMap := make(map[string][]tblmaterialrequest.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range data {
		h.Details = detailMap[h.DocNo]
		for i := range h.Details {
			h.Details[i].UsageDt = share.ToDatePicker(h.Details[i].UsageDt)
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

func (t *TblMaterialRequestRepository) Create(ctx context.Context, data *tblmaterialrequest.Create) (*tblmaterialrequest.Create, error) {
	query := `INSERT INTO tblmaterialrequesthdr 
	(
		DocNo,
		DocDt,
		SiteCode,
		Department,
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
		data.SiteCode,
		data.Department,
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
		query = `INSERT INTO tblmaterialrequestdtl (
			DocNo,
			DNo,
			CancelInd,
			SuccessInd,
			ItCode,
			Qty,
			UsageDt,
			CurCode,
			EstimatedPrice,
			Duration,
			DurationUom,
			Remark,
			CreateDt,
			CreateBy
		) VALUES `

		placeholders = placeholders[:0]
		args = args[:0]

		for _, detail := range data.Details {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				"N",
				"N",
				detail.ItCode,
				detail.Qty,
				detail.UsageDt,
				detail.CurCode,
				detail.EstimatedPrice,
				detail.Duration,
				detail.DurationUom,
				detail.Remark,
				data.CreateDt,
				data.CreateBy,
			)
		}

		// insert detail
		query += strings.Join(placeholders, ",") + ";"
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Error insert detail: %+v", err)
			return nil, fmt.Errorf("error Insert Detail: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblMaterialRequestRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblmaterialrequest.Read) (*tblmaterialrequest.Read, error) {
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

	for _, detail := range data.Details {
		// detail cancel
		placeholders = append(placeholders, ` WHEN DocNo = ? AND DNo = ? THEN ? `)
		args = append(args, data.DocNo, detail.DNo, detail.Cancel)
	}

	query := `UPDATE tblmaterialrequestdtl
		SET CancelInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE CancelInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update material request dtl: %+v", err)
		return nil, fmt.Errorf("error updating material request dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for material request dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for material request dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "MaterialRequest", lastUpDate); err != nil {
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

func (t *TblMaterialRequestRepository) GetMaterialRequest(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchItem := "%" + itemName + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) AS total FROM (
			SELECT
				mr.DocNo,
				mr.DNo
			FROM tblmaterialrequestdtl mr
			JOIN tblitem i ON mr.ItCode = i.ItCode
			JOIN tblmaterialrequesthdr h ON mr.DocNo = h.DocNo
			LEFT JOIN tblpurchaseorderreqdtl por
			ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			LEFT JOIN tblpurchaseorderdtl po
			ON por.DocNo = po.PurchaseOrderReqDocNo AND por.DNo = po.PurchaseOrderReqDNo
			LEFT JOIN tblpurchasematerialreceivedtl r
			ON po.DocNo = r.PurchaseOrderDocNo AND po.DNo = r.PurchaseOrderDNo
			WHERE mr.CancelInd = 'N' AND mr.OpenInd = 'Y'
			AND (h.DocNo LIKE ? OR i.ItName LIKE ?)
			GROUP BY mr.DocNo, mr.DNo, h.Department, mr.ItCode, i.ItName, mr.Qty, mr.EstimatedPrice, mr.UsageDt
			HAVING (mr.Qty - COALESCE(SUM(r.PurchaseQty), 0)) > 0
		) AS count_alias;`
	args = append(args, searchDoc, searchItem)

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

	var data []*tblmaterialrequest.GetMaterialRequest

	query := `SELECT
				mr.DocNo AS MaterialReqDocNo,
				mr.DNo AS MaterialReqDNo,
				mr.ItCode,
				i.ItName,
				h.Department,
				mr.EstimatedPrice,
				mr.UsageDt,
				(mr.Qty - COALESCE(SUM(r.PurchaseQty), 0)) AS OutstandingQty
			FROM tblmaterialrequestdtl mr
			JOIN tblitem i ON mr.ItCode = i.ItCode
			JOIN tblmaterialrequesthdr h ON mr.DocNo = h.DocNo
			LEFT JOIN tblpurchaseorderreqdtl por
			ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			LEFT JOIN tblpurchaseorderdtl po
			ON por.DocNo = po.PurchaseOrderReqDocNo AND por.DNo = po.PurchaseOrderReqDNo
			LEFT JOIN tblpurchasematerialreceivedtl r
			ON po.DocNo = r.PurchaseOrderDocNo AND po.DNo = r.PurchaseOrderDNo
			WHERE mr.CancelInd != 'Y' AND mr.OpenInd = 'Y' AND (mr.DocNo LIKE ? OR i.ItName LIKE ?)
			GROUP BY mr.DocNo, mr.DNo, mr.ItCode, mr.Qty
			HAVING OutstandingQty > 0
	`
	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblmaterialrequest.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch material request: %w", err)
	}

	// proses data
	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
		detail.UsageDt = share.FormatDate(detail.UsageDt)
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

func (t *TblMaterialRequestRepository) OutstandingMaterial(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) FROM (
			SELECT 
				mr.DocNo,
				i.ItName,
				(mr.Qty - COALESCE(SUM(r.PurchaseQty), 0)) AS OutstandingQty,
				mr.Qty RequestedQty
			FROM tblmaterialrequestdtl mr
			JOIN tblitem i ON mr.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
			JOIN tblmaterialrequesthdr h ON mr.DocNo = h.DocNo
			LEFT JOIN tblpurchaseorderreqdtl por
				ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			LEFT JOIN tblpurchaseorderdtl po
				ON por.DocNo = po.PurchaseOrderReqDocNo AND por.DNo = po.PurchaseOrderReqDNo
			LEFT JOIN tblpurchasematerialreceivedtl r
				ON po.DocNo = r.PurchaseOrderDocNo AND po.DNo = r.PurchaseOrderDNo
			WHERE mr.CancelInd != 'Y' AND mr.OpenInd = 'Y' AND mr.DocNo LIKE ?
			`
	args = append(args, searchDoc)

	if startDate != "" && endDate != "" {
		countQuery += " AND DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	countQuery += `GROUP BY mr.DocNo, mr.DNo, mr.ItCode, mr.Qty	HAVING OutstandingQty > 0
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

	var data []*tblmaterialrequest.OutstandingMaterial
	args = args[:0]

	query := `SELECT 
				mr.DocNo,
				h.Department,
				i.ItName,
				mr.Qty RequestedQty,
				COALESCE(SUM(r.PurchaseQty), 0) AS ReceivedQty,
				(mr.Qty - COALESCE(SUM(r.PurchaseQty), 0)) AS OutstandingQty,
				u.UomName,
				mr.UsageDt
			FROM tblmaterialrequestdtl mr
			JOIN tblitem i ON mr.ItCode = i.ItCode
			JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
			JOIN tblmaterialrequesthdr h ON mr.DocNo = h.DocNo
			LEFT JOIN tblpurchaseorderreqdtl por
				ON mr.DocNo = por.MaterialReqDocNo AND mr.DNo = por.MaterialReqDNo
			LEFT JOIN tblpurchaseorderdtl po
				ON por.DocNo = po.PurchaseOrderReqDocNo AND por.DNo = po.PurchaseOrderReqDNo
			LEFT JOIN tblpurchasematerialreceivedtl r
				ON po.DocNo = r.PurchaseOrderDocNo AND po.DNo = r.PurchaseOrderDNo
			WHERE mr.CancelInd != 'Y' AND mr.OpenInd = 'Y' AND mr.DocNo LIKE ?
	`
	args = append(args, searchDoc)

	if startDate != "" && endDate != "" {
		query += " AND i.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " GROUP BY mr.DocNo, mr.DNo, mr.ItCode, mr.Qty HAVING OutstandingQty > 0 LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblmaterialrequest.OutstandingMaterial, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch material request: %w", err)
	}

	// proses data
	j := offset
	for _, detail := range data {
		j++
		detail.Number = uint(j)
		detail.UsageDate = share.FormatDate(detail.UsageDate)
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
