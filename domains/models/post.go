package models

import (
	"encoding/json"
	"mime/multipart"
	"time"
)

type PostWithComments struct {
	Id          int64           `json:"id" sql:"Id"`
	UserId      int64           `json:"user_id" sql:"UserId"`
	UserName    string          `json:"user_name,omitempty"`
	PhotoURL    string          `json:"photo_url,omitempty"`
	Description string          `json:"description" sql:"description"`
	Media       json.RawMessage `json:"media" sql:"Media"`
	CreatedAt   time.Time       `json:"created_at" sql:"CreatedAt"`
	LikeNum     int64           `json:"like_num" sql:"LikeNum"`
	Paid        bool            `json:"paid" sql:"Paid"`
	SubLevel    string          `json:"sub_level" sql:"SubLevel"`
	Comments    json.RawMessage `json:"comments" sql:"Comments"`
}
type Post struct {
	Id          int64           `json:"id" sql:"Id"`
	UserId      int64           `json:"user_id" sql:"UserId"`
	Description string          `json:"description" sql:"description"`
	Media       json.RawMessage `json:"media" sql:"Media"`
	CreatedAt   time.Time       `json:"created_at" sql:"CreatedAt"`
	LikeNum     int64           `json:"like_num" sql:"LikeNum"`
	Paid        bool            `json:"paid" sql:"Paid"`
	SubLevel    string          `json:"sub_level" sql:"SubLevel"`
	CommentsNum int64           `json:"comments_num" sql:"CommentsNum"`
}
type PostFull struct {
	Id          int64           `json:"id" sql:"Id"`
	UserId      int64           `json:"user_id" sql:"UserId"`
	UserName    string          `json:"user_name,omitempty"`
	PhotoURL    string          `json:"photo_url,omitempty"`
	Description string          `json:"description" sql:"description"`
	Media       json.RawMessage `json:"media" sql:"Media"`
	CreatedAt   time.Time       `json:"created_at" sql:"CreatedAt"`
	LikeNum     int64           `json:"like_num" sql:"LikeNum"`
	Paid        bool            `json:"paid" sql:"Paid"`
	SubLevel    string          `json:"sub_level" sql:"SubLevel"`
	CommentsNum int64           `json:"comments_num" sql:"CommentsNum"`
}
type GetPostRequest struct {
	PostId int64 `json:"post_id" sql:"Id"`
}
type GetPostByCreatorRequest struct {
	CreatorId int64 `json:"creator_id" sql:"Id"`
}
type CreatePostRequest struct {
	Userid      int64  `json:"userid" sql:"UserId"`
	Description string `json:"description" sql:"description"`
	Media       string `json:"media" sql:"media"`
	Paid        bool   `json:"paid" sql:"paid"`
	SubLevel    string `json:"sub_Level" sql:"sub_level"`
}
type FileData struct {
	File   multipart.File
	Header *multipart.FileHeader
}
type PostTextData struct {
	Desc string
	Paid bool
	Type string
}
type UploadResult struct {
	Field string
	URL   string
	Err   error
}

//type GetPostResponse struct {
//
//}
