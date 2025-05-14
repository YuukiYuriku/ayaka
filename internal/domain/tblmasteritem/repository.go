package tblmasteritem

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	FetchItems(ctx context.Context, name, category, status string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, search string) (*Detail, error)
	Create(ctx context.Context, data *Create, confirm bool) (*Create, error)
	Update(ctx context.Context, data *Update, confirm bool) (*Update, error)
}
