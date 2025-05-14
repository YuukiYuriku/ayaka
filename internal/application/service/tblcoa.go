package service

import (
	"context"

	"gitlab.com/ayaka/internal/domain/tblcoa"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblCoaService interface {
	FetchCoa(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblCoa struct {
	TemplateRepo tblcoa.RepositoryTblCoa `inject:"tblCoaRepository"` //tblCountryCacheRepository
}

func (s *TblCoa) FetchCoa(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.FetchCoa(ctx, search, param)
}
