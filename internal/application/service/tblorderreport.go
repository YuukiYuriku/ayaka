package service

import (
	"context"
	"fmt"
	"time"

	share "gitlab.com/ayaka/internal/domain/shared"
	"gitlab.com/ayaka/internal/domain/tblorderreport"
	"gitlab.com/ayaka/internal/pkg/pagination"
)

type TblOrderReportService interface {
	ByVendor(ctx context.Context, date string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error)
}

type TblOrderReport struct {
	TemplateRepo tblorderreport.Repository `inject:"tblOrderReportRepository"`
}

func (s *TblOrderReport) ByVendor(ctx context.Context, date string, param *pagination.PaginationParam) (*pagination.PaginationResponse, error) {
	if date != "" {
		date = share.MonthYear(date)
	} else {
		date = time.Now().Format("200601")
	}
	fmt.Println("date service: ", date)
	return s.TemplateRepo.ByVendor(ctx, date, param)
}