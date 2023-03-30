package responses

import (
	"encoding/json"
	"time"
)

// Miniture representation of a profile intended for use in responses.
type MiniProfile struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	MiniAvatar string `json:"mini_avatar"`
}

// Collective representation of a post, it's owner, it's media, and other metadata.
type Post struct {
	PostId    string    `json:"post_id"`
	Title     string    `json:"title" gorm:"column:post_title"`
	Caption   string    `json:"caption" gorm:"column:post_caption"`
	CreatedAt time.Time `json:"created_at"`

	ProfileId  string `json:"profile_id"`
	Username   string `json:"username" gorm:"column:profile_username"`
	Name       string `json:"name" gorm:"column:profile_name"`
	MiniAvatar string `json:"mini_avatar" gorm:"column:profile_mini_avatar"`

	Media json.RawMessage `json:"media" gorm:"column:media_data"`

	NumLikes     int `json:"num_likes"`
	NumDislikes  int `json:"num_dislikes"`
	NumBookmarks int `json:"num_bookmarks"`

	IsLiked      bool `json:"is_liked"`
	IsDisliked   bool `json:"is_disliked"`
	IsBookmarked bool `json:"is_bookmarked"`
}
