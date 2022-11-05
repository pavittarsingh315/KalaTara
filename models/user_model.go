package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Base
	Name      string    `json:"name"`
	Contact   string    `json:"contact" gorm:"unique"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	Strikes   uint8     `json:"strikes"`
	Birthday  time.Time `json:"birthday"`
	LastLogin time.Time `json:"lastLogin"`
	BanTill   time.Time `json:"banTill"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.Base.BeforeCreate(tx)
	// If you implement hooks, you will need to call the respective Base model hook if applicable because it will not fire automatically.
	// Only implement hooks if you need specific functionality for this model. If not needed, delete the hook and the Base model hooks will fire automatically.
	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	if u.Role == "admin" {
		return errors.New("cannot delete admin user")
	}
	return nil
}
