package storage

import (
	"fmt"
	"github.com/KaffeeMaschina/ozon_test_task/graph/model"
	"github.com/google/uuid"
	"log"
	"time"
)

const (
	maxCommentLength = 2000
)

type Cache struct {
	UserCache     map[string]*model.User
	PostsCache    map[string]*model.Post
	CommentsCache map[string]*model.Comment
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	return &Cache{
		UserCache:     make(map[string]*model.User),
		PostsCache:    make(map[string]*model.Post),
		CommentsCache: make(map[string]*model.Comment),
	}
}

// GetPost returns post via id, or return error if there is no such post
func (c *Cache) GetPost(postId string) (*model.Post, error) {
	post, ok := c.PostsCache[postId]
	if !ok {
		return nil, fmt.Errorf("No such post: %v", postId)
	}
	return post, nil
}

// GetAllPosts returns all posts from cache
func (c *Cache) GetAllPosts() ([]*model.Post, error) {
	var posts []*model.Post
	for _, post := range c.PostsCache {
		posts = append(posts, post)
	}
	if len(posts) == 0 {
		log.Println("There is no post in the cache")
	}
	return posts, nil
}

// AddUser adds user to cache, and returns this user
// or returns error if there is already a user with such name or such email.
func (c *Cache) AddUser(name, email string) (*model.User, error) {

	// Check if there is a user with such name or email
	for _, user := range c.UserCache {
		if user.Username == name {
			return nil, fmt.Errorf("User with username: %v is already exists", name)
		}
		if user.Email == email {
			return nil, fmt.Errorf("User with email: %v is already exists", email)
		}
	}

	var posts []*model.Post

	// Create a user with new uuid
	id := uuid.New()
	user := &model.User{
		ID:       id.String(),
		Username: name,
		Email:    email,
		Posts:    posts,
	}

	// Add user to cache
	c.UserCache[id.String()] = user

	log.Printf("User: %v, is successfully added", name)
	return user, nil
}

// AddPost adds post to cache, and returns this post or returns error if the is no such user, empty text or title
func (c *Cache) AddPost(userID string, title string, text string, allowComments bool) (*model.Post, error) {
	var comments []*model.Comment

	// Return error if there is no such user
	user, ok := c.UserCache[userID]
	if !ok {
		return nil, fmt.Errorf("User: %v doesn't exist", userID)
	}

	// Returns error if the title
	if title == "" {
		return nil, fmt.Errorf("Title: %v is empty", userID)
	}
	// Returns error if the text is empty
	if text == "" {
		return nil, fmt.Errorf("User: %v doesn't exist", userID)
	}

	if !allowComments {
		comments = nil
	}

	// Create a post with new uuid
	id := uuid.New()
	post := &model.Post{
		ID:            id.String(),
		UserID:        userID,
		Title:         title,
		Text:          text,
		Comments:      comments,
		AllowComments: allowComments,
	}
	// Add post to users posts
	user.Posts = append(user.Posts, post)

	// Add post to cache
	c.PostsCache[post.ID] = post
	log.Printf("Post: %v, is successfully added", title)
	return post, nil
}

// AddComment adds comment to cache, and returns this comment or returns error if there is no such user or post,
// if comments are not allowed. It returns error if comment is empty or more then 2000 symbols.
func (c *Cache) AddComment(userId, postId, parentId, text string) (*model.Comment, error) {

	// Check the length of the comment, if it is empty or more the 2000 symbols return mistake
	if text == "" {
		return nil, fmt.Errorf("Comment: %v is empty", text)
	}
	if len([]rune(text)) > maxCommentLength {
		return nil, fmt.Errorf("Comment: %v is too long, it should be no more then a %v symbols", text, maxCommentLength)
	}

	// Check if there is a user
	_, ok := c.UserCache[userId]
	if !ok {
		return nil, fmt.Errorf("User: %v doesn't exist", userId)
	}
	// Check if there is a post
	post, ok := c.PostsCache[postId]
	if !ok {
		return nil, fmt.Errorf("Post: %v doesn't exist", postId)
	}
	// Check if comments are allowed
	if !post.AllowComments {
		return nil, fmt.Errorf("Comments for post: %v are not allowed", postId)
	}

	var children []*model.Comment
	// Check if parent is a post, create a comment and add it to comment cache and to post cache
	if parentId == postId {

		id := uuid.New()
		comment := &model.Comment{
			ID:        id.String(),
			UserID:    userId,
			PostID:    postId,
			ParentID:  postId,
			Text:      text,
			CreatedAt: fmt.Sprintf("%v", time.Now()),
			Children:  children,
		}

		c.CommentsCache[comment.ID] = comment

		post.Comments = append(post.Comments, comment)

		log.Printf("Comment to post: %v, is successfully added", text)
		return comment, nil
	}

	// If parent is a comment, create comment, add it to comment cache, add it to parent's children comments
	// and add it to post cache
	id := uuid.New()
	comment := &model.Comment{
		ID:        id.String(),
		UserID:    userId,
		PostID:    postId,
		ParentID:  parentId,
		Text:      text,
		CreatedAt: fmt.Sprintf("%v", time.Now()),
		Children:  children,
	}

	c.CommentsCache[comment.ID] = comment

	parent := c.CommentsCache[parentId]
	parent.Children = append(parent.Children, comment)

	post.Comments = append(post.Comments, comment)

	log.Printf("Comment to another comment: %v, is successfully added", text)
	return comment, nil
}
