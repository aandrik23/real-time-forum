package handlers

import (
	"encoding/json"
	"fmt"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/models"
	"net/http"
	"strings"
	"time"
)

type PostCreateInitResp struct {
	Categories []models.Category `json:"categories"`
}

func PostCreateInitAPIHandler(w http.ResponseWriter, r *http.Request) {
	cats, err := database.GetAllCategories()
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"categories": cats,
	})
}

type CreatePostReq struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Categories []int  `json:"categories"`
}

type CreatePostResp struct {
	OK     bool `json:"ok"`
	PostID int  `json:"post_id"`
}

func CreatePostAPIHandler(w http.ResponseWriter, r *http.Request) {
	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePostReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	var cats []models.Category
	for _, id := range req.Categories {
		cats = append(cats, models.Category{ID: id})
	}

	postID, err := database.InsertPostWithCategories(models.Post{
		Title:      title,
		Content:    content,
		Categories: cats,
		AuthorID:   payload.UserID,
		CreatedAt:  time.Now(),
	})
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreatePostResp{
		OK:     true,
		PostID: postID,
	})
}

func PostDetailAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	post, err := database.GetPostByID(id)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]any{
		"post": post,
	})
}

type DeletePostReq struct {
	PostID int `json:"post_id"`
}

type OKResp struct {
	OK bool `json:"ok"`
}

func DeleteAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeletePostReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PostID <= 0 {
		http.Error(w, "Invalid post_id", http.StatusBadRequest)
		return
	}

	if err := database.DeletePost(req.PostID); err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(OKResp{OK: true})
}
