package models

type Like struct {
	Id     int64 `json:"id" sql:"Id"`
	UserId int64 `json:"user_id" sql:"UserId"`
	Post   int64 `json:"post_id" sql:"PostId"`
}
type LikeRequest struct {
	PostId int64 `json:"post_id" sql:"PostId"`
}