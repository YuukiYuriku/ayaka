package tblwarehouse

import (
	"context"

	// "gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblWarehouse interface {
	FetchWarehouses(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	DetailWarehouse(ctx context.Context, code string) (*DetailTblWarehouse, error)
	Create(ctx context.Context, data *CreateTblWarehouse) (*CreateTblWarehouse, error)
	Update(ctx context.Context, data *UpdateTblWarehouse) (*UpdateTblWarehouse, error)
}
