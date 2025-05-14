package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblwarehousecategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblWarehouseCategoryService interface {
	FetchWarehouseCategory(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	DetailWarehouseCategory(ctx context.Context, whsCtCode string) (*tblwarehousecategory.DetailTblWarehouseCategory, error)
	Create(ctx context.Context, data *tblwarehousecategory.CreateTblWarehouseCategory, userName string) (*tblwarehousecategory.CreateTblWarehouseCategory, error)
	Update(ctx context.Context, data *tblwarehousecategory.UpdateTblWarehouseCategory, userCode string) (*tblwarehousecategory.UpdateTblWarehouseCategory, error)
}

type TblWarehouseCategory struct {
	TemplateRepo tblwarehousecategory.RepositoryTblWarehouseCategory `inject:"tblWarehouseCategoryRepository"`
}

func (s *TblWarehouseCategory) FetchWarehouseCategory(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchWarehouseCategories(ctx, search, param)
}

func (s *TblWarehouseCategory) DetailWarehouseCategory(ctx context.Context, whsCtCode string) (*tblwarehousecategory.DetailTblWarehouseCategory, error) {
	return s.TemplateRepo.DetailWarehouseCategory(ctx, whsCtCode)
}

func (s *TblWarehouseCategory) Create(ctx context.Context, data *tblwarehousecategory.CreateTblWarehouseCategory, userName string) (*tblwarehousecategory.CreateTblWarehouseCategory, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error creating Warehouse Category: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblWarehouseCategory) Update(ctx context.Context, data *tblwarehousecategory.UpdateTblWarehouseCategory, userCode string) (*tblwarehousecategory.UpdateTblWarehouseCategory, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error updating Warehouse Category: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}
