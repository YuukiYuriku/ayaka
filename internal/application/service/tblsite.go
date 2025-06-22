package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/tblsite"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblSiteService interface {
	Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblsite.Create, userName string) (*tblsite.Create, error)
	Update(ctx context.Context, data *tblsite.Update, userCode string) (*tblsite.Update, error)
}

type TblSite struct {
	TemplateRepo tblsite.Repository `inject:"tblSiteRepository"`
}

func (s *TblSite) Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, search, param)
}

func (s *TblSite) Create(ctx context.Context, data *tblsite.Create, userName string) (*tblsite.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	// bool
	data.Active = booldatatype.FromBool(data.Active.ToBool())

	// nullable
	data.Remark.SetNullIfEmpty()

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create site: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblSite) Update(ctx context.Context, data *tblsite.Update, userCode string) (*tblsite.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	// bool
	data.Active = booldatatype.FromBool(data.Active.ToBool())

	// nullable
	data.Remark.SetNullIfEmpty()

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update site: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}