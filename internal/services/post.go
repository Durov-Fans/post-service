package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"post-service/domains/models"
	"post-service/internal/lib/photo"
	"post-service/internal/lib/uploaders"
	"post-service/internal/storage"
	"strconv"
	"strings"
	"sync"

	"github.com/Durov-Fans/protos/gen/go/creator"
	"google.golang.org/grpc"
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
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("failed to connect to user service: %v", err)
	}
	defer conn.Close()
	client := creator.NewCreatorServiceClient(conn)
	paidRepeat, err := client.GetTierAndCreatorId(ctx, &creator.GetTierAndCreatorIdRequest{UserId: userId})
	if err != nil {
		log.Println("Ошибка получения подписок")
		return models.PostWithComments{}, err
	}
	paidArray := convert(paidRepeat)
	postFind, err := p.postProvider.GetPost(ctx, postId)
	userResp, err := client.GetUserInfos(ctx, &creator.GetUserInfosRequest{
		UserIds: []int64{postFind.UserId},
	})
	if err != nil {
		return models.PostWithComments{}, err
	}
	if len(paidArray) > 0 {
		for _, sub := range paidArray {
			if sub.Id == postFind.UserId && !postFind.Paid {
				currentLevelNum := getLevelNum(sub.Level)
				postLevelNum := getLevelNum(postFind.SubLevel)

				if !postFind.Paid || currentLevelNum >= postLevelNum && currentLevelNum > 0 && postLevelNum > 0 {
					return models.PostWithComments{
						UserId:      postFind.UserId,
						UserName:    userResp.Users[0].Username,
						PhotoURL:    userResp.Users[0].AvatarUrl,
						CreatedAt:   postFind.CreatedAt,
						Description: postFind.Description,
						Id:          postFind.Id,
						Media:       postFind.Media,
						Paid:        postFind.Paid,
						LikeNum:     postFind.LikeNum,
						Comments:    postFind.Comments,
						SubLevel:    postFind.SubLevel,
					}, nil
				}
			} else {
				return models.PostWithComments{}, storage.ErrPostNotFound
			}
		}
	} else {
		if !postFind.Paid {
			return models.PostWithComments{
				UserId:      postFind.UserId,
				UserName:    userResp.Users[0].Username,
				PhotoURL:    userResp.Users[0].AvatarUrl,
				CreatedAt:   postFind.CreatedAt,
				Description: postFind.Description,
				Id:          postFind.Id,
				Media:       postFind.Media,
				Paid:        postFind.Paid,
				LikeNum:     postFind.LikeNum,
				Comments:    postFind.Comments,
				SubLevel:    postFind.SubLevel,
			}, nil
		}
	}
	return models.PostWithComments{}, storage.ErrPostNotFound
}
func (p Post) CreatePost(ctx context.Context, r *http.Request, userId int64, textData models.PostTextData) error {
	Urls := make(map[string]string)
	client := uploaders.InitAWS()
	PostUuid := uuid.New().String()
	log.Println(PostUuid)

	var wg sync.WaitGroup
	results := make(chan models.UploadResult, 5)

	fields := []string{"Photo_One", "Photo_Two", "Photo_Three", "Photo_Four", "Photo_Five"}

	for _, field := range fields {
		wg.Add(1)
		go func(field string) {
			defer wg.Done()
			url, err := photo.ProcessPhoto(r, field, client, strconv.FormatInt(userId, 10), PostUuid)
			results <- models.UploadResult{field, url, err}
		}(field)
	}
	wg.Wait()
	close(results)

	for res := range results {
		if res.Err != nil {

			log.Println(res.Err)
			continue
		}
		if res.URL != "" {
			Urls[res.Field] = res.URL
		}
	}
	mediaJson, err := json.Marshal(Urls)
	if err != nil {
		return err
	}

	err = p.postProvider.CreatePost(ctx, userId, string(mediaJson), textData)

	if err != nil {
		return err
	}

	return nil
}
func (p Post) CreateComment(ctx context.Context, postId int64, userId int64, description string) error {

	if err := p.postProvider.CreateComment(ctx, postId, userId, description); err != nil {

		return err
	}

	return nil
}
func (p Post) Like(ctx context.Context, postId int64, userId int64) error {

	if err := p.postProvider.Like(ctx, postId, userId); err != nil {
		return err
	}

	return nil
}
func (p Post) GetAllPostsByCreator(ctx context.Context, creatorId int64, userId int64) ([]models.Post, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {

		log.Println("failed to connect to user service: %v", err)
		return nil, err
	}
	defer conn.Close()
	client := creator.NewCreatorServiceClient(conn)
	paidRepeat, err := client.GetTierAndCreatorId(ctx, &creator.GetTierAndCreatorIdRequest{UserId: 1})
	if err != nil {
		log.Println("Ошибка получения подписок")
		return nil, err
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

func (p Post) GetAllPosts(ctx context.Context, userId int64) ([]models.PostFull, error) {
	// Подключение к localhost
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("failed to connect to creator service: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := creator.NewCreatorServiceClient(conn)

	// Получение подписок
	paidRepeat, err := client.GetTierAndCreatorId(ctx, &creator.GetTierAndCreatorIdRequest{UserId: userId})
	if err != nil {
		log.Println("Ошибка получения подписок")
		return nil, err
	}
	paidArray := convert(paidRepeat)

	// Формируем параметры подписок для SQL
	var valueParts []string
	for _, sub := range paidArray {
		valueParts = append(valueParts, fmt.Sprintf("(%d, '%s')", sub.Id, sub.Level))
	}
	userSubsValues := strings.Join(valueParts, ", ")

	// Получаем посты
	posts, err := p.postProvider.GetAllPosts(ctx, userSubsValues)
	if err != nil {
		log.Println("Ошибка получения постов")
		return nil, err
	}
	// Извлекаем уникальные userId из постов
	userIdMap := make(map[int64]struct{})
	for _, post := range posts {
		userIdMap[post.UserId] = struct{}{}
	}
	log.Println(userIdMap)

	var userIds []int64
	for id := range userIdMap {
		userIds = append(userIds, id)
	}
	userResp, err := client.GetUserInfos(ctx, &creator.GetUserInfosRequest{
		UserIds: userIds,
	})

	if err != nil {
		log.Println("Ошибка получения информации о пользователях")
		return nil, err
	}

	// Сопоставляем user_id → UserInfo
	userMap := make(map[int64]*creator.UserInfo)
	for _, u := range userResp.Users {
		userMap[u.UserId] = u
	}
	postFull := make([]models.PostFull, len(posts))
	// Добавляем UserName и PhotoURL в каждый пост
	for i := range posts {
		if u, ok := userMap[posts[i].UserId]; ok {
			postFull[i].UserName = u.Username
			postFull[i].PhotoURL = u.AvatarUrl
			postFull[i].CreatedAt = posts[i].CreatedAt
			postFull[i].Description = posts[i].Description
			postFull[i].Id = posts[i].Id
			postFull[i].Media = posts[i].Media
			postFull[i].Paid = posts[i].Paid
			postFull[i].LikeNum = posts[i].LikeNum
			postFull[i].CommentsNum = posts[i].CommentsNum
			postFull[i].SubLevel = posts[i].SubLevel
		}
	}
	log.Println(postFull)

	return postFull, nil
}

type PostProvider interface {
	GetPost(ctx context.Context, id int64) (models.PostWithComments, error)
	GetAllPosts(ctx context.Context, subArray string) ([]models.Post, error)
	GetAllPostsByCreator(ctx context.Context, subArray models.SubInfo) ([]models.Post, error)
	CreatePost(ctx context.Context, userId int64, media string, textData models.PostTextData) error
	CreateComment(ctx context.Context, postId int64, userId int64, description string) error
	Like(ctx context.Context, userId int64, postId int64) error
}

func New(postProvider PostProvider) *Post {
	return &Post{
		postProvider: postProvider,
	}
}
