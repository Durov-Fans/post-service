package postgres

import (
	"context"
	"fmt"
	"github.com/Durov-Fans/protos/gen/go/post"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"post-service/domains/models"
	"strings"
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

	_, err = tx.Exec(ctx, `INSERT INTO posts (description, userid, media,paid,sublevel, createdat)
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
func (s Storage) CreateComment(ctx context.Context, postId int64, userId int64, description string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Println("Ошибка транзакции")
		return err
	}
	defer tx.Rollback(ctx)

	var postExist bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM posts WHERE id = $1)`, postId).Scan(&postExist)

	if err != nil || !postExist {

		return fmt.Errorf("Такого поста не существует")
	}

	_, err = tx.Exec(ctx, `INSERT INTO comments ( userid, postId,description,updatedat, createdat)
         VALUES ($1, $2, $3, NOW(), NOW())`, userId, postId, description)

	if err != nil {
		log.Fatal("Ошибка запроса к базе данных", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Ошибка комита")
	}
	return nil
}
func (s Storage) GetPost(ctx context.Context, id int64) (models.PostWithComments, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("Ошибка создания транзакции: %s", err)
		return models.PostWithComments{}, err
	}
	defer tx.Rollback(ctx)
	var findPost models.PostWithComments
	err = tx.QueryRow(ctx, `
  SELECT 
    p.UserId,
    p.Id,
    p.Description,
    p.Media,
    p.CreatedAt,
    p.LikeNum,
    p.Paid,
    p.SubLevel,
    COALESCE(json_agg(json_build_object(
      'id', c.Id,
      'userId', c.UserId,
      'postId', c.PostId,
      'description', c.Description,
      'createdAt', c.CreatedAt,
      'updatedAt', c.UpdatedAt
    ) ORDER BY c.CreatedAt) FILTER (WHERE c.Id IS NOT NULL), '[]') AS Comments
  FROM posts p
  LEFT JOIN comments c ON c.PostId = p.Id
  WHERE p.Id = $1
  GROUP BY p.UserId, p.Id, p.Description, p.Media, p.CreatedAt, p.LikeNum, p.Paid, p.SubLevel
`, id).Scan(
		&findPost.UserId,
		&findPost.Id,
		&findPost.Description,
		&findPost.Media,
		&findPost.CreatedAt,
		&findPost.LikeNum,
		&findPost.Paid,
		&findPost.SubLevel,
		&findPost.Comments,
	)
	if err != nil {
		log.Println(err)
		return models.PostWithComments{}, err
	}
	log.Println("posts:", findPost)

	if err := tx.Commit(ctx); err != nil {
		return models.PostWithComments{}, fmt.Errorf("Ошибка комита")
	}
	return findPost, nil
}
func (s Storage) GetAllPostsByCreator(ctx context.Context, subArray models.SubInfo) ([]models.Post, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("Ошибка создания транзакции: %s", err)
		return nil, err
	}
	defer tx.Rollback(ctx)
	postsRows, err := tx.Query(ctx, `
WITH sub_levels AS (
  SELECT * FROM (VALUES
    ('None', 0),
    ('Supporter', 1),
    ('Premium', 2),
    ('Exclusive', 3)
  ) AS t(level, rank)
),
user_level AS (
  SELECT $2::text AS level
),
creator_level AS (
  SELECT sl.rank AS user_rank
  FROM user_level ul
  JOIN sub_levels sl ON sl.level = ul.level
),
accessible_posts AS (
  SELECT p.*
  FROM posts p
  JOIN sub_levels ps ON ps.level = p.SubLevel::text,
  creator_level cl
  WHERE
    p.UserId = $1
    AND (
      p.Paid = false
      OR p.SubLevel = 'None'
      OR ps.rank <= cl.user_rank
    )
)
SELECT 
  ap.Id              AS "Id",
  ap.UserId          AS "UserId",
  ap.Description     AS "Description",   
  ap.Media::text     AS "Media",
  ap.CreatedAt       AS "CreatedAt",
   ap.LikeNum        AS "LikeNum",        
  ap.Paid            AS "Paid",
  ap.SubLevel        AS "SubLevel",
  COUNT(c.Id)        AS "CommentsNum"
FROM accessible_posts ap
LEFT JOIN comments c ON c.PostId = ap.Id
GROUP BY 
  ap.Id,
  ap.Description,
  ap.UserId,
  ap.Media,
  ap.CreatedAt,
  ap.LikeNum,
  ap.Paid,
  ap.SubLevel
ORDER BY ap.CreatedAt DESC;
`, subArray.Id, subArray.Level)

	posts, err := pgx.CollectRows(postsRows, pgx.RowToStructByName[models.Post])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("posts:", posts)

	if err := tx.Commit(ctx); err != nil {
		return []models.Post{}, fmt.Errorf("Ошибка комита")
	}
	return posts, nil
}
func (s Storage) GetAllPosts(ctx context.Context, subArray string) ([]models.Post, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Printf("Ошибка создания транзакции: %s", err)
		return nil, err
	}
	defer tx.Rollback(ctx)
	log.Println(subArray)
	var userSubsQuery string
	if strings.TrimSpace(subArray) == "" {
		// Подписок нет — создаём пустую таблицу нужной структуры
		userSubsQuery = `(SELECT NULL::bigint AS user_id, NULL::text AS level WHERE false)`
	} else {
		// Есть подписки — нормальный VALUES
		userSubsQuery = fmt.Sprintf(`(VALUES %s)`, subArray)
	}
	log.Println(userSubsQuery)
	postsRows, err := tx.Query(ctx, fmt.Sprintf(`
WITH sub_levels AS (
  SELECT * FROM (VALUES
    ('None', 0),
    ('Supporter', 1),
    ('Premium', 2),
    ('Exclusive', 3)
  ) AS t(level, rank)
),
user_subs AS (
  SELECT * FROM %s AS t(user_id, level)
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
),
accessible_posts AS (
  SELECT DISTINCT 
    pwl.*
  FROM post_with_levels pwl
  LEFT JOIN user_levels ul ON ul.user_id = pwl.UserId
  WHERE
    pwl.Paid = false
    OR pwl.SubLevel = 'None'
    OR (
      ul.user_id IS NOT NULL
      AND pwl.post_rank <= ul.user_rank
    )
)
SELECT 
  ap.Id              AS "Id",
  ap.UserId          AS "UserId",
  ap.Description     AS "Description",   
  ap.Media::text     AS "Media",
  ap.CreatedAt       AS "CreatedAt",
  ap.LikeNum         AS "LikeNum",        
  ap.Paid            AS "Paid",
  ap.SubLevel        AS "SubLevel",
  COUNT(c.Id)        AS "CommentsNum"
FROM accessible_posts ap
LEFT JOIN comments c ON c.PostId = ap.Id
GROUP BY 
  ap.Id,
  ap.UserId,
  ap.Description,
  ap.Media,
  ap.CreatedAt,
  ap.LikeNum,
  ap.Paid,
  ap.SubLevel
ORDER BY ap.CreatedAt DESC
`, userSubsQuery))

	posts, err := pgx.CollectRows(postsRows, pgx.RowToStructByName[models.Post])

	if err != nil {
		log.Println(err)
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
