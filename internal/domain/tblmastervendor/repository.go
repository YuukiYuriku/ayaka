package tblmastervendor

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Create(ctx context.Context, data *Create) (*Create, error)
	Fetch(ctx context.Context, name, cat string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, vendorCode string) (*Detail, error)
	Update(ctx context.Context, data *Update) (*Update, error)
	GetContact(ctx context.Context, code string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}