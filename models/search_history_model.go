package models

/*
   The SearchHistory - Profile relation is a "Has One" relation where a Profile has one SearchHistory
   ProfileId is the foreignKey to the profile and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>
*/

type SearchHistory struct {
	Base
	ProfileId string   `json:"profile_id"`
	History   []string `json:"history" gorm:"type:text[]"`
}
