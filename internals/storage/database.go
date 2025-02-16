package storage

import (
	"context"
	"fmt"
	"github.com/KaffeeMaschina/ozon_test_task/internals/graph/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strconv"
	"time"
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

// NewPostgresStorage returns PostgresStorage structure with *pgxpool.Pool inside
func NewPostgresStorage(username, password, port, database string) (*PostgresStorage, error) {

	// Connecting to database
	pool, err := PostgresConn(username, password, port, database)
	if err != nil {
		return nil, err
	}
	log.Println("Postgres is connected")
	return &PostgresStorage{DB: pool}, nil
}

// PostgresConn connects to database, pings and returns this connection
func PostgresConn(username, password, port, database string) (*pgxpool.Pool, error) {
	const op = "storage.database.PostgresConn"

	dbUrl := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_max_conns=%s pool_max_conn_lifetime=%s",
		username, password, defaultHost, port, database, sslmodeDisable, poolMaxConn, poolMaxConnLifetime)

	// New Pool
	db, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Check if connection is ok
	err = db.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ping error %s: %w", op, err)
	}
	return db, nil
}

// GetPost returns post via id, or return error if there is no such post
func (s *PostgresStorage) GetPost(postId string) (*model.Post, error) {
	const op = "storage.database.GetPost"

	post := &model.Post{
		ID: postId,
	}

	// Getting post data from database
	err := s.DB.QueryRow(context.Background(), `SELECT user_id, title, body, permission FROM posts 
                        WHERE id = $1 `, postId).Scan(&post.UserID, &post.Title, &post.Text, &post.AllowComments)
	if err != nil {
		return nil, fmt.Errorf("unable to get post at %s: %w", op, err)
	}

	// Getting all comments related to post
	rows, err := s.DB.Query(context.Background(), `SELECT id, user_id, parent_id, body, created_at 
						FROM comments WHERE post_id = $1`, postId)
	if err != nil {
		return nil, fmt.Errorf("unable to get comments at %s: %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {

		comment := &model.Comment{
			PostID: post.ID,
		}
		createdAt := time.Time{}

		if err = rows.Scan(&comment.ID, &comment.UserID, &comment.ParentID, &comment.Text, &createdAt); err != nil {
			return nil, fmt.Errorf("unable to scan row at %s: %w", op, err)
		}
		comment.CreatedAt = fmt.Sprintf("%v", createdAt)
		post.Comments = append(post.Comments, comment)
	}

	return post, nil
}

// GetAllPosts returns all posts from database
func (s *PostgresStorage) GetAllPosts() ([]*model.Post, error) {
	const op = "storage.database.GetAllPost"

	rows, err := s.DB.Query(context.Background(), `SELECT * FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("unable to get all posts at %s: %w", op, err)
	}
	defer rows.Close()

	var posts []*model.Post

	for rows.Next() {

		post := &model.Post{}
		if err = rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Text, &post.AllowComments); err != nil {
			return nil, fmt.Errorf("unable to scan posts at %s: %w", op, err)
		}

		if post.AllowComments {
			rowsComments, err := s.DB.Query(context.Background(), `SELECT * FROM comments WHERE post_id = $1`, post.ID)

			if err != nil {
				return nil, fmt.Errorf("unable to get all commmets to post: %s, at %s: %w", post.ID, op, err)
			}
			defer rowsComments.Close()

			var comments []*model.Comment

			for rowsComments.Next() {
				comment := &model.Comment{}
				createdAt := time.Time{}
				if err = rowsComments.Scan(&comment.ID, &comment.UserID, &comment.PostID,
					&comment.ParentID, &comment.Text, &createdAt); err != nil {
					return nil, fmt.Errorf("unable to scan comment at %s: %w", op, err)
				}
				comment.CreatedAt = fmt.Sprintf("%v", createdAt)

				rowsChildren, err := s.DB.Query(context.Background(), `SELECT * FROM comments WHERE parent_id = $1`,
					comment.ID)
				if err != nil {
					return nil, fmt.Errorf("unable to get comments children: %s, at %s: %w", comment.ID, op, err)
				}
				defer rowsChildren.Close()

				var children []*model.Comment

				for rowsChildren.Next() {
					childComment := &model.Comment{}
					childCreatedAt := time.Time{}
					if err = rowsChildren.Scan(&childComment.ID, &childComment.UserID, &childComment.PostID,
						&childComment.ParentID, &childComment.Text, &childCreatedAt); err != nil {
						return nil, fmt.Errorf("unable to scan childComment at %s: %w", op, err)
					}
					childComment.CreatedAt = fmt.Sprintf("%v", childCreatedAt)
					children = append(children, childComment)
				}
				comments = append(comments, comment)
			}

			post.Comments = comments
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (s *PostgresStorage) AddUser(name, email string) (*model.User, error) {
	const op = "storage.database.AddUser"
	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer func() {
		err = tx.Rollback(context.Background())
		if err != nil {
			log.Printf("Rollback at %s error: %v", op, err)
		}
	}()
	user := &model.User{
		Username: name,
		Email:    email,
		Posts:    []*model.Post{},
	}
	err = tx.QueryRow(context.Background(), `INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id`,
		name, email).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to add user at %s: %w", op, err)
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to commit insertion at %s: %w", op, err)
	}

	return user, nil
}

func (s *PostgresStorage) AddPost(userId string, title string, text string, allowComments bool) (*model.Post, error) {
	const op = "storage.database.AddPost"
	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer func() {
		err = tx.Rollback(context.Background())
		if err != nil {
			log.Printf("Rollback at %s error: %v", op, err)
		}
	}()
	post := &model.Post{
		Title:         title,
		Text:          text,
		UserID:        userId,
		Comments:      []*model.Comment{},
		AllowComments: allowComments,
	}

	intUserId, err := strconv.Atoi(userId)
	if err != nil {
		return nil, fmt.Errorf("unable to convert user id %s to int at %s: %w", userId, op, err)
	}
	err = tx.QueryRow(context.Background(), `INSERT INTO posts (user_id, title, body, permission) 
												VALUES ($1, $2, $3, $4) RETURNING id`,
		intUserId, title, text, allowComments).Scan(&post.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to add user at %s: %w", op, err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to commit insertion at %s: %w", op, err)
	}
	return post, nil
}

func (s *PostgresStorage) AddComment(userId, postId, parentId, text string) (*model.Comment, error) {
	const op = "storage.database.AddComment"

	var permission bool
	err := s.DB.QueryRow(context.Background(), "SELECT permission FROM posts WHERE id = $1", postId).Scan(&permission)
	if err != nil {
		return nil, fmt.Errorf("unable to check if post: %v has a permission to add comments %s: %w", postId, op, err)
	}
	if !permission {
		return nil, fmt.Errorf("unable to add comment: %v has no permission to add comments %s: %w", postId, op, err)
	}
	tx, err := s.DB.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer func() {
		err = tx.Rollback(context.Background())
		if err != nil {
			log.Printf("Rollback at %s error: %v", op, err)
		}
	}()

	intUserId, err := strconv.Atoi(userId)
	if err != nil {
		return nil, fmt.Errorf("unable to convert user id %s to int at %s: %w", userId, op, err)
	}
	intPostId, err := strconv.Atoi(postId)
	if err != nil {
		return nil, fmt.Errorf("unable to convert post id %s to int at %s: %w", postId, op, err)
	}
	createdAt := time.Now()

	if parentId == postId {
		comment := &model.Comment{
			UserID:    userId,
			PostID:    postId,
			ParentID:  postId,
			Text:      text,
			CreatedAt: fmt.Sprintf("%v", createdAt),
			Children:  []*model.Comment{},
		}
		err = tx.QueryRow(context.Background(), `INSERT INTO comments (user_id, post_id, parent_id, body, created_at) 
												VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			intUserId, intPostId, intPostId, text, createdAt).Scan(&comment.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to add comment at %s: %w", op, err)
		}
		err = tx.Commit(context.Background())
		if err != nil {
			return nil, fmt.Errorf("unable to commit insertion at %s: %w", op, err)
		}
		return comment, nil

	}

	intParentId, err := strconv.Atoi(parentId)
	if err != nil {
		return nil, fmt.Errorf("unable to convert parent id %s to int at %s: %w", parentId, op, err)
	}
	comment := &model.Comment{
		UserID:    userId,
		PostID:    postId,
		ParentID:  parentId,
		Text:      text,
		CreatedAt: fmt.Sprintf("%v", createdAt),
		Children:  []*model.Comment{},
	}
	err = tx.QueryRow(context.Background(), `INSERT INTO comments (user_id, post_id, parent_id, body, created_at) 
											VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		intUserId, intPostId, intParentId, text, createdAt).Scan(&comment.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to add user at %s: %w", op, err)
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to commit insertion at %s: %w", op, err)
	}

	return comment, nil
}
