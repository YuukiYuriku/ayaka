package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblvendorsector"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorSectorRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblVendorSectorRepository) Fetch(ctx context.Context, name, active string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	fmt.Println("active: ", active)
	var totalRecords int
	var args []interface{}

	search := "%" + name + "%"
	countQuery := "SELECT COUNT(*) FROM tblvendorsectorhdr WHERE (SectorCode LIKE ? OR SectorName LIKE ?)"
	args = append(args, search, search)

	if active != "" {
		countQuery += " AND Active = ?"
		args = append(args, active)
	}

	fmt.Println("count query: ", countQuery)
	fmt.Println("args: ", args)

	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, args...); err != nil {
		return nil, fmt.Errorf("error counting records: %w", err)
	}
	fmt.Println("total records: ", totalRecords)

	var totalPages int
	var offset int

	if param != nil {
		totalPages, offset = pagination.CountPagination(param, totalRecords)
		fmt.Println("total pages: ", totalPages)
		fmt.Println("offset", offset)

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

	var items []*tblvendorsector.Read
	query := `SELECT SectorCode,
				SectorName,
				Active,
				CreateDt
				FROM tblvendorsectorhdr
				WHERE (SectorCode LIKE ? OR SectorName LIKE ?)
				`

	if active != "" {
		query += " AND Active = ?"
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	fmt.Println("query: ", query)
	fmt.Println("args: ", args)

	if err := t.DB.SelectContext(ctx, &items, query, args...); err != nil {
		fmt.Println("error: ", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblvendorsector.Read, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Vendor Sector: %w", err)
	}

	// proses data
	j := offset
	docsNo := make([]string, len(items))
	for i, item := range items {
		j++
		item.Number = uint(j)
		item.CreateDate = share.FormatDate(item.CreateDate)
		docsNo[i] = item.SectorCode
	}

	if len(docsNo) == 0 {
		return &pagination.PaginationResponse{
			Data:         make([]*tblvendorsector.Read, 0),
			TotalRecords: 0,
			TotalPages:   0,
			CurrentPage:  param.Page,
			PageSize:     param.PageSize,
			HasNext:      false,
			HasPrevious:  false,
		}, nil
	}

	details := []*tblvendorsector.VendorSectorDtl{}
	detailQuery := `SELECT
						d.SectorCode, d.DNo, d.SubSectorName
					FROM tblvendorsectordtl d
					WHERE d.SectorCode IN (?) AND d.Active = 'Y';
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
	detailMap := make(map[string][]tblvendorsector.VendorSectorDtl)
	for _, d := range details {
		detailMap[d.SectorCode] = append(detailMap[d.SectorCode], *d)
	}

	// Gabungkan header dengan detail
	for _, h := range items {
		h.Details = detailMap[h.SectorCode]
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         items,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}
	fmt.Println("seccess")

	return response, nil
}

func (t *TblVendorSectorRepository) Create(ctx context.Context, data *tblvendorsector.Create) (*tblvendorsector.Create, error) {
	countDetail := len(data.Details)

	// Mulai transaksi
	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	// Pastikan rollback selalu dijalankan jika terjadi error
	defer func() {
		if err != nil {
			fmt.Println("error: ", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %+v", rbErr)
			}
		}
	}()

	// insert header
	query := `INSERT INTO tblvendorsectorhdr
				(
					SectorCode,
					SectorName,
					Active,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?, ?);`

	_, err = tx.ExecContext(ctx, query,
		data.SectorCode,
		data.SectorName,
		"Y",
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Vendor Sector Header: %w", err)
	}
	fmt.Println("success add header")

	if countDetail > 0 {
		query = `INSERT INTO tblvendorsectordtl (
			SectorCode,
			DNo,
			SubSectorName,
			CreateBy,
			CreateDt
		) VALUES `
		var placeholders []string
		var args []interface{}

		for i, detail := range data.Details {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			args = append(args,
				data.SectorCode,
				fmt.Sprintf("%05d", (i+1)),
				detail.SubSectorName,
				data.CreateBy,
				data.CreateDate,
			)
		}

		query += strings.Join(placeholders, ",")

		fmt.Println("query: ", query)
		fmt.Println("args: ", args)

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			fmt.Println("error inser detail: ", err)
			return nil, fmt.Errorf("failed to insert details: %w", err)
		}
		fmt.Println("success add")
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	fmt.Println("data: ", data)
	return data, nil
}

func (t *TblVendorSectorRepository) Update(ctx context.Context, data *tblvendorsector.Update) (*tblvendorsector.Update, error) {
	count := len(data.Details)

	if count < 1 {
		return nil, customerrors.ErrNoDataEdited
	}
	var args []interface{}
	var whenClause []string

	query := `UPDATE tblvendorsectorhdr SET
			SectorName = ?,
			Active = ?
			WHERE SectorCode = ?`

	// start transaction
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

	var resultHdr, resultDtl sql.Result
	var rowsAffectedHdr, rowsAffectedDtl int64
	// HEADER
	resultHdr, err = tx.ExecContext(ctx, query,
		data.SectorName,
		data.Active,
		data.SectorCode,
	)

	if err != nil {
		return nil, fmt.Errorf("error updating Vendor Sector: %w", err)
	}

	rowsAffectedHdr, err = resultHdr.RowsAffected()
	if err != nil {
		log.Printf("Failed to check row affected: %+v", err)
		return nil, fmt.Errorf("error check row affected header: %w", err)
	}

	// DETAIL
	if len(data.Details) > 0 {
		query = `SELECT SectorCode, DNo, SubSectorName, Active FROM tblvendorsectordtl WHERE SectorCode = ? AND Active = 'Y'`
		var exists []tblvendorsector.VendorSectorDtl

		if err := tx.SelectContext(ctx, &exists, query, data.SectorCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Detailed error: %+v", err)
				return nil, fmt.Errorf("error fetching detail: %w", err)
			}
		}

		var deleteds []tblvendorsector.VendorSectorDtl
		for _, exist := range exists {
			found := false
			for _, detailData := range data.Details {
				if exist.DNo == detailData.DNo {
					found = true
					break
				}
			}
			if !found {
				exist.Active = booldatatype.FromBool(false)
				deleteds = append(deleteds, exist)
			}
		}

		if len(deleteds) > 0 {
			data.Details = append(data.Details, deleteds...)
		}

		getDNo := `SELECT DNo FROM tblvendorsectordtl WHERE SectorCode = ? ORDER BY CreateDt DESC LIMIT 1`
		var lastDNo string

		if err := tx.GetContext(ctx, &lastDNo, getDNo, data.SectorCode); err != nil {
			log.Printf("Failed to get last DNo: %+v", err)
			return nil, fmt.Errorf("error getting last DNo: %w", err)
		}

		DNo, err := strconv.Atoi(lastDNo)
		if err != nil {
			log.Printf("Failed to convert DNo to int: %+v", err)
			return nil, fmt.Errorf("error converting DNo to int: %w", err)
		}

		for i := range data.Details {
			if data.Details[i].DNo == "" {
				data.Details[i].DNo = fmt.Sprintf("%05d", (DNo+1))
			}
		}

		query = `INSERT INTO tblvendorsectordtl 
				(SectorCode, DNo, SubSectorName, Active, CreateBy, CreateDt)
			VALUES `

		whenClause = whenClause[:0]
		args = args[:0]
		for _, detail := range data.Details {
			whenClause = append(whenClause, "(?, ?, ?, ?, ?, ?)")
			args = append(args, data.SectorCode, detail.DNo, detail.SubSectorName, detail.Active, data.LastUpdateBy, data.LastUpdateDate)
		}
		query += strings.Join(whenClause, ", ") + `
			ON DUPLICATE KEY UPDATE
				SubSectorName = VALUES(SubSectorName),
				Active = VALUES(Active);`
		
		fmt.Println("Query: ", query)
		fmt.Println("args: ", args)

		resultDtl, err = tx.ExecContext(ctx, query, args...)

		if err != nil {
			return nil, fmt.Errorf("error updating Vendor Sector: %w", err)
		}
	}

	if len(data.Details) == 0 {
		query = `UPDATE tblvendorsectordtl
			SET Active = 'N' WHERE SectorCode = ? AND ACTIVE = 'Y';`

		resultDtl, err = tx.ExecContext(ctx, query, data.SectorCode)

		if err != nil {
			return nil, fmt.Errorf("error updating Vendor Sector: %w", err)
		}
	}

	if resultDtl != nil {
		rowsAffectedDtl, err = resultDtl.RowsAffected()
		if err != nil {
			log.Printf("Failed to check row affected detail: %+v", err)
			return nil, fmt.Errorf("error check row affected: %w", err)
		}
	}

	fmt.Println("rows affected hdr: ", rowsAffectedHdr)
	fmt.Println("rows affected dtl: ", rowsAffectedDtl)

	if rowsAffectedHdr == 0 && rowsAffectedDtl == 0 {
		fmt.Println("no rows")
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, customerrors.ErrNoDataEdited
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.SectorCode, "VendorSector", data.LastUpdateDate)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		return nil, fmt.Errorf("error insert to log activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}


func (t *TblVendorSectorRepository) GetSector(ctx context.Context) ([]tblvendorsector.GetSector, error) {
	query := `SELECT SectorCode, SectorName FROM tblvendorsectorhdr WHERE Active = 'Y';`

	var data []tblvendorsector.GetSector

	if err := t.DB.SelectContext(ctx, &data, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			empty := make([]tblvendorsector.GetSector, 0)
			return empty, nil
		}
		return nil, fmt.Errorf("error Fetch Vendor Sector: %w", err)
	}

	return data, nil
}

func (t *TblVendorSectorRepository) GetSubSector(ctx context.Context, code string) ([]tblvendorsector.GetSubSector, error) {
	query := `SELECT DNo, SubSectorName FROM tblvendorsectordtl WHERE SectorCode = ? AND Active = 'Y'`

	var data []tblvendorsector.GetSubSector

	if err := t.DB.SelectContext(ctx, &data, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			empty := make([]tblvendorsector.GetSubSector, 0)
			return empty, nil
		}
		return nil, fmt.Errorf("error Fetch Vendor Sector: %w", err)
	}

	return data, nil
}