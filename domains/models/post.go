package models

import (
	"encoding/json"
	"time"
)

type PostWithComments struct {
	Id          int64           `json:"id" sql:"Id"`
	UserId      int64           `json:"user_id" sql:"UserId"`
	Description string          `json:"description" sql:"description"`
	Media       string          `json:"media" sql:"Media"`
	CreatedAt   time.Time       `json:"created_at" sql:"CreatedAt"`
	Paid        bool            `json:"paid" sql:"Paid"`
	SubLevel    string          `json:"sub_level" sql:"SubLevel"`
	Comments    json.RawMessage `json:"comments" sql:"Comments"`
}
type Post struct {
	Id          int64     `json:"id" sql:"Id"`
	UserId      int64     `json:"user_id" sql:"UserId"`
	Description string    `json:"description" sql:"description"`
	Media       string    `json:"media" sql:"Media"`
	CreatedAt   time.Time `json:"created_at" sql:"CreatedAt"`
	LikeNum     int64     `json:"like_num" sql:"LikeNum"`
	Paid        bool      `json:"paid" sql:"Paid"`
	SubLevel    string    `json:"sub_level" sql:"SubLevel"`
}

type GetPostRequest struct {
	PostId int64 `json:"post_id" sql:"Id"`
}
type GetPostByCreatorRequest struct {
	CreatorId int64 `json:"creator_id" sql:"Id"`
}

//type GetPostResponse struct {
//
//}
