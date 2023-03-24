package models

/*
   The Notification - Profile relation is a "Has Many" relation where a Profile has many Notifications
   ProfileId is the foreignKey to the profile and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>
*/

type Notification struct {
	Base
	ProfileId string `json:"profile_id" gorm:"size:191"` // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	Title     string `json:"title"`
	Body      string `json:"caption"`
	Link      string `json:"link"`
	Read      bool   `json:"read"`
}
