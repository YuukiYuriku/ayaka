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
	"gitlab.com/ayaka/internal/domain/tbldirectmaterialreceive"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblDirectMaterialReceiveService interface {
	Fetch(ctx context.Context, doc, warehouse, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
	Create(ctx context.Context, data *tbldirectmaterialreceive.Create, userName string) (*tbldirectmaterialreceive.Create, error)
	Update(ctx context.Context, data *tbldirectmaterialreceive.Read, userCode string) (*tbldirectmaterialreceive.Read, error)
}

type TblDirectMaterialReceive struct {
	TemplateRepo tbldirectmaterialreceive.Repository `inject:"tblDirectMaterialReceiveRepository"`
	ID           *formatid.GenerateIDHandler         `inject:"generateID"`
}

func (s *TblDirectMaterialReceive) Fetch(ctx context.Context, doc, warehouse, warehouseTo, startDate, endDate string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
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
	return s.TemplateRepo.Fetch(ctx, doc, warehouse, warehouseTo, startDate, endDate, param)
}

func (s *TblDirectMaterialReceive) Create(ctx context.Context, data *tbldirectmaterialreceive.Create, userName string) (*tbldirectmaterialreceive.Create, error) {
	data.CreateBy = userName
	data.CreateDt = time.Now().Format("200601021504")

	var err error
	data.DocNo, err = s.ID.GenerateID(ctx, "DirectMaterialReceive")
	data.Remark.SetNullIfEmpty()
	if err != nil {
		golog.Error(ctx, "Error generate id create direct material receive: "+err.Error(), err)
		return nil, fmt.Errorf("Error generate id create direct material receive: " + err.Error())
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
		golog.Error(ctx, "Error create direct material receive: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}

func (s *TblDirectMaterialReceive) Update(ctx context.Context, data *tbldirectmaterialreceive.Read, userCode string) (*tbldirectmaterialreceive.Read, error) {
	lastUpDt := time.Now().Format("200601021504")

	for i := 0; i < len(data.Details); i++ {
		data.Details[i].Cancel = booldatatype.FromBool(data.Details[i].Cancel.ToBool())
	}

	res, err := s.TemplateRepo.Update(ctx, userCode, lastUpDt, data)
	if err != nil {
		golog.Error(ctx, "Error update direct material receive: "+err.Error(), err)
		return nil, err
	}

	return res, nil
}
