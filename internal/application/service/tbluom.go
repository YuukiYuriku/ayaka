package service

import (
	"context"
	"errors"
	"time"

	"github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tbluom"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblUomService interface {
	FetchUom(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tbluom.CreateTblUom, userName string) (*tbluom.CreateTblUom, error)
	Update(ctx context.Context, data *tbluom.UpdateTblUom, userCode string) (*tbluom.UpdateTblUom, error)
}

type TblUom struct {
	TemplateRepo tbluom.RepositoryTblUom `inject:"tblUomRepository"`
}

func (s *TblUom) FetchUom(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchUom(ctx, search, param)
}

func (s *TblUom) Create(ctx context.Context, data *tbluom.CreateTblUom, userName string) (*tbluom.CreateTblUom, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create uom: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblUom) Update(ctx context.Context, data *tbluom.UpdateTblUom, userCode string) (*tbluom.UpdateTblUom, error) {
	data.UserCode = userCode
	data.LastUpdateDate = time.Now().Format("200601021504")

	res, err := s.TemplateRepo.Update(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error update uom: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited) {
			return data, err
		}
		return nil, err
	}

	return res, nil
}
