package post

import (
	"context"
	"fmt"
	"log"
	"post-service/domains/models"
	"strings"
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
	paidArray := []models.SubInfo{{Id: 1, Level: "Supporter"}, {Id: 2, Level: "Exclusive"}, {Id: 3, Level: "Premium "}}

	if len(paidArray) == 0 {
		return nil, fmt.Errorf("лист подписок пуст")
	}

	var valueParts []string
	for _, sub := range paidArray {
		valueParts = append(valueParts, fmt.Sprintf("(%d, '%s')", sub.Id, sub.Level))
	}
	userSubsValues := strings.Join(valueParts, ", ")

	posts, err := p.postProvider.GetAllPosts(ctx, userSubsValues)
	if err != nil {
		log.Println("Ошибка получения постов")
		return nil, err
	}
	return posts, nil
}

type PostProvider interface {
	GetPost(ctx context.Context, id string, userId int64) (models.Post, error)
	GetAllPosts(ctx context.Context, subArray string) ([]models.Post, error)
}

func New(postProvider PostProvider) *Post {
	return &Post{
		postProvider: postProvider,
	}
}
