package service

import (
	"context"

	"gitlab.com/ayaka/internal/domain/logactivity"
)

type TblLogService interface {
	GetLog(ctx context.Context, code, category string) ([]*logactivity.LogActivity, error)
}

type TblLog struct {
	TemplateRepo logactivity.Repository `inject:"tblLogRepository"`
}

func (s *TblLog) GetLog(ctx context.Context, code, category string) ([]*logactivity.LogActivity, error) {
	return s.TemplateRepo.GetActivityLog(ctx, code, category)
}
