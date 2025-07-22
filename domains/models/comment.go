package models

import "database/sql"

type Comment struct {
	Id          sql.NullInt64  `json:"id"`
	UserId      sql.NullInt64  `json:"user_id"`
	PostId      sql.NullInt64  `json:"post_id"`
	Description sql.NullString `json:"description"`
	CreatedAt   sql.NullString `json:"created_at"`
	UpdatedAt   sql.NullString `json:"updated_at"`
}
type CreateCommentRequest struct {
	PostId      int64  `json:"post_id"`
	Description string `json:"description"`
}
