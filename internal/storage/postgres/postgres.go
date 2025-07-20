package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"post-service/domains/models"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s Storage) GetPost(ctx context.Context, id string, userId int64) (models.Post, error) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) GetAllPosts(ctx context.Context, userId int64) ([]models.Post, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("Ошибка создания транзакции: %s", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	postsRows, err := tx.Query(ctx, "SELECT * FROM posts")

	posts, err := pgx.CollectRows(postsRows, pgx.RowToStructByName[models.Post])

	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Println("posts:", posts)

	if err := tx.Commit(ctx); err != nil {
		return []models.Post{}, fmt.Errorf("Ошибка комита")
	}
	return posts, nil
}

func InitDB(storagePath string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(storagePath)
	if err != nil {
		log.Fatal("Ошибка парсинга строки подключения", err)
	}
	pgxCfg.MaxConns = 1
	pgxCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Ошибка подключения к базе данных")
		return nil, err
	}

	log.Println("Подключение к Postgres успешно")
	return &Storage{db: pool}, nil
}
