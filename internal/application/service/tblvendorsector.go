package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblvendorsector"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorSectorService interface {
	Fetch(ctx context.Context, search string, status bool, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblvendorsector.Create, userName string) (*tblvendorsector.Create, error)
	Update(ctx context.Context, data *tblvendorsector.Update, userCode string) (*tblvendorsector.Update, error)
	GetSector(ctx context.Context) ([]tblvendorsector.GetSector, error)
	GetSubSector(ctx context.Context, code string) ([]tblvendorsector.GetSubSector, error)
}

type TblVendorSector struct {
	TemplateRepo tblvendorsector.Repository `inject:"tblVendorSectorRepository"`
}

func (s *TblVendorSector) Fetch(ctx context.Context, search string, status bool, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	active := ""
	if status {
		active = "Y"
	}

	return s.TemplateRepo.Fetch(ctx, search, active, param)
}

func (s *TblVendorSector) Create(ctx context.Context, data *tblvendorsector.Create, userName string) (*tblvendorsector.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")
	data.Active = booldatatype.FromBool(true)

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create vendor sector: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblVendorSector) Update(ctx context.Context, data *tblvendorsector.Update, userCode string) (*tblvendorsector.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")
	data.Active = booldatatype.FromBool(true)

	if len(data.Details) > 0 {
		for i := range data.Details {
			data.Details[i].Active = booldatatype.FromBool(true)
		}
	}

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update vendor sector: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	if len(res.Details) != 0 {
		var filtered []tblvendorsector.VendorSectorDtl
		for _, detail := range res.Details {
			if detail.Active.ToBool() {
				filtered = append(filtered, detail)
			}
		}
		res.Details = filtered
	}

	return res, nil
}

func (s *TblVendorSector) GetSector(ctx context.Context) ([]tblvendorsector.GetSector, error) {
	return s.TemplateRepo.GetSector(ctx)
}

func (s *TblVendorSector) GetSubSector(ctx context.Context, code string) ([]tblvendorsector.GetSubSector, error) {
	return s.TemplateRepo.GetSubSector(ctx, code)
}