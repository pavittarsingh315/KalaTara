package models

/*
   The SearchHistory - Profile relation is a "Has Many" relation where a Profile has many SearchHistory
   ProfileId is the foreignKey to the profile and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>
*/

type SearchHistory struct {
	Base
	ProfileId string `json:"profile_id" gorm:"size:191"` // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	Query     string `json:"query"`
}
