package tbldailystockmovement

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, warehouse, date, itemName, itemCategoryName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}