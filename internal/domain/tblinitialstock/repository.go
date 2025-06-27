package tblinitialstock

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, docNo string) (*Detail, error)
	Create(ctx context.Context, data *Create) (*Create, error)
	Update(ctx context.Context, lastUpby, lastUpDate string, data *Detail) (*Detail, error)
}
