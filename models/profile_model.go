package models

import (
	"time"
)

/*
   The Profile - User relation is a "Has One" relation where the User has one Profile.
   UserId is the foreignKey to the user and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>

   The Profile has a many to many relation with itself for 1 field: followers => https://gorm.io/docs/many_to_many.html#Self-Referential-Many2Many

   The "SearchHistory" field is for the "has many" relation between the Profile and SearchHistory models
*/

type Profile struct {
	Base
	UserId        string          `json:"user_id" gorm:"size:191"`
	Username      string          `json:"username" gorm:"unique"`
	Name          string          `json:"name"`
	Bio           string          `json:"bio" gorm:"default:🚀🚀🚀🚀🚀🚀🚀🚀"`
	Avatar        string          `json:"avatar"`
	MiniAvatar    string          `json:"mini_avatar"`
	Birthday      time.Time       `json:"birthday"`
	Followers     []*Profile      `json:"followers" gorm:"many2many:profile_followers"`
	SearchHistory []SearchHistory `json:"search_history"`
}

// This is a custom junction table for the self-referencing many-to-many relationship between a Profile and a Follower
type ProfileFollower struct {
	ProfileId  string    `json:"followed_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	FollowerId string    `json:"follower_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt  time.Time `json:"created_at" gorm:"<-:create"`                        // allow read and create (not update)
}

// IMPORTANT: Struct is meant purely for API responses, not any database interactions
type MiniProfile struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	MiniAvatar string `json:"mini_avatar"`
}

// TODO: Use this https://gorm.io/docs/associations.html#Find-Associations and https://gorm.io/docs/associations.html#Count-Associations to get #followers/#following/#whitelist
