package service

import (
	"context"
	"fmt"
	"time"

	// "fmt"

	// "errors"
	// "time"

	// "github.com/runsystemid/golog"
	"github.com/runsystemid/golog"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblstockadjustmenthdr"

	// "gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/pkg/pagination"
	// "gitlab.com/ayaka/internal/pkg/customerrors"
)

type TblStockAdjustService interface {
	Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, docNo string) (*tblstockadjustmenthdr.Detail, error)
	Create(ctx context.Context, data *tblstockadjustmenthdr.Create, userName string) (*tblstockadjustmenthdr.Create, error)
}

type TblStockAdjust struct {
	TemplateRepo tblstockadjustmenthdr.Repository `inject:"tblStockAdjustRepository"`
	ID           *formatid.GenerateIDHandler      `inject:"generateID"`
}

func (s *TblStockAdjust) Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	if startDate != "" && endDate != "" {
		var err error
		startDate, err = share.FormatToCompactDateTime(startDate)
		if err != nil {
			return nil, err
		}

		endDate, err = share.FormatToCompactDateTime(endDate)
		if err != nil {
			return nil, err
		}
	}

	return s.TemplateRepo.Fetch(ctx, doc, warehouse, startDate, endDate, param)
}

func (s *TblStockAdjust) Detail(ctx context.Context, docNo string) (*tblstockadjustmenthdr.Detail, error) {
	return s.TemplateRepo.Detail(ctx, docNo)
}

func (s *TblStockAdjust) Create(ctx context.Context, data *tblstockadjustmenthdr.Create, userName string) (*tblstockadjustmenthdr.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "StockAdjustment")
	data.Remark.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create initial stock: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create initial stock: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	total, err := s.ID.GetLastDetailNumber(ctx, "StockAdjustmentDtl")
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(data.Details); i++ {
		total++

		data.Details[i].DNo = fmt.Sprintf("%03d", total)

		if data.Details[i].Batch == "" {
			data.Details[i].Batch = data.Date
		}
		data.Details[i].Source = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Details[i].DNo)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create stock adjustment: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
