package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/realtime"
)

func DMThreadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	threads, err := database.GetDMThreads(payload.UserID)
	if err != nil {
		WriteJSONError(w, "server error", http.StatusInternalServerError)
		return
	}

	// If user has no threads: return suggested users alphabetically
	if len(threads) == 0 {
		users, err := database.GetSuggestedDMUsers(payload.UserID)
		if err != nil {
			WriteJSONError(w, "server error", http.StatusInternalServerError)
			return
		}

		// include online flag based on WS presence
		type sug struct {
			UserID   int    `json:"user_id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
			Status   string `json:"status"`
			Online   bool   `json:"online"`
		}
		outUsers := make([]sug, 0, len(users))
		for _, u := range users {
			outUsers = append(outUsers, sug{
				UserID:   u.UserID,
				Username: u.Username,
				Avatar:   u.Avatar,
				Status:   u.Status,
				Online:   realtime.DM.IsOnline(u.UserID),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"threads":         []any{},
			"suggested_users": outUsers,
		})
		return
	}

	// Add online flag to threads
	type threadOut struct {
		OtherUserID       int    `json:"other_user_id"`
		OtherUsername     string `json:"other_username"`
		OtherAvatar       string `json:"other_avatar"`
		LastMessageBody   string `json:"last_message_body"`
		LastMessageAt     int64  `json:"last_message_at"`
		LastMessageSender int    `json:"last_message_sender"`
		Online            bool   `json:"online"`
	}

	out := make([]threadOut, 0, len(threads))
	for _, t := range threads {
		out = append(out, threadOut{
			OtherUserID:       t.OtherUserID,
			OtherUsername:     t.OtherUsername,
			OtherAvatar:       t.OtherAvatar,
			LastMessageBody:   t.LastMessageBody,
			LastMessageAt:     t.LastMessageAt,
			LastMessageSender: t.LastMessageSender,
			Online:            realtime.DM.IsOnline(t.OtherUserID),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"threads": out,
	})
}

func DMMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	otherStr := r.URL.Query().Get("user_id")
	if otherStr == "" {
		WriteJSONError(w, "user_id required", http.StatusBadRequest)
		return
	}
	otherID, err := strconv.Atoi(otherStr)
	if err != nil || otherID <= 0 {
		WriteJSONError(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	if otherID == payload.UserID {
		WriteJSONError(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	beforeID := 0
	if b := r.URL.Query().Get("before_id"); b != "" {
		v, err := strconv.Atoi(b)
		if err != nil || v <= 0 {
			WriteJSONError(w, "invalid before_id", http.StatusBadRequest)
			return
		}
		beforeID = v
	}
	limit := 10

	msgs, _, lastRead, err := database.GetDMMessagesWithUser(payload.UserID, otherID, beforeID, limit)
	if err != nil {
		// Keep errors generic for safety
		WriteJSONError(w, "server error", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"messages":         msgs,
		"last_read_msg_id": lastRead,
	})
}
