package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblitemcategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblItemCatService interface {
	FetchItemCategories(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblitemcategory.Create, userName string) (*tblitemcategory.Create, error)
	Update(ctx context.Context, data *tblitemcategory.Update, userCode string) (*tblitemcategory.Update, error)
}

type TblItemCat struct {
	TemplateRepo tblitemcategory.RepositoryTblItemCategory `inject:"tblItemCatRepository"`
}

func (s *TblItemCat) FetchItemCategories(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchItemCat(ctx, search, param)
}

func (s *TblItemCat) Create(ctx context.Context, data *tblitemcategory.Create, userName string) (*tblitemcategory.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")
	data.CoaSales.SetNullIfEmpty()
	data.CoaCOGS.SetNullIfEmpty()
	data.CoaConsumptionCost.SetNullIfEmpty()
	data.CoaPurchaseReturn.SetNullIfEmpty()
	data.CoaSalesReturn.SetNullIfEmpty()
	data.CoaStock.SetNullIfEmpty()
	data.Active = booldatatype.FromBool(true)

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create item category: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblItemCat) Update(ctx context.Context, data *tblitemcategory.Update, userCode string) (*tblitemcategory.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	data.CoaSales.SetNullIfEmpty()
	data.CoaCOGS.SetNullIfEmpty()
	data.CoaConsumptionCost.SetNullIfEmpty()
	data.CoaPurchaseReturn.SetNullIfEmpty()
	data.CoaSalesReturn.SetNullIfEmpty()
	data.CoaStock.SetNullIfEmpty()
	data.Active = booldatatype.FromBool(data.Active.ToBool())

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update item category: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}
