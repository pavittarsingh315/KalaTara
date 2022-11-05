package models

import (
	"time"
)

type User struct {
	Base
	Name      string    `json:"name"`
	Contact   string    `json:"contact" gorm:"unique"`
	Password  string    `json:"password"`
	Strikes   uint8     `json:"strikes"`
	Birthday  time.Time `json:"birthday"`
	LastLogin time.Time `json:"lastLogin"`
	BanTill   time.Time `json:"banTill"`
}
