package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblcountry"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCountryService interface {
	FetchCountries(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblcountry.Createtblcountry, userName string) (*tblcountry.Createtblcountry, error)
	Update(ctx context.Context, data *tblcountry.Updatetblcountry, userCode string) (*tblcountry.Updatetblcountry, error)
}

type TblCountry struct {
	TemplateRepo tblcountry.RepositoryTblCountry `inject:"tblCountryRepository"` //tblCountryCacheRepository
}

func (s *TblCountry) FetchCountries(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchCountries(ctx, search, param)
}

func (s *TblCountry) Create(ctx context.Context, data *tblcountry.Createtblcountry, userName string) (*tblcountry.Createtblcountry, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create country: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblCountry) Update(ctx context.Context, data *tblcountry.Updatetblcountry, userCode string) (*tblcountry.Updatetblcountry, error) {
	data.UserCode = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update country: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}
