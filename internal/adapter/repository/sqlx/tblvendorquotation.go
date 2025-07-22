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
	"gitlab.com/ayaka/internal/domain/tblmasteritem"
	"gitlab.com/ayaka/internal/domain/tblvendorquotation"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorQuotationRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblVendorQuotationRepository) Fetch(ctx context.Context, doc, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchVendor := "%" + vendor + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) FROM tblvendorquotationhdr h 
		JOIN tblvendorhdr v ON h.VendorCode = v.VendorCode
		WHERE h.DocNo LIKE ? AND v.VendorName LIKE ?`
	args = append(args, searchDoc, searchVendor)

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

	var data []*tblvendorquotation.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.Status,
			w.VendorName,
			w.VendorCode,
			c.CurCode,
			i.DeliveryType,
			i.TermOfPayment,
			i.Remark
			FROM tblvendorquotationhdr i
			JOIN tblvendorhdr w ON i.VendorCode = w.VendorCode
			JOIN tblcurrency c ON i.CurCode = c.CurCode
			WHERE i.DocNo LIKE ? AND w.VendorName LIKE ?`
	args = append(args, searchDoc, searchVendor)

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
		return nil, fmt.Errorf("error Fetch vendor quotation: %w", err)
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
			Data:         make([]*tblvendorquotation.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblvendorquotation.Detail{}
	detailQuery := `SELECT 
				d.DocNo, d.DNo, d.ActiveInd, d.UsedInd, d.ItCode, d.Price,
				i.ItName
			FROM tblvendorquotationdtl d
			JOIN tblitem i ON d.ItCode = i.ItCode
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
	detailMap := make(map[string][]tblvendorquotation.Detail)
	for _, d := range details {
		detailMap[d.DocNo] = append(detailMap[d.DocNo], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range data {
		h.Details = detailMap[h.DocNo]
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

func (t *TblVendorQuotationRepository) Create(ctx context.Context, data *tblvendorquotation.Create) (*tblvendorquotation.Create, error) {
	query := `INSERT INTO tblvendorquotationhdr 
	(
		DocNo,
		DocDt,
		Status,
		VendorCode,
		TermOfPayment,
		CurCode,
		DeliveryType,
		Remark,
		CreateDt,
		CreateBy
	) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		"Outstanding",
		data.VendorCode,
		data.TermOfPayment,
		data.CurCode,
		data.DeliveryType,
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
		query = `INSERT INTO tblvendorquotationdtl (
			DocNo,
			DNo,
			ActiveInd,
			UsedInd,
			ItCode,
			Price,
			Remark,
			CreateDt,
			CreateBy
		) VALUES `

		placeholders = placeholders[:0]
		args = args[:0]

		for i, detail := range data.Details {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				fmt.Sprintf("%03d", i+1),
				"Y",
				"N",
				detail.ItCode,
				detail.Price,
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

func (t *TblVendorQuotationRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblvendorquotation.Read) (*tblvendorquotation.Read, error) {
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
		args = append(args, data.DocNo, detail.DNo, detail.ActiveInd)
	}

	query := `UPDATE tblvendorquotationdtl
		SET ActiveInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE ActiveInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update vendor quotation dtl: %+v", err)
		return nil, fmt.Errorf("error updating vendor quotation dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for vendor quotation dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for vendor quotation dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "VendorQuotation", lastUpDate); err != nil {
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

func (t *TblVendorQuotationRepository) GetVendorQuotation(ctx context.Context, itemCode string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchItem := "%" + itemCode + "%"
	var args []interface{}

	countQuery := `SELECT COUNT(*) FROM tblvendorquotationhdr h
		JOIN tblvendorquotationdtl d ON h.DocNo = d.DocNo
		WHERE d.ItCode LIKE ? AND d.ActiveInd = 'Y'`
	args = append(args, searchItem)

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

	var data []*tblvendorquotation.GetVendorQuotation

	query := `SELECT
			d.DocNo,
			d.DNo,
			h.VendorCode,
			v.VendorName,
			h.TermOfPayment,
			h.DeliveryType,
			c.CurName,
			d.Price
		FROM tblvendorquotationdtl d
		JOIN tblvendorquotationhdr h ON d.DocNo = h.DocNo
		JOIN tblvendorhdr v ON h.VendorCode = v.VendorCode
		JOIN tblcurrency c ON h.CurCode = c.CurCode
		WHERE d.ItCode LIKE ? AND d.ActiveInd = 'Y' AND d.UsedInd = 'N'
	`
	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblvendorquotation.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch vendor quotation: %w", err)
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