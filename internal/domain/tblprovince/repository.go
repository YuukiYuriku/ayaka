package tblprovince

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

// ProvinceRepository - Interface for accessing province data
type ProvinceRepository interface {
	FetchProvinces(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *CreateTblProvince) (*CreateTblProvince, error)
	DetailProvince(ctx context.Context, provCode string) (*DetailTblProvince, error)
	Update(ctx context.Context, data *UpdateTblProvince, userCode string) error
	GetGroupProvinces(ctx context.Context) ([]*datagroup.DataGroup, error)
}
