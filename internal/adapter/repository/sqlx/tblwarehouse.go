package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblwarehouse"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

// type RepositoryTblWarehouse interface {
// 	FetchWarehouses(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
// 	DetailWarehouse(ctx context.Context, code string) (*DetailTblWarehouse, error)
// 	Create(ctx context.Context, data *CreateTblWarehouse) (*CreateTblWarehouse, error)
// 	Update(ctx context.Context, data *UpdateTblWarehouse) (*UpdateTblWarehouse, error)
// }

type TblWarehouseRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblWarehouseRepository) FetchWarehouses(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	log.Printf("[INFO] Fetching warehouses with name: %s and ", name)

	var totalRecords int
	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblwarehouse WHERE WhsName LIKE ?"
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

	var warehouses []*tblwarehouse.ReadTblWarehouse
	query := "SELECT w.WhsCode, w.WhsName, w.CreateDt, c.CityCode, c.CityName, wc.WhsCtCode, wc.WhsCtName, w.PostalCd, w.Phone, w.Email, w.Fax, w.Mobile, w.ContactPerson, w.Remark FROM tblwarehouse w JOIN tblcity c ON w.CityCode = c.CityCode JOIN tblwarehousecategory wc ON w.WhsCtCode = wc.WhsCtCode WHERE WhsName LIKE ?  LIMIT ? OFFSET ?"
	log.Printf("[QUERY] %s", query)
	if err := t.DB.SelectContext(ctx, &warehouses, query, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[INFO] No warehouses found")
			return &pagination.PaginationResponse{
				Data:         make([]*tblwarehouse.ReadTblWarehouse, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		log.Printf("[ERROR] Failed to fetch warehouses: %v", err)
		return nil, fmt.Errorf("error fetching warehouses: %w", err)
	}

	j := offset
	result := make([]*tblwarehouse.ReadTblWarehouse, len(warehouses))
	for i, warehouse := range warehouses {
		j++
		result[i] = &tblwarehouse.ReadTblWarehouse{
			Number:        uint(j),
			WhsCode:       warehouse.WhsCode,
			WhsName:       warehouse.WhsName,
			WhsCtCode:     warehouse.WhsCtCode,
			WhsCtName:     warehouse.WhsCtName,
			CityCode:      warehouse.CityCode,
			CityName:      warehouse.CityName,
			PostalCd:      warehouse.PostalCd,
			Phone:         warehouse.Phone,
			Fax:           warehouse.Fax,
			Mobile:        warehouse.Mobile,
			ContactPerson: warehouse.ContactPerson,
			Remark:        warehouse.Remark,
			CreateDate:    share.FormatDate(warehouse.CreateDate),
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
	log.Printf("[INFO] Successfully fetched warehouses")
	return response, nil
}

// DetailWarehouse to get the details of a warehouse
func (t *TblWarehouseRepository) DetailWarehouse(ctx context.Context, code string) (*tblwarehouse.DetailTblWarehouse, error) {
	log.Printf("[INFO] Fetching details for warehouse with code: %s", code)

	query := "SELECT WhsCode, WhsName, Remark, Address, Phone, CreateBy, CreateDt FROM tblwarehouse WHERE WhsCode = ?"
	log.Printf("[QUERY] %s", query)

	var details tblwarehouse.DetailTblWarehouse
	if err := t.DB.GetContext(ctx, &details, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[WARN] No data found for code: %s", code)
			return nil, customerrors.ErrDataNotFound
		}
		log.Printf("[ERROR] Failed to fetch warehouse details: %v", err)
		return nil, fmt.Errorf("error fetching warehouse details: %w", err)
	}

	details.CreateDate = share.FormatDate(details.CreateDate)
	log.Printf("[INFO] Successfully fetched details for warehouse: %+v", details)
	return &details, nil
}

// Create to insert a new warehouse
func (t *TblWarehouseRepository) Create(ctx context.Context, data *tblwarehouse.CreateTblWarehouse) (*tblwarehouse.CreateTblWarehouse, error) {
	// Validasi input data
	if data.WhsCode == "" || data.WhsName == "" {
		return nil, fmt.Errorf("WhsCode and WhsName are required")
	}

	log.Printf("[INFO] Creating new warehouse: %+v", data)

	query := "INSERT INTO tblwarehouse (WhsCode, WhsName, WhsCtCode, Address, CityCode, PostalCd, Phone, Fax, Email, Mobile, ContactPerson, Remark, CreateBy, CreateDt) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	log.Printf("[QUERY] %s", query)

	if _, err := t.DB.ExecContext(ctx, query, data.WhsCode, data.WhsName, data.WhsCtCode, data.Address, data.CityCode, data.PostalCd, data.Phone, data.Fax, data.Email, data.Mobile, data.ContactPerson, data.Remark, data.CreateBy, data.CreateDate); err != nil {
		log.Printf("[ERROR] Failed to create warehouse: %v", err)
		return nil, fmt.Errorf("error creating warehouse: %w", err)
	}

	log.Printf("[INFO] Successfully created warehouse: %+v", data)
	return data, nil
}

// Update to modify an existing warehouse
func (t *TblWarehouseRepository) Update(ctx context.Context, data *tblwarehouse.UpdateTblWarehouse) (*tblwarehouse.UpdateTblWarehouse, error) {
	// Validasi input data
	if data.WhsCode == "" || data.WhsName == "" {
		return nil, fmt.Errorf("WhsCode and WhsName are required")
	}

	log.Printf("[INFO] Updating warehouse: %+v", data)

	query := "SELECT WhsName FROM tblwarehouse WHERE WhsCode = ?"
	log.Printf("[QUERY] %s", query)

	var check tblwarehouse.DetailTblWarehouse
	if err := t.DB.GetContext(ctx, &check, query, data.WhsCode); err != nil {
		log.Printf("[ERROR] Failed to check existing warehouse: %v", err)
		return nil, err
	}

	// Start transaction
	query = "UPDATE tblwarehouse SET WhsName = ?, WhsCtCode = ?, Address = ?, CityCode = ?, PostalCd = ?, Phone = ?, Fax = ?, Email = ?, Mobile = ?, ContactPerson = ?, Remark = ?, LastUpBy = ?, LastUpDt = ? WHERE WhsCode = ?"
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

	if _, err = tx.ExecContext(ctx, query, data.WhsName, data.WhsCtCode, data.Address, data.CityCode, data.PostalCd, data.Phone, data.Fax, data.Email, data.Mobile, data.ContactPerson, data.Remark, data.LastUpBy, data.LastUpdateDate, data.WhsCode); err != nil {
		_ = tx.Rollback()
		log.Printf("[ERROR] Failed to update warehouse: %v", err)
		return nil, fmt.Errorf("error updating warehouse: %w", err)
	}

	// Log activity
	logActivityQuery := "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"
	log.Printf("[QUERY] %s", logActivityQuery)

	if _, err = tx.ExecContext(ctx, logActivityQuery, data.LastUpBy, data.WhsCode, "Warehouse", data.LastUpdateDate); err != nil {
		_ = tx.Rollback()
		log.Printf("[ERROR] Failed to log activity: %v", err)
		return nil, fmt.Errorf("error logging activity: %w", err)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction: %v", err)
		_ = tx.Rollback()
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("[INFO] Successfully updated warehouse: %+v", data)
	return data, nil
}
