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
	"gitlab.com/ayaka/internal/domain/tblstocksummary"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockSummaryService interface {
	Fetch(ctx context.Context, warehouse, date, itemCatCode, itemCode, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	GetItem(ctx context.Context, itemName, itemCatCode, batch, warehouse string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblStockSummary struct {
	TemplateRepo tblstocksummary.Repository  `inject:"tblStockSummaryRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblStockSummary) Fetch(ctx context.Context, warehouse, date, itemCatCode, itemCode, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var arrayWarehouse []string
	var err error

	if warehouse != "" {
		err = json.Unmarshal([]byte(warehouse), &arrayWarehouse)
		if err != nil {
			return nil, customerrors.ErrInvalidArrayFormat
		}
	}
	if date != "" {
		date, err = share.FormatToCompactDateTime(date)
		if err != nil {
			return nil, err
		}
	}
	return s.TemplateRepo.Fetch(ctx, arrayWarehouse, date, itemCatCode, itemCode, itemName, param)
}

func (s *TblStockSummary) GetItem(ctx context.Context, itemName, itemCatCode, batch, warehouse string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetItem(ctx, itemName, itemCatCode, batch, warehouse, param)
}
