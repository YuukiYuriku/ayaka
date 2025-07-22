package tblvendorquotation

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, doc, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *Create) (*Create, error)
	Update(ctx context.Context, lastUpby, lastUpDate string, data *Read) (*Read, error)
	GetVendorQuotation(ctx context.Context, itemCode string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}