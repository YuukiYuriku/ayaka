package service

import (
	"context"
	// "encoding/json"

	"fmt"

	// "errors"
	"time"

	"github.com/runsystemid/golog"
	// share "gitlab.com/ayaka/internal/domain/shared"

	// "gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblstockmutationhdr"

	// "gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblStockMutationService interface {
	Fetch(ctx context.Context, doc, warehouse, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, docNo string) (*tblstockmutationhdr.Detail, error)
	Create(ctx context.Context, data *tblstockmutationhdr.Create, userName string) (*tblstockmutationhdr.Create, error)
	Update(ctx context.Context, data *tblstockmutationhdr.Detail, userCode string) (*tblstockmutationhdr.Detail, error)
}

type TblStockMutation struct {
	TemplateRepo tblstockmutationhdr.Repository `inject:"tblStockMutationRepository"`
	ID           *formatid.GenerateIDHandler    `inject:"generateID"`
}

func (s *TblStockMutation) Fetch(ctx context.Context, doc, warehouse, batch string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, doc, warehouse, batch, param)
}

func (s *TblStockMutation) Detail(ctx context.Context, docNo string) (*tblstockmutationhdr.Detail, error) {
	return s.TemplateRepo.Detail(ctx, docNo)
}

func (s *TblStockMutation) Create(ctx context.Context, data *tblstockmutationhdr.Create, userName string) (*tblstockmutationhdr.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "StockMutation")
	if err != nil {
		golog.Error(ctx, "Error generate id create initial stock: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create initial stock: " + err.Error())
	}

	data.Remark.SetNullIfEmpty()

	t, err := time.Parse("2006-01-02", data.DocDate)
	if err != nil {
		return nil, err
	}
	data.DocDate = t.Format("20060102")
	data.BatchNo = data.DocDate

	count := 0

	for i := 0; i < len(data.FromArray); i++ {
		count++

		if data.FromArray[i].Stock < data.FromArray[i].Qty {
			return nil, customerrors.ErrInvalidQuantity
		}

		data.FromArray[i].DNo = fmt.Sprintf("%03d", count)

		if data.FromArray[i].BatchNo == "" {
			data.FromArray[i].BatchNo = data.DocDate
		}
		data.FromArray[i].Source = fmt.Sprintf("%s*%s*%s", data.DocDate[4:6], data.DocNo, data.FromArray[i].DNo)
	}

	for i := 0; i < len(data.ToArray); i++ {
		count++

		data.ToArray[i].DNo = fmt.Sprintf("%03d", count)

		data.ToArray[i].Stock = 0

		if data.ToArray[i].BatchNo == "" {
			data.ToArray[i].BatchNo = data.DocDate
		}
		data.ToArray[i].Source = fmt.Sprintf("%s*%s*%s", data.DocDate[6:8], data.DocNo, data.ToArray[i].DNo)
	}

	return s.TemplateRepo.Create(ctx, data)
}

func (s *TblStockMutation) Update(ctx context.Context, data *tblstockmutationhdr.Detail, userCode string) (*tblstockmutationhdr.Detail, error) {
	lastUpDt := time.Now().Format("200601021504")
	data.Cancel = booldatatype.FromBool(data.Cancel.ToBool())

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update initial stock: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}