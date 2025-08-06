package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/rs/cors"
	"log"
	"net/http"
	"post-service/domains/models"
	"post-service/internal/lib/jwt"
	"post-service/internal/storage"
	"strconv"
)

type Post interface {
	GetPost(ctx context.Context, postId int64, userId int64) (models.PostWithComments, error)
	GetAllPosts(ctx context.Context, userId int64) ([]models.PostFull, error)
	GetAllPostsByCreator(ctx context.Context, creatorId int64, userId int64) ([]models.Post, error)
	CreatePost(ctx context.Context, r *http.Request, userId int64, textData models.PostTextData) error
	Like(ctx context.Context, userId int64, postId int64) error
	CreateComment(ctx context.Context, postId int64, userId int64, description string) error
}

type ServerApi struct {
	services Post
	port     string
	jwt      *jwt.JWT
}

func New(services Post, port string, jwt *jwt.JWT) *ServerApi {
	api := ServerApi{
		services: services,
		port:     port,
		jwt:      jwt,
	}
	return &api
}

func (s *ServerApi) ConfigureRoutes() *mux.Router {
	r := mux.NewRouter()
	r.Use(s.jwt.JWTMiddleware)

	apiRouter := r.PathPrefix("/posts").Subrouter()
	apiRouter.HandleFunc("/post/{id}", s.GetPost).Methods("GET")
	apiRouter.HandleFunc("/allPost", s.GetAllPost).Methods("GET")
	apiRouter.HandleFunc("/allPostByCreator/{id}", s.GetAllPostsByCreator).Methods("GET")
	apiRouter.HandleFunc("/createComment", s.CreateComment).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	// подключаем как middleware
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
	id, err := getId(r, w)
	if err != nil {
		return
	}
	ctx := r.Context()
	userId, err := getUserID(ctx, w)
	if err != nil {
		return
	}
	posts, err := s.services.GetAllPostsByCreator(ctx, id, userId)
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
	id, err := getId(r, w)
	if err != nil {
		return
	}
	ctx := r.Context()
	userId, err := getUserID(ctx, w)
	if err != nil {
		return
	}

	post, err := s.services.GetPost(ctx, id, userId)
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
func (s *ServerApi) Like(w http.ResponseWriter, r *http.Request) {
	var req models.LikeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := createLikeValidation(req, w); err != nil {
		return
	}
	ctx := r.Context()
	userId, err := getUserID(ctx, w)
	if err != nil {
		return
	}

	if err := s.services.Like(ctx, req.PostId, userId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return

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
func getId(r *http.Request, w http.ResponseWriter) (int64, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	postId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return 0, fmt.Errorf("post id обязателен")
	}
	if postId == 0 {
		http.Error(w, "post id обязателен", http.StatusBadRequest)
		return 0, fmt.Errorf("post id обязателен")
	}
	return postId, nil
}
func createPostValidation(r *http.Request) models.PostTextData {
	var Desc string
	var Paid bool
	var Type string
	if r.FormValue("Desc") == "" {
		Desc = ""
	} else {
		Desc = r.FormValue("Desc")
	}
	if r.FormValue("Type") != "None" && r.FormValue("Type") != "Supporter" && r.FormValue("Type") != "Premium" && r.FormValue("Type") != "Exclusive" {
		Type = "None"
		Paid = false
	} else {
		Type = r.FormValue("Type")
		Paid = true
	}
	return models.PostTextData{
		Desc: Desc,
		Paid: Paid,
		Type: Type,
	}
}
func getPostByCreatorValidation(req models.GetPostByCreatorRequest, w http.ResponseWriter) error {
	if req.CreatorId == 0 {
		http.Error(w, "creator id обязателен", http.StatusBadRequest)
		return fmt.Errorf("creator id обязателен")
	}
	return nil
}
func createLikeValidation(req models.LikeRequest, w http.ResponseWriter) error {
	if req.PostId == 0 {
		http.Error(w, "post id обязателен", http.StatusBadRequest)
		return fmt.Errorf("post id обязателен")
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
