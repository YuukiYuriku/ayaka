package service

import (
	"context"
	//"errors"
	"fmt"
	"time"

	"github.com/runsystemid/golog"
	// share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblrecvvdhdr"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDirectPurchaseReceiveService interface {
	Fetch(ctx context.Context, doc string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Detail(ctx context.Context, docNo string) (*tblrecvvdhdr.Detail, error)
	Create(ctx context.Context, data *tblrecvvdhdr.Create, userName string) (*tblrecvvdhdr.Create, error)
	Update(ctx context.Context, data *tblrecvvdhdr.Detail, userCode string) (*tblrecvvdhdr.Detail, error)
}

type TblDirectPurchaseReceive struct {
	TemplateRepo tblrecvvdhdr.Repository     `inject:"tblDirectPurchaseReceiveRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblDirectPurchaseReceive) Fetch(ctx context.Context, doc string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.Fetch(ctx, doc, param)
}

func (s *TblDirectPurchaseReceive) Detail(ctx context.Context, docNo string) (*tblrecvvdhdr.Detail, error) {
	return s.TemplateRepo.Detail(ctx, docNo)
}

func (s *TblDirectPurchaseReceive) Create(ctx context.Context, data *tblrecvvdhdr.Create, userName string) (*tblrecvvdhdr.Create, error) {
	data.CreateBy = userName
	data.CreateDate = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "DirectPurchaseReceive")
	data.Remark.SetNullIfEmpty()
	data.LocalCode.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create direct purchase receive: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create direct purchase receive: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	total, err := s.ID.GetLastDetailNumber(ctx, "DirectPurchaseReceiveDtl")
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(data.Detail); i++ {
		total++
		data.Detail[i].Cancel = booldatatype.FromBool(false)

		data.Detail[i].DNo = fmt.Sprintf("%03d", total)

		if data.Detail[i].Batch == "" {
			data.Detail[i].Batch = data.Date
		}
		data.Detail[i].LocalCode = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Detail[i].DNo)

		if data.Detail[i].Expired != "" {
			tFormat, err := time.Parse("2006-01-02", data.Detail[i].Expired)
			if err != nil {
				return nil, err
			}
			data.Detail[i].Expired = tFormat.Format("20060102")
		}
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create direct purchase receive: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblDirectPurchaseReceive) Update(ctx context.Context, data *tblrecvvdhdr.Detail, userCode string) (*tblrecvvdhdr.Detail, error) {
	oldData, err := s.TemplateRepo.Detail(ctx, data.DocNo)
	if err != nil {
		return nil, fmt.Errorf("error get old data: %w", err)
	}

	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Detail); i++ {
		data.Detail[i].Cancel = booldatatype.FromBool(data.Detail[i].Cancel.ToBool())
		data.Detail[i].CancelReason.SetNullIfEmpty()
	}

	res, err := s.TemplateRepo.Update(ctx, data, oldData, userCode, lastUpDt)
	if err != nil {
		golog.Error(ctx, "Error update initial stock: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
