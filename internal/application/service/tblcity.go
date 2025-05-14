package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblcity"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCityService interface {
	FetchCities(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblcity.CreateTblCity, userName string) (*tblcity.CreateTblCity, error) // Update(ctx context.Context, data *template.UpdatetblCity, userCode string) (*template.UpdatetblCity, error)
	Update(ctx context.Context, data *tblcity.UpdateTblCity, userCode string) (*tblcity.UpdateTblCity, error)
	GetGroupCities(ctx context.Context) ([]*datagroup.DataGroup, error)
}

type TblCity struct {
	TemplateRepo tblcity.RepositoryTblCity `inject:"tblCityRepository"` //tblCityCacheRepository
}

func (s *TblCity) FetchCities(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchCities(ctx, search, param)
}

func (s *TblCity) Create(ctx context.Context, data *tblcity.CreateTblCity, userName string) (*tblcity.CreateTblCity, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")
	data.RingArea.SetNullIfEmpty()
	data.Location.SetNullIfEmpty()

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create City: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblCity) Update(ctx context.Context, data *tblcity.UpdateTblCity, userCode string) (*tblcity.UpdateTblCity, error) {
	data.UserCode = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")
	data.RingArea.SetNullIfEmpty()
	data.Location.SetNullIfEmpty()

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update City: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}

func (s *TblCity) GetGroupCities(ctx context.Context) ([]*datagroup.DataGroup, error) {
	return s.TemplateRepo.GetGroupCities(ctx)
}
