package models

import "time"

/*
   The Profile - User relation is a "Has One" relation where the User has one Profile.
   UserId is the foreignKey to the user and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>
*/

type Profile struct {
	Base
	UserId        string    `json:"user_id"`
	Username      string    `json:"username" gorm:"unique"`
	Name          string    `json:"name"`
	Bio           string    `json:"bio" gorm:"default:🚀🚀🚀🚀🚀🚀🚀🚀"`
	Avatar        string    `json:"avatar"`
	MiniAvatar    string    `json:"mini_avatar"`
	NumFollowers  uint32    `json:"num_followers"`
	NumFollowing  uint32    `json:"num_following"`
	WhitelistSize uint32    `json:"whitelist_size"`
	Birthday      time.Time `json:"birthday"`
}
