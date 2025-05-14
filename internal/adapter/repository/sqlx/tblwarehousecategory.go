package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	// "gitlab.com/ayaka/internal/domain/logactivity"
	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblwarehousecategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblWarehouseCategoryRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblWarehouseCategoryRepository) FetchWarehouseCategories(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	log.Printf("[INFO] Fetching warehouse categories with name: %s", name)

	var totalRecords int
	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblwarehousecategory WHERE WhsCtName LIKE ?"
	log.Printf("[QUERY] %s", countQuery)
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search); err != nil {
		log.Printf("[ERROR] Failed to count records: %v", err)
		return nil, fmt.Errorf("error counting records: %w", err)
	}

	var totalPages, offset int
	if param != nil {
		totalPages, offset = pagination.CountPagination(param, totalRecords)
		log.Printf("[INFO] Pagination calculated - TotalRecords: %d, TotalPages: %d, Offset: %d", totalRecords, totalPages, offset)
	} else {
		param = &pagination.PaginationParam{
			PageSize: totalRecords,
			Page:     1,
		}
		totalPages = 1
		offset = 0
	}

	var categories []*tblwarehousecategory.ReadTblWarehouseCategory
	query := "SELECT WhsCtCode, WhsCtName, CreateDt FROM tblwarehousecategory WHERE WhsCtName LIKE ? LIMIT ? OFFSET ?"
	log.Printf("[QUERY] %s", query)
	if err := t.DB.SelectContext(ctx, &categories, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[INFO] No warehouse categories found")
			return &pagination.PaginationResponse{
				Data:         make([]*tblwarehousecategory.ReadTblWarehouseCategory, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		log.Printf("[ERROR] Failed to fetch warehouse categories: %v", err)
		return nil, fmt.Errorf("error fetching warehouse categories: %w", err)
	}

	j := offset
	result := make([]*tblwarehousecategory.ReadTblWarehouseCategory, len(categories))
	for i, category := range categories {
		j++
		result[i] = &tblwarehousecategory.ReadTblWarehouseCategory{
			Number:     uint(j),
			WhsCtCode:  category.WhsCtCode,
			WhsCtName:  category.WhsCtName,
			CreateDate: share.FormatDate(category.CreateDate),
		}
	}

	response := &pagination.PaginationResponse{
		Data:         result,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}
	log.Printf("[INFO] Successfully fetched warehouse categories")
	return response, nil
}

func (t *TblWarehouseCategoryRepository) DetailWarehouseCategory(ctx context.Context, code string) (*tblwarehousecategory.DetailTblWarehouseCategory, error) {
	log.Printf("[INFO] Fetching details for warehouse category with code: %s", code)

	query := "SELECT WhsCtCode, WhsCtName, CreateBy, CreateDt FROM tblwarehousecategory WHERE WhsCtCode = ?"
	log.Printf("[QUERY] %s", query)

	var details tblwarehousecategory.DetailTblWarehouseCategory
	if err := t.DB.GetContext(ctx, &details, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[WARN] No data found for code: %s", code)
			return nil, customerrors.ErrDataNotFound
		}
		log.Printf("[ERROR] Failed to fetch warehouse category details: %v", err)
		return nil, fmt.Errorf("error fetching warehouse category details: %w", err)
	}

	details.CreateDate = share.FormatDate(details.CreateDate)
	log.Printf("[INFO] Successfully fetched details for warehouse category: %+v", details)
	return &details, nil
}

func (t *TblWarehouseCategoryRepository) Create(ctx context.Context, data *tblwarehousecategory.CreateTblWarehouseCategory) (*tblwarehousecategory.CreateTblWarehouseCategory, error) {
	log.Printf("[INFO] Creating new warehouse category: %+v", data)

	query := "INSERT INTO tblwarehousecategory (WhsCtCode, WhsCtName, CreateBy, CreateDt) VALUES (?, ?, ?, ?)"
	log.Printf("[QUERY] %s", query)

	if _, err := t.DB.ExecContext(ctx, query, data.WhsCtCode, data.WhsCtName, data.CreateBy, data.CreateDate); err != nil {
		log.Printf("[ERROR] Failed to create warehouse category: %v", err)
		return nil, fmt.Errorf("error creating warehouse category: %w", err)
	}

	log.Printf("[INFO] Successfully created warehouse category: %+v", data)
	return data, nil
}

func (t *TblWarehouseCategoryRepository) Update(ctx context.Context, data *tblwarehousecategory.UpdateTblWarehouseCategory) (*tblwarehousecategory.UpdateTblWarehouseCategory, error) {
	log.Printf("[INFO] Updating warehouse category: %+v", data)

	query := "SELECT WhsCtName FROM tblwarehousecategory WHERE WhsCtCode = ?"
	log.Printf("[QUERY] %s", query)

	var check tblwarehousecategory.DetailTblWarehouseCategory
	if err := t.DB.GetContext(ctx, &check, query, data.WhsCtCode); err != nil {
		log.Printf("[ERROR] Failed to check existing warehouse category: %v", err)
		return nil, err
	}

	if check.WhsCtName == data.WhsCtName {
		log.Printf("[WARN] No changes detected for warehouse category: %s", data.WhsCtCode)
		return data, customerrors.ErrNoDataEdited
	}

	// Start transaction
	query = "UPDATE tblwarehousecategory SET WhsCtName = ?, LastUpBy = ?, LastUpDt = ? WHERE WhsCtCode = ?"
	log.Printf("[QUERY] %s", query)

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to start transaction: %v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if _, err = tx.ExecContext(ctx, query, data.WhsCtName, data.LastUpdateBy, data.LastUpdateDate, data.WhsCtCode); err != nil {
		_ = tx.Rollback()
		log.Printf("[ERROR] Failed to update warehouse category: %v", err)
		return nil, fmt.Errorf("error updating warehouse category: %w", err)
	}

	// Log activity
	logActivityQuery := "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"
	log.Printf("[QUERY] %s", logActivityQuery)

	if _, err = tx.ExecContext(ctx, logActivityQuery, data.LastUpdateBy, data.WhsCtCode, "WarehouseCategory", data.LastUpdateDate); err != nil {
		_ = tx.Rollback()
		log.Printf("[ERROR] Failed to insert log activity: %v", err)
		return nil, fmt.Errorf("error logging activity: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction: %v", err)
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("[INFO] Successfully updated warehouse category: %+v", data)
	return data, nil
}
