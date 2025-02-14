package storage

import (
	"github.com/KaffeeMaschina/ozon_test_task/internals/graph/model"
)

type Storage interface {
	AddUser(name, email string) (*model.User, error)
	AddPost(userID string, title string, text string, allowComments bool) (*model.Post, error)
	AddComment(userId, postId, parentId, text string) (*model.Comment, error)
	GetPost(postId string) (*model.Post, error)
	GetAllPosts() ([]*model.Post, error)
}
