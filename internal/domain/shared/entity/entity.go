// Package entity represents a shared kernel for domain model
package entity

import (
	"time"

	"gitlab.com/ayaka/internal/domain/shared/identity"
	"gopkg.in/guregu/null.v4"
)

// Entity represents domain Entity
type Entity struct {
	ID        identity.ID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	CreatedAt time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time   `json:"updated_at" gorm:"not null;autoUpdateTime"`
	DeletedAt null.Time   `json:"deleted_at"`
}

func NewEntity() Entity {
	now := time.Now()
	return Entity{
		ID:        identity.NewID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
