package tblcoa

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type RepositoryTblCoa interface {
	// Get all TblCoa records
	FetchCoa(ctx context.Context, name string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}
