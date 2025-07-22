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
	"gitlab.com/ayaka/internal/domain/tblpurchasereturndelivery"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblPurchaseReturnDeliveryService interface {
	Fetch(ctx context.Context, doc, warehouse, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblpurchasereturndelivery.Create, userName string) (*tblpurchasereturndelivery.Create, error)
	Update(ctx context.Context, data *tblpurchasereturndelivery.Read, userCode string) (*tblpurchasereturndelivery.Read, error)
	GetReturnMaterial(ctx context.Context, doc, warehouse, vendor, item string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblPurchaseReturnDelivery struct {
	TemplateRepo tblpurchasereturndelivery.Repository `inject:"tblPurchaseReturnDeliveryRepository"`
	ID           *formatid.GenerateIDHandler          `inject:"generateID"`
}

func (s *TblPurchaseReturnDelivery) Fetch(ctx context.Context, doc, warehouse, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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
	return s.TemplateRepo.Fetch(ctx, doc, warehouse, vendor, startDate, endDate, param)
}

func (s *TblPurchaseReturnDelivery) Create(ctx context.Context, data *tblpurchasereturndelivery.Create, userName string) (*tblpurchasereturndelivery.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "PurchaseReturnDelivery")
	data.Remark.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create purchase return delivery: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create purchase return delivery: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		data.Details[i].CancelInd = booldatatype.FromBool(false)
		data.Details[i].Remark.SetNullIfEmpty()

		if data.Details[i].BatchNo == "" {
			data.Details[i].BatchNo = data.Date
		}
		data.Details[i].Source = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Details[i].DNo)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create purchase return delivery: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblPurchaseReturnDelivery) Update(ctx context.Context, data *tblpurchasereturndelivery.Read, userCode string) (*tblpurchasereturndelivery.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].CancelInd = booldatatype.FromBool(data.Details[i].CancelInd.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update purchase return delivery: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblPurchaseReturnDelivery) GetReturnMaterial(ctx context.Context, doc, warehouse, vendor, item string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetReturnMaterial(ctx, doc, warehouse, vendor, item, param)
}