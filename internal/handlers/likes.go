package handlers

import (
	"encoding/json"
	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"net/http"
	"strconv"
	"strings"
)

func LikesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//     w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusUnauthorized)
	// w.Write([]byte(`{"error":"unauthorized"}`))
	// return

	// === parse JSON or form ===
	var targetType, action string
	var targetID int
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		var req struct {
			TargetType string `json:"target_type"`
			TargetID   int    `json:"target_id"`
			Action     string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		targetType, targetID, action = req.TargetType, req.TargetID, req.Action
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		targetType = r.FormValue("target_type")
		idstr := r.FormValue("target_id")
		action = strings.ToLower(r.FormValue("action"))
		var err error
		targetID, err = strconv.Atoi(idstr)
		if err != nil {
			http.Error(w, "Invalid target ID", http.StatusBadRequest)
			return
		}
	}

	// === validate ===
	if targetType != "post" && targetType != "comment" {
		http.Error(w, "Invalid target", http.StatusBadRequest)
		return
	}
	isLike := action == "like"
	if action != "like" && action != "dislike" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// err = database.InsertorUpdateReaction(payload.UserID, targetType, targetID, isLike)

	if err := database.InsertorUpdateReaction(payload.UserID, targetType, targetID, isLike); err != nil { // MOCK this for now
		http.Error(w, "Failed to save reaction", http.StatusInternalServerError)
		return
	}

	// Fetch updated counts and user reaction
	likes, dislikes, userReaction, err := database.GetReactionStatsAndUserReaction(payload.UserID, targetType, targetID)
	if err != nil {
		http.Error(w, "Failed to fetch updated reactions", http.StatusInternalServerError)
		return
	}

	numcomments := 0
	if targetType == "post" {
		n, err := database.CountComments(targetID)
		if err == nil {
			numcomments = n
		}
	}

	// Return JSON response with updated info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := struct {
		Likes        int     `json:"likes"`
		Dislikes     int     `json:"dislikes"`
		NumComments  int     `json:"numcomments"`   //test
		UserReaction *string `json:"user_reaction"` // "like", "dislike", or null
	}{
		Likes:       likes,
		Dislikes:    dislikes,
		NumComments: numcomments, //test
	}

	if userReaction == nil {
		resp.UserReaction = nil
	} else {
		resp.UserReaction = userReaction
	}

	json.NewEncoder(w).Encode(resp)
}
