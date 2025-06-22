package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tbltaxgroup"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblTaxGroupService interface {
	Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tbltaxgroup.Create, userName string) (*tbltaxgroup.Create, error)
	Update(ctx context.Context, data *tbltaxgroup.Update, userCode string) (*tbltaxgroup.Update, error)
}

type TblTaxGroup struct {
	TemplateRepo tbltaxgroup.RepositoryTblTaxGroup `inject:"tblTaxGroupRepository"`
}

func (s *TblTaxGroup) Fetch(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, search, param)
}

func (s *TblTaxGroup) Create(ctx context.Context, data *tbltaxgroup.Create, userName string) (*tbltaxgroup.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create tax group: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblTaxGroup) Update(ctx context.Context, data *tbltaxgroup.Update, userCode string) (*tbltaxgroup.Update, error) {
	data.LastUpdateBy = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update tax group: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}