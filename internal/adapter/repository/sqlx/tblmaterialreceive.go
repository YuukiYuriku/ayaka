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
	"gitlab.com/ayaka/internal/domain/tblmaterialreceive"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblMaterialReceiveRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblMaterialReceiveRepository) Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchDoc := "%" + doc + "%"
	searchWhsFrom := "%" + warehouseFrom + "%"
	searchWhsTo := "%" + warehouseTo + "%"
	var args []interface{}

	countQuery := "SELECT COUNT(*) FROM tblmaterialreceivehdr WHERE DocNo LIKE ? AND WhsCodeFrom LIKE ? AND WhsCodeTo LIKE ?"
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

	var data []*tblmaterialreceive.Read
	args = args[:0]

	query := `SELECT 
			i.DocNo,
			i.DocDt,
			i.WhsCodeFrom,
			w.WhsName AS WhsNameFrom,
			i.WhsCodeTo,
			w2.WhsName AS WhsNameTo,
			i.Remark
			FROM tblmaterialreceivehdr i
			JOIN tblwarehouse w ON i.WhsCodeFrom = w.WhsCode
			JOIN tblwarehouse w2 ON i.WhsCodeTo = w2.WhsCode
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
				Data:         make([]*tblmaterialreceive.Read, 0),
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
			Data:         make([]*tblmaterialreceive.Read, 0),
			TotalRecords: totalRecords,
			TotalPages:   totalPages,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      param.Page < totalPages,
			HasPrevious:  param.Page > 1,
		}, nil
	}

	details := []*tblmaterialreceive.Detail{}
	detailQuery := `SELECT 
				d.DocNo, 
				d.DNo,
				d.DocNoMaterialTransfer,
				d.ItCode,
				i.ItName,
				d.BatchNo,
				d.Source,
				d.QtyTransfer,
				d.QtyActual,
				u.UomName,
				d.Remark
			FROM tblmaterialreceivedtl d
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
	detailMap := make(map[string][]tblmaterialreceive.Detail)
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

func (t *TblMaterialReceiveRepository) Create(ctx context.Context, data *tblmaterialreceive.Create) (*tblmaterialreceive.Create, error) {
	query := `INSERT INTO tblmaterialreceivehdr 
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
		query = `INSERT INTO tblmaterialreceivedtl (
			DocNo,
			DNo,
			DocNoMaterialTransfer,
			ItCode,
			BatchNo,
			Source,
			QtyTransfer,
			QtyActual,
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

		// transfer between warehouse
		queryTransfer := `INSERT INTO tbltransferbetweenwhs (
			DocNo,
			DocDt,
			WhsFrom,
			WhsTo,
			ItCode,
			BatchNo,
			Qty,
			Remark,
			CreateBy,
			CreateDt
		) VALUES `
		var placeholdersTransfer []string
		var argsTransfer []interface{}

		// update material transfer
		queryUpdate := `UPDATE tblmaterialtransferdtl
			SET SuccessInd = CASE
		`
		var whensUpdate []string
		var wheresUpdate []string
		var argsUpdate []interface{}

		for _, detail := range data.Details {
			// detail
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.DocNo,
				detail.DNo,
				detail.DocNoMaterialTransfer,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				detail.QtyTransfer,
				detail.QtyActual,
				detail.Remark,
				data.CreateDt,
				data.CreateBy,
			)

			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			////////////////// WAREHOUSE FROM
			argsStockSummary = append(argsStockSummary,
				data.WhsCodeFrom,
				"-",
				"-",
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				0,
				detail.QtyActual,
				data.CreateBy,
				data.Date,
			)
			argsStockMovement = append(argsStockMovement,
				"Material Transfer",
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
				detail.QtyActual,
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

			////////////////// WAREHOUSE TO
			// stock summary
			placeholdersStockSummary = append(placeholdersStockSummary, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			// stock movement
			placeholdersStockMovement = append(placeholdersStockMovement, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

			argsStockSummary = append(argsStockSummary,
				data.WhsCodeTo,
				"-",
				"-",
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.QtyActual,
				0,
				data.CreateBy,
				data.Date,
			)
			argsStockMovement = append(argsStockMovement,
				"Material Receive",
				data.DocNo,
				detail.DNo,
				"N",
				data.Date,
				data.WhsCodeTo,
				detail.Source,
				detail.ItCode,
				detail.BatchNo,
				0,
				detail.QtyActual,
				0,
				data.Remark,
				data.CreateBy,
				data.Date,
			)

			placeholdersTransfer = append(placeholdersTransfer, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			argsTransfer = append(argsTransfer,
				data.DocNo,
				data.Date,
				data.WhsCodeFrom,
				data.WhsCodeTo,
				detail.ItCode,
				detail.BatchNo,
				detail.QtyActual,
				detail.Remark,
				data.CreateBy,
				data.CreateDt,
			)

			whensUpdate = append(whensUpdate, `
				WHEN DocNo = ? AND ItCode = ? AND BatchNo = ? AND Qty <= (
					SELECT COALESCE(SUM(QtyActual), 0)
					FROM tblmaterialreceivedtl
					WHERE DocNoMaterialTransfer = ? AND ItCode = ? AND BatchNo = ?
				) THEN 'Y'
			`)
			argsUpdate = append(argsUpdate,
				detail.DocNoMaterialTransfer,
				detail.ItCode,
				detail.BatchNo,
				detail.DocNoMaterialTransfer,
				detail.ItCode,
				detail.BatchNo,
			)
			wheresUpdate = append(wheresUpdate, "(DocNo = ? AND ItCode = ? AND BatchNo = ?)")
			argsUpdate = append(argsUpdate, detail.DocNoMaterialTransfer, detail.ItCode, detail.BatchNo)
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

		// insert report
		queryTransfer += strings.Join(placeholdersTransfer, ",") + ";"
		if _, err = tx.ExecContext(ctx, queryTransfer, argsTransfer...); err != nil {
			log.Printf("Error insert report transfer: %+v", err)
			return nil, fmt.Errorf("error Insert report transfer: %w", err)
		}

		// update detail
		queryUpdate += strings.Join(whensUpdate, " ")
		queryUpdate += `
			ELSE SuccessInd
			END
		WHERE ` + strings.Join(wheresUpdate, " OR ")
		fmt.Println("query: ", queryUpdate)
		fmt.Println("args: ", argsUpdate)
		if _, err = tx.ExecContext(ctx, queryUpdate, argsUpdate...); err != nil {
			log.Printf("Error Update Material Transfer: %+v", err)
			return nil, fmt.Errorf("error Update Material Transfer: %w", err)
		}

		////////////////////////////////////////// CHECK HEADER
		queryUpdateHeader := `
			UPDATE tblmaterialtransferhdr h
			SET Status = CASE 
				-- Jika semua detail sudah success
				WHEN (
					SELECT COUNT(*) FROM tblmaterialtransferdtl 
					WHERE DocNo = h.DocNo AND (CancelInd IS NULL OR CancelInd != 'Y')
				) = (
					SELECT COUNT(*) FROM tblmaterialtransferdtl 
					WHERE DocNo = h.DocNo AND SuccessInd = 'Y' AND (CancelInd IS NULL OR CancelInd != 'Y')
				) AND (
					SELECT COUNT(*) FROM tblmaterialtransferdtl 
					WHERE DocNo = h.DocNo AND (CancelInd IS NULL OR CancelInd != 'Y')
				) > 0
				THEN 'Success'

				-- Jika ada yang sudah mulai diterima (QtyActual > 0)
				WHEN EXISTS (
					SELECT 1 FROM tblmaterialreceivedtl r
					WHERE r.DocNoMaterialTransfer = h.DocNo AND r.QtyActual > 0
				)
				THEN 'Partial'

				-- Tetap status sebelumnya
				ELSE Status
			END
			WHERE h.DocNo IN (?)
		`

		docNos := make([]string, 0)
		for _, d := range data.Details {
			if d.DocNoMaterialTransfer != "" {
				docNos = append(docNos, d.DocNoMaterialTransfer)
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
			log.Printf("Error batch update materialtransferhdr status: %+v", err)
			return nil, fmt.Errorf("error batch update header status: %w", err)
		}

	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}
