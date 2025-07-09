package service

import (
	"context"
	//"errors"
	"fmt"
	"time"

	"github.com/runsystemid/golog"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tbldirectsalesdelivery"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDirectSalesDeliveryService interface {
	Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tbldirectsalesdelivery.Create, userName string) (*tbldirectsalesdelivery.Create, error)
	Update(ctx context.Context, data *tbldirectsalesdelivery.Read, userCode string) (*tbldirectsalesdelivery.Read, error)
}

type TblDirectSalesDelivery struct {
	TemplateRepo tbldirectsalesdelivery.Repository `inject:"tblDirectSalesDeliveryRepository"`
	ID           *formatid.GenerateIDHandler       `inject:"generateID"`
}

func (s *TblDirectSalesDelivery) Fetch(ctx context.Context, doc, warehouse, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	var err error
	if startDate != "" {
		startDate, err = share.FormatToCompactDateTime(startDate)
		if err != nil {
			return nil, err
		}
	}
	if endDate != "" {
		endDate, err = share.FormatToCompactDateTime(endDate)
		if err != nil {
			return nil, err
		}
	}
	return s.TemplateRepo.Fetch(ctx, doc, warehouse, startDate, endDate, param)
}

func (s *TblDirectSalesDelivery) Create(ctx context.Context, data *tbldirectsalesdelivery.Create, userName string) (*tbldirectsalesdelivery.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "DirectSalesDelivery")
	data.Remark.SetNullIfEmpty()
	data.Address.SetNullIfEmpty()
	data.CityCode.SetNullIfEmpty()
	data.PostalCode.SetNullIfEmpty()
	data.Phone.SetNullIfEmpty()
	data.Email.SetNullIfEmpty()
	data.Mobile.SetNullIfEmpty()
	data.TaxCode.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create direct sales delivery: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create direct sales delivery: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		data.Details[i].Cancel = booldatatype.FromBool(false)

		if data.Details[i].BatchNo == "" {
			data.Details[i].BatchNo = data.Date
		}
		data.Details[i].Source = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Details[i].DNo)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create direct sales delivery: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblDirectSalesDelivery) Update(ctx context.Context, data *tbldirectsalesdelivery.Read, userCode string) (*tbldirectsalesdelivery.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].Cancel = booldatatype.FromBool(data.Details[i].Cancel.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update direct sales delivery: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
