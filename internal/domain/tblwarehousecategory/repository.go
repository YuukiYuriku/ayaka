package tblwarehousecategory

import (
	"context"

	// "gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblWarehouseCategory interface {
	FetchWarehouseCategories(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) // FetchWarehouseCategory(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	DetailWarehouseCategory(ctx context.Context, code string) (*DetailTblWarehouseCategory, error)
	Create(ctx context.Context, data *CreateTblWarehouseCategory) (*CreateTblWarehouseCategory, error)
	Update(ctx context.Context, data *UpdateTblWarehouseCategory) (*UpdateTblWarehouseCategory, error)
}
