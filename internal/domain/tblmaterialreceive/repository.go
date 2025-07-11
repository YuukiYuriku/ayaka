package tblmaterialreceive

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *Create) (*Create, error)
}