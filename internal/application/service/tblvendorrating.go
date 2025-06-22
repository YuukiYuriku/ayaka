package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblvendorrating"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorRatingService interface {
	Fetch(ctx context.Context, search string, status bool, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblvendorrating.Create, userName string) (*tblvendorrating.Create, error)
	Update(ctx context.Context, data *tblvendorrating.Update, userCode string) (*tblvendorrating.Update, error)
}

type TblVendorRating struct {
	TemplateRepo tblvendorrating.Repository `inject:"tblVendorRatingRepository"`
}

func (s *TblVendorRating) Fetch(ctx context.Context, search string, status bool, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	active := ""
	if status {
		active = "Y"
	}

	return s.TemplateRepo.Fetch(ctx, search, active, param)
}

func (s *TblVendorRating) Create(ctx context.Context, data *tblvendorrating.Create, userName string) (*tblvendorrating.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")
	data.Active = booldatatype.FromBool(data.Active.ToBool())

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create vendor rating: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblVendorRating) Update(ctx context.Context, data *tblvendorrating.Update, userCode string) (*tblvendorrating.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")
	data.Active = booldatatype.FromBool(data.Active.ToBool())

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update vendor rating: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}