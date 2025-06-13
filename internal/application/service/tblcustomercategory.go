package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblcustomercategory"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCustomerCategoryService interface {
	Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblcustomercategory.Create, userName string) (*tblcustomercategory.Create, error)
	Update(ctx context.Context, data *tblcustomercategory.Update, userCode string) (*tblcustomercategory.Update, error)
}

type TblCustomerCategory struct {
	TemplateRepo tblcustomercategory.Repository `inject:"tblCustomerCategoryRepository"`
}

func (s *TblCustomerCategory) Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, search, param)
}

func (s *TblCustomerCategory) Create(ctx context.Context, data *tblcustomercategory.Create, userName string) (*tblcustomercategory.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create customer category: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblCustomerCategory) Update(ctx context.Context, data *tblcustomercategory.Update, userCode string) (*tblcustomercategory.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update customer category: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}