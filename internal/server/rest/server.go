package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"post-service/domains/models"
	"post-service/internal/lib/jwt"
	"post-service/internal/storage"
	"strings"
	"github.com/rs/cors"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type Post interface {
	GetPost(ctx context.Context, postId int64, userId int64) (models.PostWithComments, error)
	GetAllPosts(ctx context.Context, userId int64) ([]models.PostFull, error)
	GetAllPostsByCreator(ctx context.Context, creatorId int64, userId int64) ([]models.Post, error)
	CreateComment(ctx context.Context, postId int64, userId int64, description string) error
}

type ServerApi struct {
	services Post
	port     string
	secret   string
}

func NewServer(services Post, port string, secret string) *http.Server {
	api := &ServerApi{
		services: services,
		port:     port,
		secret:   secret,
	}
	route := api.ConfigureRoutes()
	return &http.Server{
		Addr:    api.port,
		Handler: route,
	}
}
func (s ServerApi) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := jwt.ValidateToken(tokenStr, s.secret)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		log.Println(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *ServerApi) ConfigureRoutes() *mux.Router {
	r := mux.NewRouter()
	r.Use(s.JWTMiddleware)
	apiRouter := r.PathPrefix("/posts").Subrouter()
	apiRouter.HandleFunc("/post", s.GetPost).Methods("POST")
	apiRouter.HandleFunc("/allPost", s.GetAllPost).Methods("GET")
	apiRouter.HandleFunc("/allPostByCreator", s.GetAllPostsByCreator).Methods("POST")
	apiRouter.HandleFunc("/createComment", s.CreateComment).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	r.Use(func(next http.Handler) http.Handler {
		return c.Handler(next)
	})
	return r
}

func (s *ServerApi) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCommentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := createCommentValidation(req, w); err != nil {
		return
	}
	ctx := r.Context()
	userId, err := getUserID(ctx, w)

	if err != nil {
		return
	}

	if err := s.services.CreateComment(ctx, req.PostId, userId, req.Description); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}
func (s *ServerApi) GetAllPostsByCreator(w http.ResponseWriter, r *http.Request) {
	var req models.GetPostByCreatorRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := getPostByCreatorValidation(req, w); err != nil {
		return
	}
	ctx := r.Context()
	userId, err := getUserID(ctx, w)
	if err != nil {
		return
	}
	posts, err := s.services.GetAllPostsByCreator(ctx, req.CreatorId, userId)
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
func (s *ServerApi) GetPost(w http.ResponseWriter, r *http.Request) {
	var req models.GetPostRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := getPostValidation(req, w); err != nil {
		return
	}
	ctx := r.Context()
	userId, err := getUserID(ctx, w)
	if err != nil {
		return
	}

	post, err := s.services.GetPost(ctx, req.PostId, userId)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) || errors.As(err, &storage.ErrPostNotFound) {
			http.Error(w, "Пост не найден", http.StatusNotFound)
			return
		}

		http.Error(w, "ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"post": post,
	})
	if err != nil {
		return
	}

}

func (s *ServerApi) GetAllPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, err := getUserID(ctx, w)
	if err != nil {
		return
	}

	posts, err := s.services.GetAllPosts(ctx, userId)
	log.Println(posts)
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
func getPostValidation(req models.GetPostRequest, w http.ResponseWriter) error {
	if req.PostId == 0 {
		http.Error(w, "post id обязателен", http.StatusBadRequest)
		return fmt.Errorf("post id обязателен")
	}
	return nil
}

func getPostByCreatorValidation(req models.GetPostByCreatorRequest, w http.ResponseWriter) error {
	if req.CreatorId == 0 {
		http.Error(w, "creator id обязателен", http.StatusBadRequest)
		return fmt.Errorf("creator id обязателен")
	}
	return nil
}
func createCommentValidation(req models.CreateCommentRequest, w http.ResponseWriter) error {
	if req.Description == "" {
		http.Error(w, "description обязателен", http.StatusBadRequest)
		return fmt.Errorf("description обязателен")
	}
	if req.PostId == 0 {
		http.Error(w, "post id обязателен", http.StatusBadRequest)
		return fmt.Errorf("post id обязателен")
	}
	return nil
}
func getUserID(ctx context.Context, w http.ResponseWriter) (int64, error) {
	userId := ctx.Value("userID").(int64)
	if userId == 0 {
		http.Error(w, "user id обязателен", http.StatusBadRequest)
		return 0, fmt.Errorf("UserId не найден")
	}
	return userId, nil
}
