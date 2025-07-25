package tbltransferbetweenwhs

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	GetMaterial(ctx context.Context, itemName, batch, warehouseFrom, warehouseTo string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Fetch(ctx context.Context, item, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}