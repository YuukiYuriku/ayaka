package tblcity

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/datagroup"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblCity interface {
	//get all cities or get city by name with pagination
	FetchCities(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	//get detail city by city code
	// DetailCity(ctx context.Context, CityCode string) (*DetailTblCity, error)
	// create city
	Create(ctx context.Context, data *CreateTblCity) (*CreateTblCity, error)
	// update a city by city code
	Update(ctx context.Context, data *UpdateTblCity) (*UpdateTblCity, error)
	// get all cities group by province
	GetGroupCities(ctx context.Context) ([]*datagroup.DataGroup, error)
}
