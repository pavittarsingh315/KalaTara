package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// if there are any errors with the models, try dropping the tables and remigrating.

type Base struct {
	Id        string    `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	b.Id = uuid.NewString()
	return nil
}
