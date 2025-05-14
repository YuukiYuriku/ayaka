package service

import (
	"context"

	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblmasteritem"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblItemService interface {
	FetchItems(ctx context.Context, search, category string, status bool, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, itemCode string) (*tblmasteritem.Detail, error)
	Create(ctx context.Context, data *tblmasteritem.Create, userName string, confirm bool) (*tblmasteritem.Create, error)
	Update(ctx context.Context, data *tblmasteritem.Update, userCode string, confirm bool) (*tblmasteritem.Update, error)
}

type TblItem struct {
	TemplateRepo tblmasteritem.Repository `inject:"tblItemRepository"`
}

func (s *TblItem) FetchItems(ctx context.Context, search, category string, status bool, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var act string
	if status {
		act = "Y"
	} else {
		act = ""
	}
	return s.TemplateRepo.FetchItems(ctx, search, category, act, param)
}

func (s *TblItem) Detail(ctx context.Context, itemCode string) (*tblmasteritem.Detail, error) {
	return s.TemplateRepo.Detail(ctx, itemCode)
}

func (s *TblItem) Create(ctx context.Context, data *tblmasteritem.Create, userName string, confirm bool) (*tblmasteritem.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")
	data.LocalCode.SetNullIfEmpty()
	data.ForeignName.SetNullIfEmpty()
	data.OldCode.SetNullIfEmpty()
	data.Spesification.SetNullIfEmpty()
	data.HSCode.SetNullIfEmpty()
	data.Remark.SetNullIfEmpty()
	data.Active = booldatatype.FromBool(data.Active.ToBool())
	data.InventoryItem = booldatatype.FromBool(data.InventoryItem.ToBool())
	data.SalesItem = booldatatype.FromBool(data.SalesItem.ToBool())
	data.PurchaseItem = booldatatype.FromBool(data.PurchaseItem.ToBool())
	data.SalesItem = booldatatype.FromBool(data.SalesItem.ToBool())
	data.TaxLiable = booldatatype.FromBool(data.TaxLiable.ToBool())

	res, err := s.TemplateRepo.Create(ctx, data, confirm)
	if err != nil {
		golog.Error(ctx, "Error create item: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblItem) Update(ctx context.Context, data *tblmasteritem.Update, userCode string, confirm bool) (*tblmasteritem.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	data.LocalCode.SetNullIfEmpty()
	data.ForeignName.SetNullIfEmpty()
	data.OldCode.SetNullIfEmpty()
	data.Spesification.SetNullIfEmpty()
	data.HSCode.SetNullIfEmpty()
	data.Remark.SetNullIfEmpty()
	data.Active = booldatatype.FromBool(data.Active.ToBool())
	data.InventoryItem = booldatatype.FromBool(data.InventoryItem.ToBool())
	data.SalesItem = booldatatype.FromBool(data.SalesItem.ToBool())
	data.PurchaseItem = booldatatype.FromBool(data.PurchaseItem.ToBool())
	data.SalesItem = booldatatype.FromBool(data.SalesItem.ToBool())
	data.TaxLiable = booldatatype.FromBool(data.TaxLiable.ToBool())

	res, err := s.TemplateRepo.Update(ctx, data, confirm)
	if err != nil {
		golog.Error(ctx, "Error update item: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}
