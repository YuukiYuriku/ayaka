package tblhistoryofstock

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, item, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}