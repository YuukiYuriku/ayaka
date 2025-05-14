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
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblmasteritem"

	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblItemRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblItemRepository) FetchItems(ctx context.Context, name, category, act string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	searchName := "%" + name + "%"
	status := "%" + act + "%"

	if strings.TrimSpace(category) != "" {
		countQuery := "SELECT COUNT(*) FROM tblitem WHERE (ItName LIKE ? OR ItCode LIKE ?) AND ItCtCode = ? AND ActInd LIKE ?"
		if err := t.DB.GetContext(ctx, &totalRecords, countQuery, searchName, searchName, category, status); err != nil {
			return nil, fmt.Errorf("error counting records: %w", err)
		}
	} else {
		countQuery := "SELECT COUNT(*) FROM tblitem WHERE (ItName LIKE ? OR ItCode LIKE ?) AND ActInd LIKE ?"
		if err := t.DB.GetContext(ctx, &totalRecords, countQuery, searchName, searchName, status); err != nil {
			return nil, fmt.Errorf("error counting records: %w", err)
		}
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

	var items []*tblmasteritem.Read
	var query string

	if strings.TrimSpace(category) != "" {
		query = `SELECT i.ItCode,
				i.ItName,
				i.ItName,
				u.UomName,
				i.ItCodeInternal,
				i.ForeignName,
				i.ItCodeOld,
				c.ItCtName,
				i.Specification,
				i.ActInd,
				i.ItemRequestDocNo,
				i.PurchaseUomCode,
				i.HSCode,
				i.Remark,
				i.InventoryItemInd,
				i.SalesItemInd,
				i.PurchaseItemInd,
				i.ServiceItemInd,
				i.TaxLiableInd,
				i.CreateDt
				FROM tblitem i
				JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
				JOIN tblitemcategory c ON i.ItCtCode = c.ItCtCode
				WHERE (i.ItCode LIKE ? OR i.ItName LIKE ?)
				AND i.ItCtCode = ?
				AND i.ActInd LIKE ?
				LIMIT ? OFFSET ?`
	} else {
		query = `SELECT i.ItCode,
				i.ItName,
				i.ItName,
				u.UomName,
				i.ItCodeInternal,
				i.ForeignName,
				i.ItCodeOld,
				c.ItCtName,
				i.Specification,
				i.ActInd,
				i.ItemRequestDocNo,
				i.PurchaseUomCode,
				i.HSCode,
				i.Remark,
				i.InventoryItemInd,
				i.SalesItemInd,
				i.PurchaseItemInd,
				i.ServiceItemInd,
				i.TaxLiableInd,
				i.CreateDt
				FROM tblitem i
				JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
				JOIN tblitemcategory c ON i.ItCtCode = c.ItCtCode
				WHERE (ItCode LIKE ? OR ItName LIKE ?)
				AND i.ActInd LIKE ?
				LIMIT ? OFFSET ?`
	}

	if strings.TrimSpace(category) != "" {
		if err := t.DB.SelectContext(ctx, &items, query, searchName, searchName, category, status, param.PageSize, offset); err != nil {
			if errors.Is(err, sql.ErrNoRows) || items == nil {
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
			return nil, fmt.Errorf("error Fetch Items: %w", err)
		}
	} else {
		if err := t.DB.SelectContext(ctx, &items, query, searchName, searchName, status, param.PageSize, offset); err != nil {
			if errors.Is(err, sql.ErrNoRows) || items == nil {
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
			return nil, fmt.Errorf("error Fetch Items: %w", err)
		}
	}

	// proses data
	j := offset
	for _, item := range items {
		j++
		item.Number = uint(j)
		item.CreateDate = share.FormatDate(item.CreateDate)
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

	return response, nil
}

func (t *TblItemRepository) Detail(ctx context.Context, search string) (*tblmasteritem.Detail, error) {
	query := `SELECT
				i.ItemRequestDocNo,
				i.PurchaseUomCode,
				u.UomName,
				i.HSCode,
				i.Remark,
				i.InventoryItemInd,
				i.SalesItemInd,
				i.PurchaseItemInd,
				i.ServiceItemInd,
				i.TaxLiableInd
				FROM tblitem i
				JOIN tbluom u ON i.PurchaseUOMCode = u.UomCode
				WHERE i.ItCode = ?`

	var item tblmasteritem.Detail

	if err := t.DB.GetContext(ctx, &item, query, search); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customerrors.ErrDataNotFound
		}
		return nil, fmt.Errorf("error Get Items: %w", err)
	}

	return &item, nil
}

func (t *TblItemRepository) Create(ctx context.Context, data *tblmasteritem.Create, confirm bool) (*tblmasteritem.Create, error) {
	if !confirm {
		qcheck := "SELECT ItCode, ItName FROM tblitem WHERE ItName = ? LIMIT 1"
		var check1 tblmasteritem.Check
		if err := t.DB.GetContext(ctx, &check1, qcheck, data.ItemName); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("error check name: %w", err)
			}
		}

		qcheck = "SELECT ItCode, ItName FROM tblitem WHERE ItCodeInternal = ? LIMIT 1"
		var check2 tblmasteritem.Check
		if err := t.DB.GetContext(ctx, &check2, qcheck, data.LocalCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("error check code: %w", err)
			}
		}

		var errorMessages []string
		if check1.ItemCode != "" {
			errorMessages = append(errorMessages, fmt.Sprintf("Item Name already exists on Item Code: %s, Item Name: %s", check1.ItemCode, check1.ItemName))
		}
		if check2.ItemCode != "" {
			if check1.ItemCode != "" {
				errorMessages = append(errorMessages, "\n")
			}
			errorMessages = append(errorMessages, fmt.Sprintf("Item Local Code already exists on Item Code: %s, Item Name: %s", check2.ItemCode, check2.ItemName))
		}

		if len(errorMessages) > 0 {
			return nil, errors.New(strings.Join(errorMessages, "; "))
		}
	}

	query := "SELECT ItCode, ItName FROM tblitem ORDER BY CreateDt DESC LIMIT 1"
	var check tblmasteritem.Check
	var id string
	prefixId := "ITC0004"
	if err := t.DB.GetContext(ctx, &check, query); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error check code: %w", err)
		}
		id = fmt.Sprintf("%s-00001", prefixId)
	} else {
		id, err = formatid.FormatId(check.ItemCode, prefixId)
		if err != nil {
			return nil, fmt.Errorf("error format id: %w", err)
		}
	}
	data.ItemCode = id

	query = `INSERT INTO tblitem
				(
					ItCode,
					ItName,
					ItCodeInternal,
					ForeignName,
					ItCodeOld,
					ItCtCode,
					Specification,
					ActInd,
					CreateDt,
					ItemRequestDocNo,
					PurchaseUomCode,
					HSCode,
					Remark,
					InventoryItemInd,
					SalesItemInd,
					PurchaseItemInd,
					ServiceItemInd,
					TaxLiableInd,
					CreateBy,
					ItScCode,
					SalesUomCode,
					SalesUOMCode2,
					InventoryUOMCode,
					InventoryUOMCode2,
					InventoryUOMCode3,
					PlanningUomCode,
					PlanningUomCode2
				)
				VALUES
				(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.ItemCode,
		data.ItemName,
		data.LocalCode,
		data.ForeignName,
		data.OldCode,
		data.Category,
		data.Spesification,
		data.Active,
		data.CreateDate,
		data.ItemRequest,
		data.Uom,
		data.HSCode,
		data.Remark,
		data.InventoryItem,
		data.SalesItem,
		data.PurchaseItem,
		data.ServiceItem,
		data.TaxLiable,
		data.CreateBy,
		data.Source,
		data.Uom,
		data.Uom,
		data.Uom,
		data.Uom,
		data.Uom,
		data.Uom,
		data.Uom,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Item: %w", err)
	}

	return data, nil
}

func (t *TblItemRepository) Update(ctx context.Context, data *tblmasteritem.Update, confirm bool) (*tblmasteritem.Update, error) {
	if !confirm {
		qcheck := "SELECT ItCode, ItName FROM tblitem WHERE ItName = ? LIMIT 1"
		var check1 tblmasteritem.Check
		if err := t.DB.GetContext(ctx, &check1, qcheck, data.ItemName); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("error check name: %w", err)
			}
		}

		qcheck = "SELECT ItCode, ItName FROM tblitem WHERE ItCodeInternal = ? LIMIT 1"
		var check2 tblmasteritem.Check
		if err := t.DB.GetContext(ctx, &check2, qcheck, data.LocalCode); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("error check code: %w", err)
			}
		}

		var errorMessages []string
		if check1.ItemCode != "" && check1.ItemCode != data.ItemCode {
			errorMessages = append(errorMessages, fmt.Sprintf("Item Name already exists on Item Code: %s, Item Name: %s", check1.ItemCode, check1.ItemName))
		}
		if check2.ItemCode != "" && check2.ItemCode != data.ItemCode {
			errorMessages = append(errorMessages, fmt.Sprintf("Item Local Code already exists on Item Code: %s, Item Name: %s", check2.ItemCode, check2.ItemName))
		}

		if len(errorMessages) > 0 {
			return nil, errors.New(strings.Join(errorMessages, "; "))
		}
	}

	query := `UPDATE tblitem SET
			ItName = ?,
			ItCodeInternal = ?,
			ForeignName = ?,
			ItCodeOld = ?,
			Specification = ?,
			ActInd = ?,
			HSCode = ?,
			Remark = ?,
			InventoryItemInd = ?,
			SalesItemInd = ?,
			PurchaseItemInd = ?,
			ServiceItemInd = ?,
			TaxLiableInd = ?,
			LastUpBy = ?,
			LastUpDt = ?
			WHERE ItCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, query,
		data.ItemName,
		data.LocalCode,
		data.ForeignName,
		data.OldCode,
		data.Spesification,
		data.Active,
		data.HSCode,
		data.Remark,
		data.InventoryItem,
		data.SalesItem,
		data.PurchaseItem,
		data.ServiceItem,
		data.TaxLiable,
		data.LastUpdateBy,
		data.LastUpdateDate,
		data.ItemCode,
	)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Failed to rollback transaction: %+v", rbErr)
			return nil, fmt.Errorf("error rolling back transaction: %w", rbErr)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return data, customerrors.ErrNoDataEdited
		}
		return nil, fmt.Errorf("error updating item's category: %w", err)
	}

	query = "INSERT INTO tbllogactivity (UserCode, Code, Category, LastUpDt) VALUES (?, ?, ?, ?)"

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.ItemCode, "Item", data.LastUpdateDate)

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
