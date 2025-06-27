package service

import (
	"context"
	"fmt"

	// "errors"
	"time"

	"github.com/runsystemid/golog"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblinitialstock"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblInitStockService interface {
	Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, docNo string) (*tblinitialstock.Detail, error)
	Create(ctx context.Context, data *tblinitialstock.Create, userName string) (*tblinitialstock.Create, error)
	Update(ctx context.Context, data *tblinitialstock.Detail, userCode string) (*tblinitialstock.Detail, error)
}

type TblInitStock struct {
	TemplateRepo tblinitialstock.Repository  `inject:"tblInitStockRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblInitStock) Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

func (s *TblInitStock) Detail(ctx context.Context, docNo string) (*tblinitialstock.Detail, error) {
	return s.TemplateRepo.Detail(ctx, docNo)
}

func (s *TblInitStock) Create(ctx context.Context, data *tblinitialstock.Create, userName string) (*tblinitialstock.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "StockInitial")
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

	total, err := s.ID.GetLastDetailNumber(ctx, "StockInitialDtl")
	if err != nil {
		return nil, err
	}

	data.DocType = "Initial Stock"

	for i := 0; i < len(data.Detail); i++ {
		total++
		data.Detail[i].Cancel = booldatatype.FromBool(false)

		data.Detail[i].DNo = fmt.Sprintf("%03d", total)

		if data.Detail[i].Batch == "" {
			data.Detail[i].Batch = data.Date
		}
		data.Detail[i].Source = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Detail[i].DNo)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create initial stock: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblInitStock) Update(ctx context.Context, data *tblinitialstock.Detail, userCode string) (*tblinitialstock.Detail, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Detail); i++ {
		data.Detail[i].Cancel = booldatatype.FromBool(data.Detail[i].Cancel.ToBool())
	}

	prevData, err := s.TemplateRepo.Detail(ctx, data.DocNo)
	if err != nil {
		golog.Error(ctx, "Error get prev initial stock: "+err.Error(), err)
		return nil, err
	}

	for i, detail := range data.Detail {
		for _, detailPrev := range prevData.Detail {
			if (detail.DNo == detailPrev.DNo) && (detail.Cancel == booldatatype.FromBool(false)) && (detailPrev.Cancel == booldatatype.FromBool(true)) {
				return nil, customerrors.ErrInvalidInput
			}
		}
		data.Detail[i].Cancel = booldatatype.FromBool(detail.Cancel.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update initial stock: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
