package handlers

import "forum/internal/models"

type PageData struct {
	Username      string
	Role          string
	User          bool
	Categories    []models.Category
	Posts         []models.Post
	LoginError    string
	RegisterError string
	Bio           string
	Avatar        string
	Stats         models.UserStats
}
