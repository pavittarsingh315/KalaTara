package models

import (
	"time"
)

type TemporaryObject struct {
	Base
	VerificationCode string `json:"code" gorm:"<-:create"`           // allow read and create (not update)
	Contact          string `json:"contact" gorm:"unique;<-:create"` // allow read and create (not update)
}

func (obj *TemporaryObject) IsExpired() bool {
	unixTimeNow := time.Now().Unix()
	unixTimeFiveMinAfterObjCreated := obj.CreatedAt.Add(time.Minute * 5).Unix()
	if unixTimeFiveMinAfterObjCreated <= unixTimeNow { // tempObj is expired
		return true
	} else { // tempObj is NOT expired
		return false
	}
}
