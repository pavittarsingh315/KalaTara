package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IMPORTANT: For more information on gorm tags: https://gorm.io/docs/models.html#Field-Level-Permission

type Base struct {
	Id        string    `json:"id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt time.Time `json:"created_at" gorm:"<-:create"`               // allow read and create (not update)
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	b.Id = uuid.NewString()
	return nil
}

// if there are any errors with the models, try dropping the tables and remigrating.
