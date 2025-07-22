package tblmaterialrequest

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *Create) (*Create, error)
	Update(ctx context.Context, lastUpby, lastUpDate string, data *Read) (*Read, error)
	GetMaterialRequest(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	OutstandingMaterial(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}