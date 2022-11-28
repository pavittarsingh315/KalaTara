package models

import (
	"time"
)

/*
   The Profile - User relation is a "Has One" relation where the User has one Profile.
   UserId is the foreignKey to the user and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>

   The Profile has a many to many relation with itself for 1 field: followers => https://gorm.io/docs/many_to_many.html#Self-Referential-Many2Many
*/

type Profile struct {
	Base
	UserId     string     `json:"user_id"`
	Username   string     `json:"username" gorm:"unique"`
	Name       string     `json:"name"`
	Bio        string     `json:"bio" gorm:"default:🚀🚀🚀🚀🚀🚀🚀🚀"`
	Avatar     string     `json:"avatar"`
	MiniAvatar string     `json:"mini_avatar"`
	Birthday   time.Time  `json:"birthday"`
	Followers  []*Profile `json:"followers" gorm:"many2many:profile_followers"`
}

type ProfileFollower struct {
	ProfileId  string    `json:"followed_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	FollowerId string    `json:"follower_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt  time.Time `json:"created_at" gorm:"<-:create"`                        // allow read and create (not update)
}
