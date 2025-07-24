package grpc

import (
	"context"
	"fmt"
	"github.com/Durov-Fans/protos/gen/go/post"
	"google.golang.org/grpc"
	"log"
	"net"
)

type PostGRPCServer struct {
	post.UnimplementedPostServiceServer
	PostProvider PostProvider
}
type PostProvider interface {
	CreatePost(ctx context.Context, req *post.CreatePostRequest) (*post.CreatePostResponse, error)
}

func (s PostGRPCServer) CreatePost(ctx context.Context, req *post.CreatePostRequest) (*post.CreatePostResponse, error) {
	log.Println("вызван CreatePost ")

	if err := createPostValidate(req); err != nil {
		return &post.CreatePostResponse{Success: false}, err
	}

	response, err := s.PostProvider.CreatePost(ctx, req)
	if err != nil {
		log.Println("Ошибка CreatePost")
		return &post.CreatePostResponse{Success: false}, err
	}

	return response, nil
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

func StartGRPCServer(postProvider PostProvider) {
	lis, err := net.Listen("tcp", ":5003")
	if err != nil {
		log.Fatalf("не удалось слушать порт: %v", err)
	}

	s := grpc.NewServer()
	post.RegisterPostServiceServer(s, &PostGRPCServer{PostProvider: postProvider})

	log.Println(" gRPC-сервер запущен на :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
