package tblcountry

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblCountry interface {
	//to get all countries or get countries by name with pagination
	FetchCountries(ctx context.Context, search string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	//create new country
	Create(ctx context.Context, data *Createtblcountry) (*Createtblcountry, error)
	//update country
	Update(ctx context.Context, data *Updatetblcountry) (*Updatetblcountry, error)
}
