package tblorderreport

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	ByVendor(ctx context.Context, date string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}