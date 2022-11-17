package models

type TemporaryObject struct {
	Base
	VerificationCode string `json:"code" gorm:"<-:create"`           // allow read and create (not update)
	Contact          string `json:"contact" gorm:"unique;<-:create"` // allow read and create (not update)
}
