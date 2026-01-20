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

	// Always fetch suggested users
	users, err := database.GetSuggestedDMUsers(payload.UserID)
	if err != nil {
		WriteJSONError(w, "server error", http.StatusInternalServerError)
		return
	}

	// Build suggested users list with online status
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

	// Build threads list with online status and unread counts
	type threadOut struct {
		OtherUserID       int    `json:"other_user_id"`
		OtherUsername     string `json:"other_username"`
		OtherAvatar       string `json:"other_avatar"`
		LastMessageBody   string `json:"last_message_body"`
		LastMessageAt     int64  `json:"last_message_at"`
		LastMessageSender int    `json:"last_message_sender"`
		Online            bool   `json:"online"`
		UnreadCount       int    `json:"unread_count"`
	}

	outThreads := make([]threadOut, 0, len(threads))
	for _, t := range threads {
		convID, ok, err := database.GetConversationIDIfExists(
			payload.UserID,
			t.OtherUserID,
		)
		if err != nil || !ok {
			convID = 0
		}

		unread := 0
		if convID > 0 {
			if c, err := database.GetUnreadCountForConversation(
				payload.UserID,
				convID,
			); err == nil {
				unread = c
			}
		}

		outThreads = append(outThreads, threadOut{
			OtherUserID:       t.OtherUserID,
			OtherUsername:     t.OtherUsername,
			OtherAvatar:       t.OtherAvatar,
			LastMessageBody:   t.LastMessageBody,
			LastMessageAt:     t.LastMessageAt,
			LastMessageSender: t.LastMessageSender,
			Online:            realtime.DM.IsOnline(t.OtherUserID),
			UnreadCount:       unread,
		})
	}

	// Always return both threads and suggested_users
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"threads":         outThreads,
		"suggested_users": outUsers,
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
