package service

import (
	"context"
	"encoding/json"

	// "fmt"

	// "errors"
	// "time"

	// "github.com/runsystemid/golog"
	share "gitlab.com/ayaka/internal/domain/shared"
	// "gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblstockmovement"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockMovementService interface {
	Fetch(ctx context.Context, warehouse, dateRangeStart, dateRangeEnd, docType, itemCategory, itemName, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblStockMovement struct {
	TemplateRepo tblstockmovement.Repository `inject:"tblStockMovementRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblStockMovement) Fetch(ctx context.Context, warehouse, dateRangeStart, dateRangeEnd, docType, itemCategory, itemName, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var arrayWarehouse []string
	var err error

	if warehouse != "" {
		err = json.Unmarshal([]byte(warehouse), &arrayWarehouse)
		if err != nil {
			return nil, customerrors.ErrInvalidArrayFormat
		}
	}
	if dateRangeStart != "" {
		dateRangeStart, err = share.FormatToCompactDateTime(dateRangeStart)
		if err != nil {
			return nil, err
		}
	}

	if dateRangeEnd != "" {
		dateRangeEnd, err = share.FormatToCompactDateTime(dateRangeEnd)
		if err != nil {
			return nil, err
		}
	}
	return s.TemplateRepo.Fetch(ctx, arrayWarehouse, dateRangeStart, dateRangeEnd, docType, itemCategory, itemName, batch, param)
}
