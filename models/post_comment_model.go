package models

import "time"

type Comment struct {
	Base
	PostId             string    `json:"post_id" gorm:"size:191"`               // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	CommenterId        string    `json:"commenter_id" gorm:"size:191"`          // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	CommentRepliedToId string    `json:"comment_replied_to_id" gorm:"size:191"` // for info on the size parameter: https://github.com/go-gorm/gorm/issues/3369
	Body               string    `json:"body"`
	Commenter          Profile   `json:"commenter" gorm:"foreignKey:CommenterId;constraint:OnDelete:CASCADE;"`
	CommentRepliedTo   *Comment  `json:"comment_replied_to" gorm:"foreignKey:CommentRepliedToId;constraint:OnDelete:CASCADE;"`
	IsEdited           bool      `json:"is_edited"`
	Likes              []Profile `json:"likes" gorm:"many2many:comment_likes;constraint:OnDelete:CASCADE;"`
	Dislikes           []Profile `json:"dislikes" gorm:"many2many:comment_likes;constraint:OnDelete:CASCADE;"`
}

// This is a custom junction table for the many-to-many relationship between a Comment and a Liker(profile)
type CommentLike struct {
	CommentId string    `json:"comment_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	ProfileId string    `json:"liker_id" gorm:"primary_key;type:uuid;<-:create"`   // allow read and create (not update)
	CreatedAt time.Time `json:"created_at" gorm:"index;<-:create"`                 // allow read and create (not update)
}

// This is a custom junction table for the many-to-many relationship between a Comment and a Disliker(profile)
type CommentDislike struct {
	CommentId string    `json:"comment_id" gorm:"primary_key;type:uuid;<-:create"`  // allow read and create (not update)
	ProfileId string    `json:"disliker_id" gorm:"primary_key;type:uuid;<-:create"` // allow read and create (not update)
	CreatedAt time.Time `json:"created_at" gorm:"index;<-:create"`                  // allow read and create (not update)
}
