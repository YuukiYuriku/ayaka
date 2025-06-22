package tblsite

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *Create) (*Create, error)
	Update(ctx context.Context, data *Update) (*Update, error)
}
