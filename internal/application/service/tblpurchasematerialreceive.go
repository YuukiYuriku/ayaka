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
	"gitlab.com/ayaka/internal/domain/tblpurchasematerialreceive"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblPurchaseMaterialReceiveService interface {
	Fetch(ctx context.Context, doc, warehouse, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblpurchasematerialreceive.Create, userName string) (*tblpurchasematerialreceive.Create, error)
	Update(ctx context.Context, data *tblpurchasematerialreceive.Read, userCode string) (*tblpurchasematerialreceive.Read, error)
	Reporting(ctx context.Context, doc, warehouse, vendor, item, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblPurchaseMaterialReceive struct {
	TemplateRepo tblpurchasematerialreceive.Repository `inject:"tblPurchaseMaterialReceiveRepository"`
	ID           *formatid.GenerateIDHandler           `inject:"generateID"`
}

func (s *TblPurchaseMaterialReceive) Fetch(ctx context.Context, doc, warehouse, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

func (s *TblPurchaseMaterialReceive) Create(ctx context.Context, data *tblpurchasematerialreceive.Create, userName string) (*tblpurchasematerialreceive.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "PurchaseMaterialReceive")
	data.Remark.SetNullIfEmpty()
	data.SiteCode.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create purchase material receive: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create purchase material receive: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		data.Details[i].CancelInd = booldatatype.FromBool(false)

		if data.Details[i].BatchNo == "" {
			data.Details[i].BatchNo = data.Date
		}
		data.Details[i].Source = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Details[i].DNo)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create purchase material receive: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblPurchaseMaterialReceive) Update(ctx context.Context, data *tblpurchasematerialreceive.Read, userCode string) (*tblpurchasematerialreceive.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].CancelInd = booldatatype.FromBool(data.Details[i].CancelInd.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update purchase material receive: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblPurchaseMaterialReceive) Reporting(ctx context.Context, doc, warehouse, vendor, item, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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
	return s.TemplateRepo.Reporting(ctx, doc, warehouse, vendor, item, startDate, endDate, param)
}