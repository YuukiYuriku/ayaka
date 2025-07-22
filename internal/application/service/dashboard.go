package service

import (
	"context"

	"gitlab.com/ayaka/internal/domain/dashboard"
)

type DashboardService interface {
	Fetch(ctx context.Context) (*dashboard.Read, error)
}

type Dashboard struct {
	TemplateRepo dashboard.Repository `inject:"DashboardRepository"`
}

func (s *Dashboard) Fetch(ctx context.Context) (*dashboard.Read, error) {
	return s.TemplateRepo.Fetch(ctx)
}