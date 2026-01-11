package handlers

import (
	"encoding/json"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/models"
	"log"
	"net/http"
	"sort"
	"strconv"
)

type HomeAPIResponse struct {
	User       bool              `json:"user"`
	Username   string            `json:"username,omitempty"`
	Role       string            `json:"role,omitempty"`
	Posts      []models.Post     `json:"posts"`
	Categories []models.Category `json:"categories"`
}

func HomeAPIHandler(w http.ResponseWriter, r *http.Request) {
	payload := authutils.GetJWTFromContext(r.Context())

	resp := HomeAPIResponse{
		Posts: []models.Post{},
	}

	if payload != nil && payload.Role != authutils.RoleAnonymous {
		resp.User = true
		resp.Username = payload.Username
		resp.Role = payload.Role
	} else {
		resp.User = false
	}

	// same query params as your HTML handler
	filter := r.URL.Query().Get("filter")
	categoryIDStr := r.URL.Query().Get("category")

	var posts []models.Post
	var err error

	if categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category id", http.StatusBadRequest)
			return
		}
		posts, err = database.GetPostsByCategoryID(categoryID)
	} else {
		switch filter {
		case "created":
			if payload != nil {
				posts, err = database.GetPostsByAuthorID(payload.UserID)
			}
		case "liked":
			if payload != nil {
				posts, err = database.GetLikedPostsByUserID(payload.UserID, "liked")
			}
		case "disliked":
			if payload != nil {
				posts, err = database.GetLikedPostsByUserID(payload.UserID, "disliked")
			}
		default:
			posts, err = database.GetAllPosts()
		}
	}

	if err != nil {
		log.Printf("Error loading posts with filter %q: %v", filter, err)
		http.Error(w, "unable to load posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})
	resp.Posts = posts

	cats, err := database.GetAllCategories()
	if err != nil {
		http.Error(w, "unable to load categories", http.StatusInternalServerError)
		return
	}
	resp.Categories = cats

	// JSON response
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "failed to encode json", http.StatusInternalServerError)
		return
	}
}
