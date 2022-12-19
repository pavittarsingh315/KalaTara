package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// IMPORTANT: Use the link below when looking how to delete associated models: https://gorm.io/docs/associations.html#delete_with_select

/*
   The "Profile" field is for the "has one" relation between the User and Profile models
*/

type User struct {
	Base
	Name      string    `json:"name"`
	Contact   string    `json:"contact" gorm:"unique"`
	Password  string    `json:"password"`
	Role      string    `json:"role" gorm:"<-:create"` // allow read and create (not update)
	Strikes   uint8     `json:"strikes"`
	Birthday  time.Time `json:"birthday"`
	LastLogin time.Time `json:"last_login"`
	BanTill   time.Time `json:"ban_till"`
	Profile   Profile   `json:"profile"`
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
