package logactivity

import (
	"context"
)

type Repository interface {
	GetActivityLog(ctx context.Context, code, category string) ([]*LogActivity, error)
}