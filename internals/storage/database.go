package storage

import (
	"context"
	"fmt"
	"github.com/KaffeeMaschina/ozon_test_task/graph/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strconv"
)

const (
	defaultHost         = "localhost"
	sslmodeDisable      = "disable"
	poolMaxConn         = "10"
	poolMaxConnLifetime = "1h30m"
)

type PostgresStorage struct {
	DB *pgxpool.Pool
}

func NewPostgresStorage(username, password, port, database string) (*PostgresStorage, error) {
	pool, err := PostgresConn(username, password, port, database)
	if err != nil {
		return nil, err
	}
	log.Println("Postgres is connected")
	return &PostgresStorage{DB: pool}, nil
}

func PostgresConn(username, password, port, database string) (*pgxpool.Pool, error) {
	const op = "storage.database.PostgresConn"

	dbUrl := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_max_conns=%s pool_max_conn_lifetime=%s",
		username, password, defaultHost, port, database, sslmodeDisable, poolMaxConn, poolMaxConnLifetime)

	db, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = db.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ping error %s: %w", op, err)
	}
	return db, nil
}

func (s *PostgresStorage) GetPost(postId string) (*model.Post, error) {
	const op = "storage.database.GetPost"

	permission := new(string)
	post := &model.Post{}

	err := s.DB.QueryRow(context.Background(), `SELECT post_id, fk_user_id, title, body, permission FROM posts 
                        WHERE post_id = $1`, postId).Scan(&post.ID, &post.UserID, &post.Title, &post.Text, permission)
	if err != nil {
		return nil, fmt.Errorf("unable to get post at %s: %w", op, err)
	}
	if *permission == "no" {
		post.AllowComments = false
		return post, nil
	}
	rows, err := s.DB.Query(context.Background(), `SELECT comment_id, fk_user_id, fk_parent_id, body, created_at 
						FROM comments WHERE fk_post_id = $1`, postId)
	if err != nil {
		return nil, fmt.Errorf("unable to get comments at %s: %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var comment *model.Comment
		if err = rows.Scan(&comment.ID, &comment.UserID, &comment.ParentID, &comment.Text, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("unable to scan row at %s: %w", op, err)
		}
		post.Comments = append(post.Comments, comment)
	}
	return post, nil
}
func (s *PostgresStorage) GetAllPosts() ([]*model.Post, error) {
	const op = "storage.database.GetAllPost"
	//TODO
	return nil, nil
}

func (s *PostgresStorage) AddUser(name, email string) (*model.User, error) {
	const op = "storage.database.AddUser"
	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer func() {
		err := tx.Rollback(context.Background())
		if err != nil {
			log.Printf("Rollback at %s error: %v", op, err)
		}
	}()
	user := &model.User{
		Username: name,
		Email:    email,
		Posts:    []*model.Post{},
	}
	err = tx.QueryRow(context.Background(), `INSERT INTO users (username, email) VALUES ($1, $2) RETURNING user_id`,
		name, email).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to add user at %s: %w", op, err)
	}
	return user, nil
}

func (s *PostgresStorage) AddPost(userID string, title string, text string, allowComments bool) (*model.Post, error) {
	const op = "storage.database.AddPost"
	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer func() {
		err := tx.Rollback(context.Background())
		if err != nil {
			log.Printf("Rollback at %s error: %v", op, err)
		}
	}()
	post := &model.Post{
		Title:         title,
		Text:          text,
		UserID:        userID,
		Comments:      []*model.Comment{},
		AllowComments: allowComments,
	}
	permission := "no"
	if allowComments {
		permission = "yes"
	}
	intUserId, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("unable to convert user id %s to int at %s: %w", userID, op, err)
	}
	err = tx.QueryRow(context.Background(), `INSERT INTO posts (fk_user_id, title, body, permission) 
												VALUES ($1, $2, $3, $4) RETURNING post_id`,
		intUserId, title, text, permission).Scan(&post.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to add user at %s: %w", op, err)
	}

	return post, nil
}

func (s *PostgresStorage) AddComment(userId, postId, parentId, text string) (*model.Comment, error) {
	const op = "storage.database.AddComment"
	return nil, nil
}
