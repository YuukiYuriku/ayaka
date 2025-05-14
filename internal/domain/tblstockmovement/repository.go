package tblstockmovement

import (
	"context"

	"gitlab.com/ayaka/internal/pkg/pagination"
)

type Repository interface {
	Fetch(
		ctx context.Context,
		warehouse []string,
		dateRangeStart, dateRangeEnd string,
		docType, itemCategory, itemName, batch string,
		param *pagination.PaginationParam,
	) (*pagination.PaginationResponse, error)
}
