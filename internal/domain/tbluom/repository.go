package tbluom

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblUom interface {
	FetchUom(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *CreateTblUom) (*CreateTblUom, error)
	Update(ctx context.Context, data *UpdateTblUom) (*UpdateTblUom, error)
}
