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
	"gitlab.com/ayaka/internal/domain/tblmaterialrequest"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblMaterialRequestService interface {
	Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblmaterialrequest.Create, userName string) (*tblmaterialrequest.Create, error)
	Update(ctx context.Context, data *tblmaterialrequest.Read, userCode string) (*tblmaterialrequest.Read, error)
	GetMaterialRequest(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	OutstandingMaterial(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblMaterialRequest struct {
	TemplateRepo tblmaterialrequest.Repository `inject:"tblMaterialRequestRepository"`
	ID           *formatid.GenerateIDHandler   `inject:"generateID"`
}

func (s *TblMaterialRequest) Fetch(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

func (s *TblMaterialRequest) Create(ctx context.Context, data *tblmaterialrequest.Create, userName string) (*tblmaterialrequest.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "MaterialRequest")
	data.SiteCode.SetNullIfEmpty()
	data.Remark.SetNullIfEmpty()

	if err != nil {
		golog.Error(ctx, "Error generate id create material request: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create material request: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		data.Details[i].Cancel = booldatatype.FromBool(false)
		data.Details[i].UsageDt, _ = share.FormatToCompactDateTime(data.Details[i].UsageDt)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create material request: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblMaterialRequest) Update(ctx context.Context, data *tblmaterialrequest.Read, userCode string) (*tblmaterialrequest.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].Cancel = booldatatype.FromBool(data.Details[i].Cancel.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update material request: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblMaterialRequest) GetMaterialRequest(ctx context.Context, doc, itemName string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	return s.TemplateRepo.GetMaterialRequest(ctx, doc, itemName, param)
}

func (s *TblMaterialRequest) OutstandingMaterial(ctx context.Context, doc, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

	return s.TemplateRepo.OutstandingMaterial(ctx, doc, startDate, endDate, param)
}