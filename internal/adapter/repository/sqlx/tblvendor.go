package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblmastervendor"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblVendorRepository) Create(ctx context.Context, data *tblmastervendor.Create) (*tblmastervendor.Create, error) {
	query := `INSERT INTO tblvendorhdr 
		(
			VendorCode, 
			VendorName, 
			VendorCatCode, 
			Address, 
			CityCode, 
			PostalCode, 
			Website, 
			HeadOffice, 
			Phone, 
			Mobile, 
			Email, 
			Remark,
			CreateDt,
			CreateBy
		) VALUES 
		`
	var args []interface{}
	var placeholders []string

	placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);")
	args = append(args,
		data.VendorCode,
		data.VendorName,
		data.VendorCatCode,
		data.Address,
		data.CityCode,
		data.PostalCode,
		data.Website,
		data.HeadOffice,
		data.Phone,
		data.Mobile,
		data.Email,
		data.Remark,
		data.CreateDate,
		data.CreateBy)

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

	// insert header
	query += strings.Join(placeholders, " ")
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Vendor Header: %w", err)
	}

	if len(data.ContactVendor) > 0 {
		query = `INSERT INTO tblcontactvendordtl
			(
				VendorCode,
				DNo,
				Name,
				Number,
				Position,
				Type,
				Active,
				CreateDt,
				CreateBy
			) VALUES 
			`
		placeholders = placeholders[:0] // Reset placeholders
		args = args[:0]                 // Reset args
		for i, contact := range data.ContactVendor {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.VendorCode,
				fmt.Sprintf("%05d", i+1),
				contact.Name,
				contact.Number,
				contact.Position,
				contact.Type,
				"Y",
				data.CreateDate,
				data.CreateBy)
		}

		// insert contact
		query += strings.Join(placeholders, ", ") + ";"
		fmt.Println("query contact: ", query)
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Create Contact Vendor: %w", err)
		}
	}

	if len(data.ItemCategoryVendor) > 0 {
		query = `INSERT INTO tblitemcategoryvendordtl
				(
					VendorCode,
					ItCatCode,
					Active,
					CreateDt,
					CreateBy
				) VALUES 
			`
		placeholders = placeholders[:0] // Reset placeholders
		args = args[:0]                 // Reset args
		for _, itemCategory := range data.ItemCategoryVendor {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
			args = append(args,
				data.VendorCode,
				itemCategory.ItemCategoryCode,
				"Y",
				data.CreateDate,
				data.CreateBy)
		}

		// insert item category
		query += strings.Join(placeholders, ", ") + ";"
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Create Item Category Vendor: %w", err)
		}
	}

	if len(data.SectorVendor) > 0 {
		query = `INSERT INTO tblsectorvendordtl
			(
				VendorCode,
				VendorSectorCode,
				DNoVendorSector,
				Active,
				CreateDt,
				CreateBy
			) VALUES 
			`
		placeholders = placeholders[:0] // Reset placeholders
		args = args[:0]                 // Reset args
		for _, sector := range data.SectorVendor {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.VendorCode,
				sector.VendorSectorCode,
				sector.DNoVendorSector,
				"Y",
				data.CreateDate,
				data.CreateBy)
		}

		// insert sector vendor
		query += strings.Join(placeholders, ", ") + ";"
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Create Sector Vendor: %w", err)
		}
	}

	if len(data.RatingVendor) > 0 {
		query = `INSERT INTO tblratingvendordtl
			(
				VendorCode,
				VendorRatingCode,
				Value,
				Active,
				CreateDt,
				CreateBy
			) VALUES
			`
		placeholders = placeholders[:0] // Reset placeholders
		args = args[:0]                 // Reset args
		for _, rating := range data.RatingVendor {
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?)")
			args = append(args,
				data.VendorCode,
				rating.VendorRatingCode,
				rating.Value,
				"Y",
				data.CreateDate,
				data.CreateBy)
		}
		// insert rating vendor
		query += strings.Join(placeholders, ", ") + ";"
		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Create Rating Vendor: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %+v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return data, nil
}

func (t *TblVendorRepository) Fetch(ctx context.Context, name, cat string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int
	var args []interface{}

	search := "%" + name + "%"
	countQuery := "SELECT COUNT(*) FROM tblvendorhdr WHERE VendorCode LIKE ? OR VendorName LIKE ?"
	args = append(args, search, search)

	if cat != "" {
		countQuery += " AND VendorCatCode = ?"
		args = append(args, cat)
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

	var items []*tblmastervendor.Read
	query := `SELECT
			v.VendorCode,
			v.VendorName,
			v.Address,
			c.CityName,
			v.CreateDt
		FROM tblvendorhdr v JOIN tblcity c ON v.CityCode = c.CityCode
		WHERE v.VendorCode LIKE ? OR v.VendorName LIKE ?`

	if cat != "" {
		query += " AND v.VendorCatCode = ?"
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)

	if err := t.DB.SelectContext(ctx, &items, query, args...); err != nil {
		fmt.Println("error: ", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblmastervendor.Read, 0),
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
		docsNo[i] = item.VendorCode
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

func (t *TblVendorRepository) Detail(ctx context.Context, vendorCode string) (*tblmastervendor.Detail, error) {
	query := `SELECT
			VendorCode,
			VendorName,
			VendorCatCode,
			Address,
			CityCode,
			PostalCode,
			Website,
			HeadOffice,
			Phone,
			Mobile,
			Email,
			Remark
		FROM tblvendorhdr
		WHERE VendorCode = ?;`

	var detail tblmastervendor.Detail
	if err := t.DB.GetContext(ctx, &detail, query, vendorCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, fmt.Errorf("error fetching vendor detail: %w", err)
	}

	detail.ContactVendor = make([]tblmastervendor.ContactVendorDetail, 0)
	query = `SELECT
				VendorCode,
				DNo,
				Name,
				Number,
				Position,
				Type
			FROM tblcontactvendordtl
			WHERE VendorCode = ? AND Active = 'Y'`
	if err := t.DB.SelectContext(ctx, &detail.ContactVendor, query, vendorCode); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error fetching contact vendor: %w", err)
		}
	}

	detail.ItemCategoryVendor = make([]tblmastervendor.ItemCategoryVendorDetail, 0)
	query = `SELECT
				VendorCode,
				ItCatCode
			FROM tblitemcategoryvendordtl
			WHERE VendorCode = ? AND Active = 'Y'`
	if err := t.DB.SelectContext(ctx, &detail.ItemCategoryVendor, query, vendorCode); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error fetching item category vendor: %w", err)
		}
	}

	detail.SectorVendor = make([]tblmastervendor.SectorVendorDetail, 0)
	query = `SELECT
				VendorCode,
				VendorSectorCode,
				DNoVendorSector
			FROM tblsectorvendordtl
			WHERE VendorCode = ? AND Active = 'Y'`
	if err := t.DB.SelectContext(ctx, &detail.SectorVendor, query, vendorCode); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error fetching sector vendor: %w", err)
		}
	}

	detail.RatingVendor = make([]tblmastervendor.RatingVendorDetail, 0)
	query = `SELECT
				VendorCode,
				VendorRatingCode,
				Value
			FROM tblratingvendordtl
			WHERE VendorCode = ? AND Active = 'Y'`
	if err := t.DB.SelectContext(ctx, &detail.RatingVendor, query, vendorCode); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error fetching rating vendor: %w", err)
		}
	}

	return &detail, nil
}

func (t *TblVendorRepository) Update(ctx context.Context, data *tblmastervendor.Update) (*tblmastervendor.Update, error) {
	query := `UPDATE tblvendorhdr SET 
			VendorName = ?, 
			VendorCatCode = ?, 
			Address = ?, 
			CityCode = ?, 
			PostalCode = ?, 
			Website = ?, 
			HeadOffice = ?, 
			Phone = ?, 
			Mobile = ?, 
			Email = ?, 
			Remark = ?
		WHERE VendorCode = ?`

	var args []interface{}

	args = append(args,
		data.VendorName,
		data.VendorCatCode,
		data.Address,
		data.CityCode,
		data.PostalCode,
		data.Website,
		data.HeadOffice,
		data.Phone,
		data.Mobile,
		data.Email,
		data.Remark,
		data.VendorCode,
	)

	// start transaction
	var err error
	var resultHdr, resultContact, resultItCategory, resultSector, resultRating sql.Result
	var rowsAffectedHdr int64 = 0
	var rowsAffectedContact int64 = 0
	var rowsAffectedSector int64 = 0
	var rowsAffectedItemCategory int64 = 0
	var rowsAffectedRating int64 = 0

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

	// Header
	fmt.Println("--Update Header--")
	resultHdr, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Update Vendor Header: %w", err)
	}

	// Contact Vendor
	if len(data.ContactVendor) > 0 {
		query = `SELECT VendorCode, DNo, Name, Number, Active, Position, Type FROM tblcontactvendordtl WHERE VendorCode = ? AND Active = 'Y'`
		var existingContacts []tblmastervendor.ContactVendor

		if err := tx.SelectContext(ctx, &existingContacts, query, data.VendorCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Detailed error: %+v", err)
				return nil, fmt.Errorf("error fetching existing contact vendor: %w", err)
			}
		}

		var deletedContacts []tblmastervendor.ContactVendor
		for _, exist := range existingContacts {
			found := false
			for _, contact := range data.ContactVendor {
				if exist.DNo == contact.DNo {
					found = true
					break
				}
			}
			if !found {
				exist.Active = booldatatype.FromBool(false)
				deletedContacts = append(deletedContacts, exist)
			}
		}

		if len(deletedContacts) > 0 {
			data.ContactVendor = append(data.ContactVendor, deletedContacts...)
		}

		query = `INSERT INTO tblcontactvendordtl 
				(
					VendorCode,
					DNo,
					Name,
					Number,
					Position,
					Type,
					Active
				) VALUES `
		var placeholdersContact []string
		var argsContact []interface{}
		for _, contact := range data.ContactVendor {
			placeholdersContact = append(placeholdersContact, "(?, ?, ?, ?, ?, ?, ?)")
			argsContact = append(argsContact,
				data.VendorCode,
				contact.DNo,
				contact.Name,
				contact.Number,
				contact.Position,
				contact.Type,
				contact.Active)
		}

		query += strings.Join(placeholdersContact, ", ") + `
			ON DUPLICATE KEY UPDATE
				Name = VALUES(Name),
				Number = VALUES(Number),
				Position = VALUES(Position),
				Type = VALUES(Type),
				Active = VALUES(Active);`

		fmt.Println("--Update Contact 1--")
		if resultContact, err = tx.ExecContext(ctx, query, argsContact...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Contact Vendor: %w", err)
		}
	}

	// If slice of contact vendor is empty
	if len(data.ContactVendor) == 0 {
		query = `UPDATE tblcontactvendordtl
				SET Active = 'N' WHERE VendorCode = ? AND Active = 'Y';`

		fmt.Println("--Update Contact 2--")
		if resultItCategory, err = tx.ExecContext(ctx, query, data.VendorCode); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Contact Vendor second condition: %w", err)
		}
	}
	//===========================================================================================================================================================//

	// Item Category Vendor
	if len(data.ItemCategoryVendor) > 0 {
		query = `SELECT VendorCode, ItCatCode, Active FROM tblitemcategoryvendordtl WHERE VendorCode = ? AND Active = 'Y'`
		var existingItCats []tblmastervendor.ItemCategoryVendor

		if err := tx.SelectContext(ctx, &existingItCats, query, data.VendorCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Detailed error: %+v", err)
				return nil, fmt.Errorf("error fetching existing item category vendor: %w", err)
			}
		}

		var deletedItCats []tblmastervendor.ItemCategoryVendor
		for _, exist := range existingItCats {
			found := false
			for _, detail := range data.ItemCategoryVendor {
				if exist.ItemCategoryCode == detail.ItemCategoryCode {
					found = true
					break
				}
			}
			if !found {
				exist.Active = booldatatype.FromBool(false)
				deletedItCats = append(deletedItCats, exist)
			}
		}

		if len(deletedItCats) > 0 {
			data.ItemCategoryVendor = append(data.ItemCategoryVendor, deletedItCats...)
		}

		query = `INSERT INTO tblitemcategoryvendordtl 
				(
					VendorCode,
					ItCatCode,
					Active
				) VALUES `
		var placeholdersItCat []string
		var argsItCat []interface{}
		for _, detail := range data.ItemCategoryVendor {
			placeholdersItCat = append(placeholdersItCat, "(?, ?, ?)")
			argsItCat = append(argsItCat,
				data.VendorCode,
				detail.ItemCategoryCode,
				detail.Active)
		}

		query += strings.Join(placeholdersItCat, ", ") + `
			ON DUPLICATE KEY UPDATE
				Active = VALUES(Active);`

		fmt.Println("--Update Item Category 1--")
		if resultItCategory, err = tx.ExecContext(ctx, query, argsItCat...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update ItCategory Vendor: %w", err)
		}
	}

	// If slice of item category vendor is empty
	if len(data.ItemCategoryVendor) == 0 {
		query = `UPDATE tblitemcategoryvendordtl
				SET Active = 'N' WHERE VendorCode = ? AND Active = 'Y';`

		fmt.Println("--Update Item Category 2--")
		if resultItCategory, err = tx.ExecContext(ctx, query, data.VendorCode); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Item Category Vendor second condition: %w", err)
		}
	}
	//===========================================================================================================================================================//

	// Sector Vendor
	if len(data.SectorVendor) > 0 {
		query = `SELECT VendorCode, VendorSectorCode, DNoVendorSector, Active FROM tblsectorvendordtl WHERE VendorCode = ? AND Active = 'Y'`
		var existingSectors []tblmastervendor.SectorVendor

		if err := tx.SelectContext(ctx, &existingSectors, query, data.VendorCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Detailed error: %+v", err)
				return nil, fmt.Errorf("error fetching existing sector vendor: %w", err)
			}
		}

		var deletedSectors []tblmastervendor.SectorVendor
		for _, exist := range existingSectors {
			found := false
			for _, detail := range data.SectorVendor {
				if exist.VendorSectorCode == detail.VendorSectorCode && exist.DNoVendorSector == detail.DNoVendorSector {
					found = true
					break
				}
			}
			if !found {
				exist.Active = booldatatype.FromBool(false)
				deletedSectors = append(deletedSectors, exist)
			}
		}

		if len(deletedSectors) > 0 {
			data.SectorVendor = append(data.SectorVendor, deletedSectors...)
		}

		query = `INSERT INTO tblsectorvendordtl 
					(
						VendorCode,
						VendorSectorCode,
						DNoVendorSector,
						Active
					) VALUES `
		var placeholdersSector []string
		var argsSector []interface{}
		for _, detail := range data.SectorVendor {
			placeholdersSector = append(placeholdersSector, "(?, ?, ?, ?)")
			argsSector = append(argsSector,
				data.VendorCode,
				detail.VendorSectorCode,
				detail.DNoVendorSector,
				detail.Active)
		}

		query += strings.Join(placeholdersSector, ", ") + `
				ON DUPLICATE KEY UPDATE
					Active = VALUES(Active);`

		fmt.Println("--Update Sector 1--")
		if resultSector, err = tx.ExecContext(ctx, query, argsSector...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Sector Vendor: %w", err)
		}
	}

	// If slice of sector vendor is empty
	if len(data.SectorVendor) == 0 {
		query = `UPDATE tblsectorvendordtl
					SET Active = 'N' WHERE VendorCode = ? AND Active = 'Y';`

		fmt.Println("--Update Sector 2--")
		if resultSector, err = tx.ExecContext(ctx, query, data.VendorCode); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Sector Vendor second condition: %w", err)
		}
	}
	//===========================================================================================================================================================//

	// Rating Vendor
	if len(data.RatingVendor) > 0 {
		query = `SELECT VendorCode, VendorRatingCode, Value, Active FROM tblratingvendordtl WHERE VendorCode = ? AND Active = 'Y'`
		var existingRatings []tblmastervendor.RatingVendor

		if err := tx.SelectContext(ctx, &existingRatings, query, data.VendorCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Detailed error: %+v", err)
				return nil, fmt.Errorf("error fetching existing rating vendor: %w", err)
			}
		}

		var deletedRatings []tblmastervendor.RatingVendor
		for _, exist := range existingRatings {
			found := false
			for _, detail := range data.RatingVendor {
				if exist.VendorRatingCode == detail.VendorRatingCode {
					found = true
					break
				}
			}
			if !found {
				exist.Active = booldatatype.FromBool(false)
				deletedRatings = append(deletedRatings, exist)
			}
		}

		if len(deletedRatings) > 0 {
			data.RatingVendor = append(data.RatingVendor, deletedRatings...)
		}

		query = `INSERT INTO tblratingvendordtl 
						(
							VendorCode,
							VendorRatingCode,
							Value,
							Active
						) VALUES `
		var placeholdersRating []string
		var argsRating []interface{}
		for _, detail := range data.RatingVendor {
			placeholdersRating = append(placeholdersRating, "(?, ?, ?, ?)")
			argsRating = append(argsRating,
				data.VendorCode,
				detail.VendorRatingCode,
				detail.Value,
				detail.Active)
		}

		query += strings.Join(placeholdersRating, ", ") + `
					ON DUPLICATE KEY UPDATE
						Value = VALUES(Value),
						Active = VALUES(Active);`

		fmt.Println("--Update Rating 1--")
		if resultRating, err = tx.ExecContext(ctx, query, argsRating...); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Rating Vendor: %w", err)
		}
	}

	// If slice of rating vendor is empty
	if len(data.RatingVendor) == 0 {
		query = `UPDATE tblratingvendordtl
						SET Active = 'N' WHERE VendorCode = ? AND Active = 'Y';`

		fmt.Println("--Update Rating 2--")
		if resultRating, err = tx.ExecContext(ctx, query, data.VendorCode); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error Update Rating Vendor second condition: %w", err)
		}
	}
	//===========================================================================================================================================================//

	// checking rows affected
	if resultHdr != nil {
		if rowsAffectedHdr, err = resultHdr.RowsAffected(); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error getting rows affected for header: %w", err)
		}
	}

	if resultContact != nil {
		if rowsAffectedContact, err = resultContact.RowsAffected(); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error getting rows affected for contact: %w", err)
		}
	}

	if resultItCategory != nil {
		if rowsAffectedItemCategory, err = resultItCategory.RowsAffected(); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error getting rows affected for ItCategory: %w", err)
		}
	}

	if resultSector != nil {
		if rowsAffectedSector, err = resultSector.RowsAffected(); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error getting rows affected for Sector: %w", err)
		}
	}

	if resultRating != nil {
		if rowsAffectedRating, err = resultRating.RowsAffected(); err != nil {
			log.Printf("Detailed error: %+v", err)
			return nil, fmt.Errorf("error getting rows affected for Rating: %w", err)
		}
	}

	if rowsAffectedHdr == 0 && rowsAffectedContact == 0 && rowsAffectedItemCategory == 0 && rowsAffectedSector == 0 && rowsAffectedRating == 0 {
		fmt.Println("NO DATA UPDATED-------")
		return data, customerrors.ErrNoDataEdited
	}

	// Insert to log activity
	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	fmt.Println("--Update Log--")
	if _, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.VendorCode, "MasterVendor", data.LastUpdateDate); err != nil {
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
