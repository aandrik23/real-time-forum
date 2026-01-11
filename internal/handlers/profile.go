package handlers

import (
	"encoding/json"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/models"
	"net/http"
)

type ProfileAPIResponse struct {
	User       bool              `json:"user"`
	Username   string            `json:"username,omitempty"`
	Role       string            `json:"role,omitempty"`
	Bio        string            `json:"bio,omitempty"`
	Avatar     string            `json:"avatar,omitempty"`
	Stats      models.UserStats  `json:"stats"`
	Posts      []models.Post     `json:"posts"`
	Categories []models.Category `json:"categories"`
}

func ProfileAPIHandler(w http.ResponseWriter, r *http.Request) {
	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp := ProfileAPIResponse{
		User:  true,
		Posts: []models.Post{},
	}

	resp.Username = payload.Username
	resp.Role = payload.Role
	resp.Bio = payload.Bio
	resp.Avatar = payload.Avatar

	posts, err := database.GetPostsByAuthorID(payload.UserID)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}
	if len(posts) > 2 {
		posts = posts[:2]
	}
	resp.Posts = posts

	stats, err := database.GetUserStats(payload.UserID)
	if err != nil {
		http.Error(w, "Failed to load user stats", http.StatusInternalServerError)
		return
	}
	resp.Stats = stats

	cats, err := database.GetAllCategories()
	if err != nil {
		http.Error(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}
	resp.Categories = cats

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(resp)
}

type ProfileUpdateRequest struct {
	Username   string `json:"username"`
	Bio        string `json:"bio"`
	AvatarSeed string `json:"avatarSeed"`
}

type profileUpdateErrorResponse struct {
	Error string `json:"error"`
}

type profileUpdateSuccessResponse struct {
	Username   string `json:"username"`
	Bio        string `json:"bio"`
	AvatarSeed string `json:"avatarSeed"`
}

func writeProfileError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(profileUpdateErrorResponse{Error: msg})
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeProfileError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		writeProfileError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProfileError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Basic validation
	if len(req.Username) < 3 || len(req.Username) > 20 {
		writeProfileError(w, http.StatusBadRequest, "Username must be 3–20 characters")
		return
	}

	// Persist changes
	if err := database.ChangeUserDataFromDb(
		payload.UserID,
		req.Username,
		req.Bio,
		req.AvatarSeed,
	); err != nil {
		writeProfileError(w, http.StatusInternalServerError, "Update failed")
		return
	}

	// Expire old tokens
	authutils.ExpireTokens(w, r)

	// Generate new tokens with updated bio/avatar
	authutils.CreateTokens(
		w,
		req.Username,
		payload.Role,
		req.Bio,
		req.AvatarSeed,
		payload.UserID,
	)

	// Return JSON so SPA can optionally use it
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(profileUpdateSuccessResponse{
		Username:   req.Username,
		Bio:        req.Bio,
		AvatarSeed: req.AvatarSeed,
	})
}
