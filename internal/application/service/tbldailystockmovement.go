package service

import (
	"context"
	"time"

	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tbldailystockmovement"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDailyStockMovementService interface {
	Fetch(ctx context.Context, warehouse, date, itemName, itemCategoryName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblDailyStockMovement struct {
	TemplateRepo tbldailystockmovement.Repository `inject:"tblDailyStockMovementRepository"`
}

func (s *TblDailyStockMovement) Fetch(ctx context.Context, warehouse, date, itemName, itemCategoryName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var err error

	if date != "" {
		date, err = share.FormatToCompactDateTime(date)
		if err != nil {
			return nil, err
		}
	}

	if date == "" {
		date, err = share.FormatToCompactDateTime(time.Now().Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
	}
	return s.TemplateRepo.Fetch(ctx, warehouse, date, itemName,itemCategoryName, param)
}