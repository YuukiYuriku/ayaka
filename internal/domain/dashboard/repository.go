package dashboard

import "context"

type Repository interface {
	Fetch(ctx context.Context) (*Read, error)
}