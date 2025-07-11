package tbldirectmaterialreceive

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Create(ctx context.Context, data *Create) (*Create, error)
	Fetch(ctx context.Context, doc, warehouse, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Update(ctx context.Context, lastUpby, lastUpDate string, data *Read) (*Read, error)
}