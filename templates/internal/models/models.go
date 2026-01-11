package models

import (
	"time"
)

type Category struct {
	ID   int
	Name string
}

type Post struct {
	ID          int
	Title       string
	Content     string
	Author      string
	AuthorID    int
	CreatedAt   time.Time
	Categories  []Category
	Snippet     string
	Likes       int
	Dislikes    int
	Comments    []Comments
	NumComments int
}

type UserStats struct {
	PostCount     int
	LikesGiven    int
	DislikesGiven int
}

type Comments struct {
	ID        int
	AuthorID  int
	Author    string
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
}
