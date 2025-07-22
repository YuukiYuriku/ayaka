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
	"gitlab.com/ayaka/internal/domain/tblvendorquotation"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblVendorQuotationService interface {
	Fetch(ctx context.Context, doc, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblvendorquotation.Create, userName string) (*tblvendorquotation.Create, error)
	Update(ctx context.Context, data *tblvendorquotation.Read, userCode string) (*tblvendorquotation.Read, error)
	GetVendorQuotation(ctx context.Context, itemCode string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblVendorQuotation struct {
	TemplateRepo tblvendorquotation.Repository  `inject:"tblVendorQuotationRepository"`
	ID           *formatid.GenerateIDHandler `inject:"generateID"`
}

func (s *TblVendorQuotation) Fetch(ctx context.Context, doc, vendor, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

	return s.TemplateRepo.Fetch(ctx, doc, vendor, startDate, endDate, param)
}

func (s *TblVendorQuotation) Create(ctx context.Context, data *tblvendorquotation.Create, userName string) (*tblvendorquotation.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "VendorQuotation")
	data.Remark.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create vendor quotation: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create vendor quotation: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")


	for i := 0; i < len(data.Details); i++ {
		data.Details[i].ActiveInd = booldatatype.FromBool(true)

		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create vendor quotation: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblVendorQuotation) Update(ctx context.Context, data *tblvendorquotation.Read, userCode string) (*tblvendorquotation.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i, details := range data.Details {
		if details.UsedInd.ToBool() && !details.ActiveInd.ToBool() {
			return nil, customerrors.ErrInvalidInput
		}
		data.Details[i].ActiveInd = booldatatype.FromBool(details.ActiveInd.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update vendor quotation: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblVendorQuotation) GetVendorQuotation(ctx context.Context, itemCode string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetVendorQuotation(ctx, itemCode, param)
}