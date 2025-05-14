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
		detail.Date = share.FormatDate(detail.Date)
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
			d.ItCode,
			i.ItName,
			d.BatchNo,
			(
				SELECT 
					Qty - Qty2 - Qty3
				FROM tblstocksummary ss
				WHERE ss.WhsCode = h.WhsCode
				AND ss.ItCode = i.ItCode
			) AS Stock,
			d.Qty,
			u.UomName
		FROM tblstockmutationdtl d
		JOIN tblstockmutationhdr h ON h.DocNo = d.DocNo
		JOIN tblitem i ON i.ItCode = d.ItCode
		JOIN tbluom u ON u.UomCode = i.PurchaseUOMCode
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

	return &data, nil
}

func (t *TblStockMutationRepository) Create(ctx context.Context, data *tblstockmutationhdr.Create) (*tblstockmutationhdr.Create, error) {
	// Mulai transaksi
	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	var txErr error
	defer func() {
		if txErr != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %+v", rbErr)
			}
		}
	}()

	var args []interface{}

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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
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

	// insert header
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		txErr = fmt.Errorf("failed to insert header stock mutation: %w", err)
		return nil, txErr
	}

	if len(data.FromArray) > 0 && len(data.ToArray) > 0 {
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
		queryInsertSum := `
			INSERT INTO tblstocksummary (
				WhsCode,
				Lot,
				Bin,
				Source,
				ItCode,
				BatchNo,
				Qty,
				Qty2,
				Qty3,
				Remark,
				CreateBy,
				CreateDt
			) VALUES 
		`

		queryMov := `INSERT INTO tblstockmovement (
			DocType,
			DocNo,
			DNo,
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

		var placeholders []string
		var whenClauses []string
		var movementValues []string
		var summaryValues []string

		args = args[:0]
		var argsClauses, argsMov, argsSummary []interface{}

		// set insert and value for from array
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

			whenClauses = append(whenClauses, ` WHEN WhsCode = ? AND ItCode = ? THEN ?`)
			argsClauses = append(argsClauses, data.WarehouseCode, detail.ItCode, detail.Qty)

			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			argsMov = append(argsMov,
				"Stock Mutation",
				data.DocNo,
				detail.DNo,
				data.DocDate,
				data.WarehouseCode,
				detail.Source,
				detail.ItCode,
				data.BatchNo,
				detail.Qty,
				detail.Qty,
				detail.Qty,
				data.Remark,
				data.CreateBy,
				data.CreateDate,
			)
		}

		// set insert and value for to array
		for _, detail := range data.ToArray {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				detail.DocNo,
				detail.DNo,
				detail.ItCode,
				detail.BatchNo,
				detail.Source,
				detail.Qty,
				"To",
				data.CreateDate,
				data.CreateBy,
			)

			summaryValues = append(summaryValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			argsSummary = append(argsSummary,
				data.WarehouseCode,
				"-",
				"-",
				detail.ItCode,
				detail.ItCode,
				detail.BatchNo,
				detail.Qty,
				0,
				0,
				data.Remark,
				data.CreateBy,
				data.CreateDate,
			)

			movementValues = append(movementValues, `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			argsMov = append(argsMov,
				"Stock Mutation",
				data.DocNo,
				detail.DNo,
				data.DocDate,
				data.WarehouseCode,
				detail.Source,
				detail.ItCode,
				data.BatchNo,
				detail.Qty,
				detail.Qty,
				detail.Qty,
				data.Remark,
				data.CreateBy,
				data.CreateDate,
			)
		}

		// insert detail
		query += strings.Join(placeholders, ",")
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			txErr = fmt.Errorf("failed to insert detail stock mutation: %w", err)
			return nil, txErr
		}

		// insert stock sum
		queryInsertSum += strings.Join(summaryValues, ",") + `
			ON DUPLICATE KEY UPDATE
				Qty = Qty + VALUES(Qty),
				Qty2 = Qty2,
				Qty3 = Qty3,
				LastUpBy = VALUES(CreateBy),
				LastUpDt = VALUES(CreateDt);`
		_, err = tx.ExecContext(ctx, queryInsertSum, argsSummary...)
		if err != nil {
			txErr = fmt.Errorf("failed to insert stock summary: %w", err)
			return nil, txErr
		}

		// insert movement
		queryMov += strings.Join(movementValues, ",")
		_, err = tx.ExecContext(ctx, queryMov, argsMov...)
		if err != nil {
			txErr = fmt.Errorf("failed to insert stock movement: %w", err)
			return nil, txErr
		}

		// Update ke tabel stock summary
		querySum := `UPDATE tblstocksummary 
				SET
					Qty3 = CASE ` + strings.Join(whenClauses, " ") + `
					ELSE Qty3
					END,
					LastUpBy = ?,
					LastUpDt = ?
				WHERE WhsCode = ?;`
		argsClauses = append(argsClauses, data.CreateBy, data.CreateDate, data.WarehouseCode)

		_, err = tx.ExecContext(ctx, querySum, argsClauses...)
		if err != nil {
			txErr = fmt.Errorf("failed to update stock summary: %w", err)
			return nil, txErr
		}
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}
