package service

import (
	"context"
	"errors"
	"fmt"

	// "errors"
	"time"

	"github.com/runsystemid/golog"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/booldatatype"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblpurchaseorderrequest"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblPurchaseOrderRequestService interface {
	Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblpurchaseorderrequest.Create, userName string) (*tblpurchaseorderrequest.Create, error)
	Update(ctx context.Context, data *tblpurchaseorderrequest.Read, userCode string) (*tblpurchaseorderrequest.Read, error)
	GetPurchaseOrderRequest(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblPurchaseOrderRequest struct {
	TemplateRepo tblpurchaseorderrequest.Repository `inject:"tblPurchaseOrderRequestRepository"`
	ID           *formatid.GenerateIDHandler        `inject:"generateID"`
}

func (s *TblPurchaseOrderRequest) Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

	return s.TemplateRepo.Fetch(ctx, doc, startDate, endDate, param)
}

func (s *TblPurchaseOrderRequest) Create(ctx context.Context, data *tblpurchaseorderrequest.Create, userName string) (*tblpurchaseorderrequest.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "PurchaseOrderRequest")
	data.Remark.SetNullIfEmpty()

	if err != nil {
		golog.Error(ctx, "Error generate id create purchase order request: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create purchase order request: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		data.Details[i].CancelInd = booldatatype.FromBool(false)
		data.Details[i].SuccessInd = booldatatype.FromBool(false)
		data.Details[i].Remark.SetNullIfEmpty()
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create purchase order request: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblPurchaseOrderRequest) Update(ctx context.Context, data *tblpurchaseorderrequest.Read, userCode string) (*tblpurchaseorderrequest.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].CancelInd = booldatatype.FromBool(data.Details[i].CancelInd.ToBool())
		if data.Details[i].SuccessInd.ToBool() && data.Details[i].CancelInd.ToBool() {
			return nil, customerrors.ErrInvalidInput
		}
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update purchase order request: "+err.Error(), err)
		if errors.Is(err, customerrors.ErrNoDataEdited){
			return res, nil
		}
		return nil, err
	}

	return res, nil
}

func (s *TblPurchaseOrderRequest) GetPurchaseOrderRequest(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetVendorQuotation(ctx, doc, itemName, param)
}