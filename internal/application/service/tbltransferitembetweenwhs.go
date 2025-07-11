package service

import (
	"context"
	// "fmt"

	// "errors"
	// "time"

	// "github.com/runsystemid/golog"
	"gitlab.com/ayaka/internal/domain/tbltransferbetweenwhs"
	// "gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblTransferItemBetweenWhsService interface {
	// Fetch(ctx context.Context, warehouse, date, itemCatCode, itemCode, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	GetMaterial(ctx context.Context, itemName, itemCatCode, batch, warehouse string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblTransferItemBetweenWhs struct {
	TemplateRepo tbltransferbetweenwhs.Repository  `inject:"tblTransferItemBetweenWhsRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblTransferItemBetweenWhs) GetMaterial(ctx context.Context, itemName, batch, warehouseFrom, warehouseTo string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetMaterial(ctx, itemName, batch, warehouseFrom, warehouseTo, param)
}