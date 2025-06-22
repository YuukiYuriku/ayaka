package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tbltax"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblTaxService interface {
	Fetch(ctx context.Context, search, category string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tbltax.Create, userName string) (*tbltax.Create, error)
	Update(ctx context.Context, data *tbltax.Update, userCode string) (*tbltax.Update, error)
}

type TblTax struct {
	TemplateRepo tbltax.Repository `inject:"tblTaxRepository"`
}

func (s *TblTax) Fetch(ctx context.Context, search, category string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, search, category, param)
}

func (s *TblTax) Create(ctx context.Context, data *tbltax.Create, userName string) (*tbltax.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create Tax: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblTax) Update(ctx context.Context, data *tbltax.Update, userCode string) (*tbltax.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update Tax: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}