package tblstockmutationhdr

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, doc, warehouse, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, docNo string) (*Detail, error)
	Create(ctx context.Context, data *Create) (*Create, error)
}
