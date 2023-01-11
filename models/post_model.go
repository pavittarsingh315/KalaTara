package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

/*
   The Post - Profile relation is a "Has Many" relation where a Profile has many Posts
   ProfileId is the foreignKey to the profile and the syntax has to match: <OwnerModelName><OwnerModelPrimaryKeyName>

   The "Media" field is for the "has many" relation between the Post and PostMedia models
*/

// When in doubt relationships between to models, ask ChatGPT something like this "what is the relationship between a users and a posts table in a mysql server?"

type Post struct {
	Base
	ProfileId          string      `json:"profile_id" gorm:"size:191"` // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	Title              string      `json:"title"`
	Caption            string      `json:"caption"`
	ForSubscribersOnly bool        `json:"for_subscribers_only" gorm:"<-:create"` // allow read and create (not update)
	IsArchived         bool        `json:"is_archived"`
	Media              []PostMedia `json:"media"`
	Likes              []Profile   `json:"likes" gorm:"many2many:post_likes"`
	Dislikes           []Profile   `json:"dislikes" gorm:"many2many:post_dislikes"`
	Bookmarks          []Profile   `json:"bookmarks" gorm:"many2many:post_bookmarks"`
}

type PostMedia struct {
	Base
	PostId   string `json:"post_id" gorm:"size:191"` // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	Position int    `json:"position"`
	MediaUrl string `json:"media_url"`
	IsImage  bool   `json:"is_image"`
	IsVideo  bool   `json:"is_video"`
	IsAudio  bool   `json:"is_audio"`
}

func (pm *PostMedia) BeforeCreate(tx *gorm.DB) error {
	if (pm.IsImage && !pm.IsVideo && !pm.IsAudio) || (!pm.IsImage && pm.IsVideo && !pm.IsAudio) || (!pm.IsImage && !pm.IsVideo && pm.IsAudio) {
		pm.Base.BeforeCreate(tx) // refer to user_model.go BeforeCreate to learn reasoning behind this
		return nil
	}
	return errors.New("only one of the following fields can be true: is_image, is_video, is_audio. one field also must be true")
}

// This is a custom junction table for the many-to-many relationship between a Post and a Liker(profile)
type PostLike struct {
	PostId    string    `json:"post_id" gorm:"primary_key;type:uuid;<-:create"`  // allow read and create (not update)
	ProfileId string    `json:"liker_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt time.Time `json:"created_at" gorm:"index;<-:create"`               // allow read and create (not update)
}

// This is a custom junction table for the many-to-many relationship between a Post and a Disliker(profile)
type PostDislike struct {
	PostId    string    `json:"post_id" gorm:"primary_key;type:uuid;<-:create"`     // allow read and create (not update)
	ProfileId string    `json:"disliker_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt time.Time `json:"created_at" gorm:"index;<-:create"`                  // allow read and create (not update)
}

// This is a custom junction table for the many-to-many relationship between a Post and a Disliker(profile)
type PostBookmark struct {
	PostId    string    `json:"post_id" gorm:"primary_key;type:uuid;<-:create"`       // allow read and create (not update)
	ProfileId string    `json:"bookmarker_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt time.Time `json:"created_at" gorm:"index;<-:create"`                    // allow read and create (not update)
}
