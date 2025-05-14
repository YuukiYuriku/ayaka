package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"gitlab.com/ayaka/internal/adapter/repository"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblitemcategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblItemCatRepository struct {
	DB *repository.Sqlx `inject:"database"`
}

func (t *TblItemCatRepository) FetchItemCat(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var totalRecords int

	search := "%" + name + "%"

	countQuery := "SELECT COUNT(*) FROM tblitemcategory WHERE ItCtCode LIKE ? OR ItCtName LIKE ?"
	if err := t.DB.GetContext(ctx, &totalRecords, countQuery, search, search); err != nil {
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

	var itemCategories []*tblitemcategory.ReadTblItemCategory
	query := `SELECT ItCtCode,
				i.ItCtName,
				i.ActInd,
				i.CreateDt,
				i.AcNo,
				c1.AcDesc AS AcDesc,
				i.AcNo2,
				c2.AcDesc AS AcDesc2,
				i.AcNo3,
				c3.AcDesc AS AcDesc3,
				i.AcNo4,
				c4.AcDesc AS AcDesc4,
				i.AcNo5,
				c5.AcDesc AS AcDesc5,
				i.AcNo6,
				c6.AcDesc AS AcDesc6
				FROM tblitemcategory i
				LEFT JOIN tblcoa c1 ON i.AcNo = c1.AcNo
				LEFT JOIN tblcoa c2 ON i.AcNo2= c2.AcNo
				LEFT JOIN tblcoa c3 ON i.AcNo3= c3.AcNo
				LEFT JOIN tblcoa c4 ON i.AcNo4= c4.AcNo
				LEFT JOIN tblcoa c5 ON i.AcNo5= c5.AcNo
				LEFT JOIN tblcoa c6 ON i.AcNo6= c6.AcNo
				WHERE ItCtCode LIKE ? OR ItCtName LIKE ?
				LIMIT ? OFFSET ?`

	if err := t.DB.SelectContext(ctx, &itemCategories, query, search, search, param.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &pagination.PaginationResponse{
				Data:         make([]*tblitemcategory.ReadTblItemCategory, 0),
				TotalRecords: 0,
				TotalPages:   0,
				CurrentPage:  param.Page,
				PageSize:     param.PageSize,
				HasNext:      false,
				HasPrevious:  false,
			}, nil
		}
		return nil, fmt.Errorf("error Fetch Item Categories: %w", err)
	}

	// proses data
	j := offset
	result := make([]*tblitemcategory.ReadTblItemCategory, len(itemCategories))
	for i, itemCategory := range itemCategories {
		j++
		result[i] = &tblitemcategory.ReadTblItemCategory{
			Number:                 uint(j),
			ItemCategoryCode:       itemCategory.ItemCategoryCode,
			ItemCategoryName:       itemCategory.ItemCategoryName,
			Active:                 itemCategory.Active,
			CreateDate:             share.FormatDate(itemCategory.CreateDate),
			CoaStock:               itemCategory.CoaStock,
			CoaStockDesc:           itemCategory.CoaStockDesc,
			CoaSales:               itemCategory.CoaSales,
			CoaSalesDesc:           itemCategory.CoaSalesDesc,
			CoaCOGS:                itemCategory.CoaCOGS,
			CoaCOGSDesc:            itemCategory.CoaCOGSDesc,
			CoaSalesReturn:         itemCategory.CoaSalesReturn,
			CoaSalesReturnDesc:     itemCategory.CoaSalesReturnDesc,
			CoaPurchaseReturn:      itemCategory.CoaPurchaseReturn,
			CoaPurchaseReturnDesc:  itemCategory.CoaPurchaseReturnDesc,
			CoaConsumptionCost:     itemCategory.CoaConsumptionCost,
			CoaConsumptionCostDesc: itemCategory.CoaConsumptionCostDesc,
		}
	}

	// response
	response := &pagination.PaginationResponse{
		Data:         result,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  param.Page,
		PageSize:     param.PageSize,
		HasNext:      param.Page < totalPages,
		HasPrevious:  param.Page > 1,
	}

	return response, nil
}

func (t *TblItemCatRepository) Create(ctx context.Context, data *tblitemcategory.Create) (*tblitemcategory.Create, error) {
	query := `INSERT INTO tblitemcategory
				(
					ItCtCode,
					ItCtName,
					ActInd,
					AcNo,
					AcNo2,
					AcNo3,
					AcNo4,
					AcNo5,
					AcNo6,
					CreateDt,
					CreateBy
				)
				VALUES
				(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := t.DB.ExecContext(ctx, query,
		data.ItemCategoryCode,
		data.ItemCategoryName,
		data.Active,
		data.CoaStock,
		data.CoaSales,
		data.CoaCOGS,
		data.CoaSalesReturn,
		data.CoaPurchaseReturn,
		data.CoaConsumptionCost,
		data.CreateDate,
		data.CreateBy,
	)

	if err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, fmt.Errorf("error Create Item Category: %w", err)
	}

	return data, nil
}

func (t *TblItemCatRepository) Update(ctx context.Context, data *tblitemcategory.Update) (*tblitemcategory.Update, error) {
	query := `SELECT ItCtCode,
					ItCtName,
					ActInd,
					AcNo,
					AcNo2,
					AcNo3,
					AcNo4,
					AcNo5,
					AcNo6
				FROM tblitemcategory WHERE ItCtCode = ?`

	var check tblitemcategory.ReadTblItemCategory

	if err := t.DB.GetContext(ctx, &check, query, data.ItemCategoryCode); err != nil {
		log.Printf("Detailed error: %+v", err)
		return nil, err
	}

	query = `UPDATE tblitemcategory SET
			ItCtName = ?,
			ActInd = ?,
			AcNo = ?,
			AcNo2 = ?,
			AcNo3 = ?,
			AcNo4 = ?,
			AcNo5 = ?,
			AcNo6 = ?,
			LastUpDt = ?,
			LastUpBy = ?
			WHERE ItCtCode = ?`

	tx, err := t.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %+v", err)
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, query,
		data.ItemCategoryName,
		data.Active,
		data.CoaStock,
		data.CoaSales,
		data.CoaCOGS,
		data.CoaSalesReturn,
		data.CoaPurchaseReturn,
		data.CoaConsumptionCost,
		data.LastUpdateDate,
		data.LastUpdateBy,
		data.ItemCategoryCode)

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

	_, err = tx.ExecContext(ctx, query, data.LastUpdateBy, data.ItemCategoryCode, "ItemCategory", data.LastUpdateDate)

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
