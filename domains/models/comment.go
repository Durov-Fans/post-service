package models

type Comment struct {
	Id          int64  `json:"id"`
	UserId      int64  `json:"user_id"`
	PostId      int64  `json:"post_id"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
