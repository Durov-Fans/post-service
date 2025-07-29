package grpc

import (
	"fmt"
	"github.com/Durov-Fans/protos/gen/go/post"
	"log"
)

type PostGRPCServer struct {
	post.UnimplementedPostServiceServer
}

func createPostValidate(req *post.CreatePostRequest) error {
	if req.GetMedia() == "" {
		log.Printf("Медиа файлы обязательны")
		return fmt.Errorf("Медиа файлы обязательны")
	}
	if req.GetTitle() == "" {
		log.Printf("Название  обязательно")
		return fmt.Errorf("Название обязательно")
	}

	if req.GetUserid() == 0 {
		log.Printf("id пользователя обязателен")
		return fmt.Errorf("id пользователя обязателен")
	}
	return nil
}
