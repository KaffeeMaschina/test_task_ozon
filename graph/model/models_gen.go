// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Comment struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	PostID    string     `json:"postId"`
	ParentID  string     `json:"parentId"`
	Text      string     `json:"text"`
	CreatedAt string     `json:"createdAt"`
	Children  []*Comment `json:"children,omitempty"`
}

type Mutation struct {
}

type Post struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	Title         string     `json:"title"`
	Text          string     `json:"text"`
	Comments      []*Comment `json:"comments,omitempty"`
	AllowComments bool       `json:"allowComments"`
}

type Query struct {
}

type User struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Posts    []*Post `json:"posts,omitempty"`
}
