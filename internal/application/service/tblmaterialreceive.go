package service

import (
	"context"
	"fmt"

	// "errors"
	"time"

	"github.com/runsystemid/golog"
	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/shared/formatid"
	"gitlab.com/ayaka/internal/domain/tblmaterialreceive"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblMaterialReceiveService interface {
	Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tblmaterialreceive.Create, userName string) (*tblmaterialreceive.Create, error)
}

type TblMaterialReceive struct {
	TemplateRepo tblmaterialreceive.Repository `inject:"tblMaterialReceiveRepository"`
	ID           *formatid.GenerateIDHandler   `inject:"generateID"`
}

func (s *TblMaterialReceive) Fetch(ctx context.Context, doc, warehouseFrom, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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

func (s *TblMaterialReceive) Create(ctx context.Context, data *tblmaterialreceive.Create, userName string) (*tblmaterialreceive.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "MaterialReceive")
	data.Remark.SetNullIfEmpty()

	if err != nil {
		golog.Error(ctx, "Error generate id create material receive: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create material receive: " + err.Error())
	}

	t, err := time.Parse("2006-01-02", data.Date)
	if err != nil {
		return nil, err
	}
	data.Date = t.Format("20060102")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].Remark.SetNullIfEmpty()

		data.Details[i].DNo = fmt.Sprintf("%03d", i+1)
		if data.Details[i].BatchNo == "" {
			data.Details[i].BatchNo = data.Date
		}
		data.Details[i].Source = fmt.Sprintf("%s*%s*%s", data.Date[6:8], data.DocNo, data.Details[i].DNo)
	}

	res, err := s.TemplateRepo.Create(ctx, data)
	if err != nil {
		golog.Error(ctx, "Error create material receive: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}