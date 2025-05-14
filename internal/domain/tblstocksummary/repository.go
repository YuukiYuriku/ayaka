package tblstocksummary

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, warehouse []string, date, itemCatCode, itemCode, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	GetItem(ctx context.Context, itemName, itemCatCode, batch, warehouse string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}
