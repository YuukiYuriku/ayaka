package tblitemcategory

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblItemCategory interface {
	FetchItemCat(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *Create) (*Create, error)
	Update(ctx context.Context, data *Update) (*Update, error)
}
