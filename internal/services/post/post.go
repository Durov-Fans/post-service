package post

import (
	"context"
	"log"
	"post-service/domains/models"
)

type Post struct {
	log          *log.Logger
	postProvider PostProvider
}

func (p Post) GetPost(ctx context.Context, id string, userId int64) (models.Post, error) {
	//TODO implement me
	panic("implement me")
}

func (p Post) GetAllPosts(ctx context.Context, userId int64) ([]models.Post, error) {
	//TODO добавить проверку на права просмотра этого пользователя

	posts, err := p.postProvider.GetAllPosts(ctx, userId)
	if err != nil {
		log.Println("Ошибка получения постов")
		return nil, err
	}
	return posts, nil
}

type PostProvider interface {
	GetPost(ctx context.Context, id string, userId int64) (models.Post, error)
	GetAllPosts(ctx context.Context, userId int64) ([]models.Post, error)
}

func New(postProvider PostProvider) *Post {
	return &Post{
		postProvider: postProvider,
	}
}
