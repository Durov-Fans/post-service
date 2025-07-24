package post

import (
	"context"
	"fmt"
	"github.com/Durov-Fans/protos/gen/go/creator"
	"google.golang.org/grpc"
	"log"
	"post-service/domains/models"
	"post-service/internal/storage"
	"strings"
	"time"
)

type Post struct {
	log          *log.Logger
	postProvider PostProvider
}

func getLevelNum(level string) int {
	switch level {
	case "Supporter":
		return 1
	case "Premium":
		return 2
	case "Exclusive":
		return 3
	default:
		return 0
	}
}
func convert(resp *creator.GetTierAndCreatorIdResponse) []models.SubInfo {
	var result []models.SubInfo
	for _, item := range resp.Data {
		result = append(result, models.SubInfo{
			Id:    int64(item.CreatorId),
			Level: item.TierType,
		})
	}
	return result
}
func (p Post) GetPost(ctx context.Context, postId int64, userId int64) (models.PostWithComments, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("failed to connect to user service: %v", err)
	}
	defer conn.Close()
	client := creator.NewCreatorServiceClient(conn)
	grpcCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	paidRepeat, err := client.GetTierAndCreatorId(grpcCtx, &creator.GetTierAndCreatorIdRequest{UserId: 1})
	if err != nil {
		log.Println("Ошибка получения подписок")
	}
	paidArray := convert(paidRepeat)
	postFind, err := p.postProvider.GetPost(ctx, postId)
	if err != nil {
		return models.PostWithComments{}, err
	}
	for _, sub := range paidArray {
		if sub.Id == postFind.UserId {
			currentLevelNum := getLevelNum(sub.Level)
			postLevelNum := getLevelNum(postFind.SubLevel)

			if !postFind.Paid || currentLevelNum >= postLevelNum && currentLevelNum > 0 && postLevelNum > 0 {
				return postFind, nil
			}
		} else {
			return models.PostWithComments{}, storage.ErrPostNotFound
		}
	}
	return models.PostWithComments{}, storage.ErrPostNotFound
}
func (p Post) CreateComment(ctx context.Context, postId int64, userId int64, description string) error {

	if err := p.postProvider.CreateComment(ctx, postId, userId, description); err != nil {

		return err
	}

	return nil
}
func (p Post) GetAllPostsByCreator(ctx context.Context, creatorId int64, userId int64) ([]models.Post, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("failed to connect to user service: %v", err)
	}
	defer conn.Close()
	client := creator.NewCreatorServiceClient(conn)
	grpcCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	paidRepeat, err := client.GetTierAndCreatorId(grpcCtx, &creator.GetTierAndCreatorIdRequest{UserId: 1})
	if err != nil {
		log.Println("Ошибка получения подписок")
	}
	paidArray := convert(paidRepeat)
	var subByCreator models.SubInfo

	for _, sub := range paidArray {
		if sub.Id == creatorId {
			subByCreator = sub
		}
	}

	posts, err := p.postProvider.GetAllPostsByCreator(ctx, subByCreator)
	if err != nil {
		log.Println("Ошибка получения постов")
		return nil, err
	}
	return posts, nil

}

func (p Post) GetAllPosts(ctx context.Context, userId int64) ([]models.Post, error) {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("failed to connect to user service: %v", err)
	}
	defer conn.Close()
	client := creator.NewCreatorServiceClient(conn)
	grpcCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	paidRepeat, err := client.GetTierAndCreatorId(grpcCtx, &creator.GetTierAndCreatorIdRequest{UserId: 1})
	if err != nil {
		log.Println("Ошибка получения подписок")
	}
	paidArray := convert(paidRepeat)
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
	GetPost(ctx context.Context, id int64) (models.PostWithComments, error)
	GetAllPosts(ctx context.Context, subArray string) ([]models.Post, error)
	GetAllPostsByCreator(ctx context.Context, subArray models.SubInfo) ([]models.Post, error)
	CreateComment(ctx context.Context, postId int64, userId int64, description string) error
}

func New(postProvider PostProvider) *Post {
	return &Post{
		postProvider: postProvider,
	}
}
