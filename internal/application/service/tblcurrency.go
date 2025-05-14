package service

import (
	"context"
	// "errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblcurrency"
	"gitlab.com/ayaka/internal/pkg/pagination"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
)

type TblCurrencyService interface {
	Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblcurrency.Create, userName string) (*tblcurrency.Create, error)
	Update(ctx context.Context, data *tblcurrency.Update, userCode string) (*tblcurrency.Update, error)
}

type TblCurrency struct {
	TemplateRepo tblcurrency.Repository `inject:"tblCurrencyRepository"` //tblCurrencyCacheRepository
}

func (s *TblCurrency) Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, search, param)
}

func (s *TblCurrency) Create(ctx context.Context, data *tblcurrency.Create, userName string) (*tblcurrency.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create uom: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblCurrency) Update(ctx context.Context, data *tblcurrency.Update, userCode string) (*tblcurrency.Update, error) {
	data.UserCode = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update currency: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
