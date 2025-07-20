package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"post-service/domains/models"
)

type Post interface {
	GetPost(ctx context.Context, id string, userId int64) (models.Post, error)
	GetAllPosts(ctx context.Context, userId int64) ([]models.Post, error)
}

type ServerApi struct {
	services Post
	port     string
}

func NewServer(services Post, port string) *http.Server {
	api := &ServerApi{
		services: services,
		port:     port,
	}
	route := api.ConfigureRoutes()
	return &http.Server{
		Addr:    api.port,
		Handler: route,
	}
}

func (s *ServerApi) ConfigureRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/post", s.GetPost).Methods("GET")
	r.HandleFunc("/allPost", s.GetAllPost).Methods("GET")

	return r
}

func (s *ServerApi) GetPost(w http.ResponseWriter, r *http.Request) {

}

func (s *ServerApi) GetAllPost(w http.ResponseWriter, r *http.Request) {
	var req models.GetPostsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := getAllPostValidation(req, w); err != nil {
		return
	}
	ctx := r.Context()

	posts, err := s.services.GetAllPosts(ctx, req.UserId)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": posts,
	})
	if err != nil {
		return
	}
}
func getAllPostValidation(req models.GetPostsRequest, w http.ResponseWriter) error {
	if req.UserId == 0 {
		http.Error(w, "user id обязателен", http.StatusBadRequest)
		return fmt.Errorf("user id обязателен")
	}
	return nil
}
