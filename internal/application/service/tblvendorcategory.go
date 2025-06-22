package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblvendorcategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorCategoryService interface {
	Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblvendorcategory.Create, userName string) (*tblvendorcategory.Create, error)
	Update(ctx context.Context, data *tblvendorcategory.Update, userCode string) (*tblvendorcategory.Update, error)
}

type TblVendorCategory struct {
	TemplateRepo tblvendorcategory.Repository `inject:"tblVendorCategoryRepository"`
}

func (s *TblVendorCategory) Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, search, param)
}

func (s *TblVendorCategory) Create(ctx context.Context, data *tblvendorcategory.Create, userName string) (*tblvendorcategory.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create vendor category: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblVendorCategory) Update(ctx context.Context, data *tblvendorcategory.Update, userCode string) (*tblvendorcategory.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update vendor category: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}