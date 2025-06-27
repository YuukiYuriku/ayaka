package service

import (
	"context"

	"gitlab.com/ayaka/internal/domain/tblhistoryofstock"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblHistoryOfStockService interface {
	Fetch(ctx context.Context, item, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblHistoryOfStock struct {
	TemplateRepo tblhistoryofstock.Repository `inject:"tblHistoryOfStockRepository"`
}

func (s *TblHistoryOfStock) Fetch(ctx context.Context, item, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, item, batch, param)
}