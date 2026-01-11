package handlers

import (
	"encoding/json"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"net/http"
	"strconv"
	"time"
)

type CreateCommentResp struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Likes     int       `json:"likes"`
	Dislikes  int       `json:"dislikes"`
}

func CommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment struct {
		PostID  int    `json:"post_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if comment.Content == "" || comment.PostID == 0 {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	c, err := database.InsertComment(payload.UserID, comment.PostID, comment.Content)
	if err != nil {
		http.Error(w, "Failed to save comment", http.StatusInternalServerError)
		return
	}

	resp := CreateCommentResp{
		ID:        c.ID,
		Author:    payload.Username,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		Likes:     0,
		Dislikes:  0,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

// GET /api/posts/comments?post_id=123
func CommentsGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		http.Error(w, "Invalid post_id", http.StatusBadRequest)
		return
	}

	comments, err := database.GetCommentsForPost(postID)
	if err != nil {
		http.Error(w, "Failed to load comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(comments)
}
