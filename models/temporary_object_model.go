package models

import (
	"time"
)

// TODO: Remove this table from the database and instead use Redis. You can give this an expires flag in Redis which will auto delete the object when its expired. Also its a lot quicker cause ya know it'd be in memory.
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
