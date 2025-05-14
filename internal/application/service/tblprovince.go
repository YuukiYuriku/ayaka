package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tblprovince"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblProvinceService interface {
	FetchProvinces(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	DetailProvince(ctx context.Context, provCode string) (*tblprovince.DetailTblProvince, error)
	Create(ctx context.Context, data *tblprovince.CreateTblProvince, userName string) (*tblprovince.CreateTblProvince, error)
	Update(ctx context.Context, data *tblprovince.UpdateTblProvince, userCode string) (*tblprovince.UpdateTblProvince, error)
	GetGroupProvinces(ctx context.Context) ([]*datagroup.DataGroup, error)
}

type TblProvince struct {
	TemplateRepo tblprovince.ProvinceRepository `inject:"tblProvinceRepository"`
}

func (s *TblProvince) FetchProvinces(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchProvinces(ctx, search, param)
}

func (s *TblProvince) DetailProvince(ctx context.Context, provCode string) (*tblprovince.DetailTblProvince, error) {
	return s.TemplateRepo.DetailProvince(ctx, provCode)
}

func (s *TblProvince) Create(ctx context.Context, data *tblprovince.CreateTblProvince, userName string) (*tblprovince.CreateTblProvince, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create Province: "+err.Error(), err)
		return data, err
	}

	return res, nil
}

func (s *TblProvince) Update(ctx context.Context, data *tblprovince.UpdateTblProvince, userCode string) (*tblprovince.UpdateTblProvince, error) {
	data.UserCode = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	err := s.TemplateRepo.Update(ctx, data, userCode)
	if err != nil {
		golog.Error(ctx, "Error update Province: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return nil, err
		}
		return data, nil
	}

	return data, nil
}

func (s *TblProvince) GetGroupProvinces(ctx context.Context) ([]*datagroup.DataGroup, error) {
	return s.TemplateRepo.GetGroupProvinces(ctx)
}
