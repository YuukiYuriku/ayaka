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
	"gitlab.com/ayaka/internal/domain/tblmaterialtransfer"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblMaterialTransferService interface {
	Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblmaterialtransfer.Create, userName string) (*tblmaterialtransfer.Create, error)
	Update(ctx context.Context, data *tblmaterialtransfer.Read, userCode string) (*tblmaterialtransfer.Read, error)
}

type TblMaterialTransfer struct {
	TemplateRepo tblmaterialtransfer.Repository `inject:"tblMaterialTransferRepository"`
	ID           *formatid.GenerateIDHandler    `inject:"generateID"`
}

func (s *TblMaterialTransfer) Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

	return s.TemplateRepo.Fetch(ctx, doc, warehouseFrom, warehouseTo, startDate, endDate, param)
}

func (s *TblMaterialTransfer) Create(ctx context.Context, data *tblmaterialtransfer.Create, userName string) (*tblmaterialtransfer.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "MaterialTransfer")
	data.Note.SetNullIfEmpty()
	data.VendorCode.SetNullIfEmpty()
	data.Driver.SetNullIfEmpty()
	data.TransportType.SetNullIfEmpty()
	data.LicenceNo.SetNullIfEmpty()
	data.Status = "Outstanding"

	if err != nil {
		golog.Error(ctx, "Error generate id create material transfer: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create material transfer: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].CancelInd = booldatatype.FromBool(false)

		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		data.Details[i].CancelInd = booldatatype.FromBool(data.Details[i].CancelInd.ToBool())
		data.Details[i].SuccesInd = booldatatype.FromBool(data.Details[i].SuccesInd.ToBool())

		if data.Details[i].BatchNo == "" {
			data.Details[i].BatchNo = data.Date
		}
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create material transfer: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblMaterialTransfer) Update(ctx context.Context, data *tblmaterialtransfer.Read, userCode string) (*tblmaterialtransfer.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].CancelInd = booldatatype.FromBool(data.Details[i].CancelInd.ToBool())
		if data.Details[i].SuccesInd.ToBool() {
			data.Details[i].CancelInd = booldatatype.FromBool(false)
		}
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update material transfer: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
