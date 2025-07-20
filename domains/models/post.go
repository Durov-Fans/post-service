package models

import "time"

type Post struct {
	Id        int64     `json:"id" sql:"Id"`
	UserId    int64     `json:"user_id" sql:"UserId"`
	Title     string    `json:"title" sql:"Title"`
	Media     string    `json:"media" sql:"Media"`
	CreatedAt time.Time `json:"created_at" sql:"CreatedAt"`
}

type GetPostsRequest struct {
	UserId int64
}

//type GetPostResponse struct {
//
//}
