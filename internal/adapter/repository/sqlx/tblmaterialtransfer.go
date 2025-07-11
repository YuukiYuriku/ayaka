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
	"gitlab.com/ayaka/internal/domain/tblmaterialtransfer"

	// "gitlab.com/ayaka/internal/domain/shared/formatid"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblMaterialTransferRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblMaterialTransferRepository) Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhsFrom := "%" + warehouseFrom + "%"
	searchWhsTo := "%" + warehouseTo + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblmaterialtransferhdr WHERE DocNo LIKE ? AND WhsCodeFrom LIKE ? AND WhsCodeTo LIKE ?"
	args = append(args, searchDoc, searchWhsFrom, searchWhsTo)

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

	var data []*tblmaterialtransfer.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.Status,
			i.WhsCodeFrom,
			w.WhsName AS WhsNameFrom,
			i.WhsCodeTo,
			w2.WhsName AS WhsNameTo,
			i.VendorCode,
			i.Driver,
			i.TransportType,
			i.LicenceNo,
			i.Note
			FROM tblmaterialtransferhdr i
			JOIN tblwarehouse w ON i.WhsCodeFrom = w.WhsCode
			JOIN tblwarehouse w2 ON i.WhsCodeTo = w2.WhsCode
			LEFT JOIN tblvendorhdr v ON i.VendorCode = v.VendorCode
			WHERE i.DocNo LIKE ? AND i.WhsCodeFrom LIKE ? AND i.WhsCodeTo LIKE ?`
	args = append(args, searchDoc, searchWhsFrom, searchWhsTo)

	if startDate != "" && endDate != "" {
		query += " AND i.DocDt BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &data, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblmaterialtransfer.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch material transfer: %w", err)
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
			Data:         make([]*tblmaterialtransfer.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblmaterialtransfer.Detail{}
	detailQuery := `SELECT 
				d.DocNo, 
				d.DNo,
				d.ItCode,
				i.ItName,
				d.BatchNo,
				d.Stock,
				d.Qty,
				u.UomName,
				d.CancelInd,
				d.SuccessInd
			FROM tblmaterialtransferdtl d
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
	detailMap := make(map[string][]tblmaterialtransfer.Detail)
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

func (t *TblMaterialTransferRepository) Create(ctx context.Context, data *tblmaterialtransfer.Create) (*tblmaterialtransfer.Create, error) {
	query := `INSERT INTO tblmaterialtransferhdr 
	(
		DocNo,
		DocDt,
		Status,
		WhsCodeFrom,
		WhsCodeTo,
		VendorCode,
		Driver,
		TransportType,
		LicenceNo,
		Note,
		CreateDt,
		CreateBy
	) VALUES `

	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.DocNo,
		data.Date,
		data.Status,
		data.WhsCodeFrom,
		data.WhsCodeTo,
		data.VendorCode,
		data.Driver,
		data.TransportType,
		data.LicenceNo,
		data.Note,
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
		query = `INSERT INTO tblmaterialtransferdtl (
			DocNo,
			DNo,
			CancelInd,
			SuccessInd,
			ItCode,
			BatchNo,
			Stock,
			Qty,
			Remark,
			CreateDt,
			CreateBy
		) VALUES `

		placeholders = placeholders[:0]
		args = args[:0]

		for _, detail := range data.Details {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				"N",
				"N",
				detail.ItCode,
				detail.BatchNo,
				detail.Stock,
				detail.Qty,
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

func (t *TblMaterialTransferRepository) Update(ctx context.Context, lastUpby, lastUpDate string, data *tblmaterialtransfer.Read) (*tblmaterialtransfer.Read, error) {
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
		args = append(args, data.DocNo, detail.DNo, detail.CancelInd)
	}

	query := `UPDATE tblmaterialtransferdtl
		SET CancelInd = CASE
			` + strings.Join(placeholders, " ") + `
			ELSE CancelInd
		END
		WHERE DocNo = ?	
	`
	args = append(args, data.DocNo)

	if resultDetail, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Failed to update material transfer dtl: %+v", err)
		return nil, fmt.Errorf("error updating material transfer dtl: %w", err)
	}

	if rowsAffectedDtl, err = resultDetail.RowsAffected(); err != nil {
		log.Printf("Failed to get rows affected for material transfer dtl: %+v", err)
		return nil, fmt.Errorf("error getting rows affected for material transfer dtl: %w", err)
	}

	if rowsAffectedDtl == 0 {
		err = customerrors.ErrNoDataEdited
		return data, err
	}

	// Update log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, lastUpby, data.DocNo, "MaterialTransfer", lastUpDate); err != nil {
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
