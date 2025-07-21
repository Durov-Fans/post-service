package postgres

import (
	"context"
	"fmt"
	"github.com/Durov-Fans/protos/gen/go/post"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"post-service/domains/models"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s Storage) CreatePost(ctx context.Context, req *post.CreatePostRequest) (*post.CreatePostResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Println("Ошибка транзакции")
		return &post.CreatePostResponse{Success: false}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO posts (title, userid, media,paid,sublevel, createdat)
         VALUES ($1, $2, $3,$4,$5, NOW())`, req.GetTitle(), req.GetUserid(), req.GetMedia(), req.GetPaid(), req.GetSubLevel())

	if err != nil {
		log.Fatal("Ошибка запроса к базе данных")
		return &post.CreatePostResponse{Success: false}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return &post.CreatePostResponse{Success: false}, fmt.Errorf("Ошибка комита")
	}
	return &post.CreatePostResponse{Success: true}, nil
}

func (s Storage) GetPost(ctx context.Context, id string, userId int64) (models.Post, error) {
	panic("dfsdf")
}

func (s Storage) GetAllPosts(ctx context.Context, subArray string) ([]models.Post, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("Ошибка создания транзакции: %s", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	postsRows, err := tx.Query(ctx, fmt.Sprintf(`WITH sub_levels AS (
  SELECT * FROM (VALUES
    ('None', 0),
    ('Supporter', 1),
    ('Premium', 2),
    ('Exclusive', 3)
  ) AS t(level, rank)
),
user_subs AS (
  SELECT * FROM (VALUES %s) AS t(user_id, level)
),
post_with_levels AS (
  SELECT
    p.*,
    sl.rank AS post_rank
  FROM posts p
  JOIN sub_levels sl ON sl.level = p.SubLevel::text
),
user_levels AS (
  SELECT
    us.user_id,
    us.level,
    sl.rank AS user_rank
  FROM user_subs us
  JOIN sub_levels sl ON sl.level = us.level
)
SELECT DISTINCT 
  pwl.UserId,
  pwl.Id,
  pwl.Title,
  pwl.Media,
  pwl.CreatedAt,
  pwl.Paid,
  pwl.SubLevel
FROM post_with_levels pwl
LEFT JOIN user_levels ul ON ul.user_id = pwl.UserId
WHERE
  pwl.Paid = false
  OR pwl.SubLevel = 'None'
  OR (
    ul.user_id IS NOT NULL
    AND pwl.post_rank <= ul.user_rank
  )
ORDER BY pwl.CreatedAt DESC
`, subArray))

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
