package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	authutils "forum/internal/authUtils"
	"forum/internal/database"
	"forum/internal/realtime"

	"github.com/gorilla/websocket"
)

var dmUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Same-origin planned: allow same host.

	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // some clients omit it
		}
		// allow exact same host
		return strings.Contains(origin, r.Host)
	},
}

type dmClientMsg struct {
	Type             string `json:"type"`
	ToUserID         int    `json:"to_user_id"`
	Body             string `json:"body"`
	ClientMsgID      string `json:"client_msg_id"`
	ConversationWith int    `json:"conversation_with"`
	LastReadMsgID    int    `json:"last_read_msg_id"`
}

func DMWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Must be authenticated already via apiAuth (AuthMiddleware)
	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Enforce access JTI exists (same as API)
	if payload.JTI == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ok, err := database.TokenExists(payload.JTI)
	if err != nil || !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade
	conn, err := dmUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	sendErr := func(msg string) {
		_ = conn.WriteJSON(map[string]any{
			"type":  "dm_error",
			"error": msg,
		})
	}

	userID := payload.UserID

	// Register presence
	becameOnline := realtime.DM.AddConn(userID, conn)

	// 1) presence snapshot (ONCE)
	onlineIDs := realtime.DM.ListOnlineUserIDs()
	_ = conn.WriteJSON(map[string]any{
		"type":       "presence_snapshot",
		"online_ids": onlineIDs,
	})

	// 2) directory snapshot (NEW) — same shape as /api/dm/threads suggested_users
	users, err := database.GetSuggestedDMUsers(userID)
	if err == nil {
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

		_ = conn.WriteJSON(map[string]any{
			"type":  "user_directory_update",
			"users": outUsers,
		})
	}
	if becameOnline {
		realtime.DM.SendToAll(map[string]any{
			"type": "user_upsert",
			"user": map[string]any{
				"user_id":  userID,
				"username": payload.Username,
				"online":   true,
			},
		})
	}

	if becameOnline {
		undelivered, err := database.GetUndeliveredDMsForUser(userID)
		if err == nil && len(undelivered) > 0 {
			now := time.Now().Unix()

			for _, m := range undelivered {
				_ = database.MarkDMDelivered(m.MsgID, now)

				// notify original sender
				realtime.DM.SendToUser(m.SenderID, map[string]any{
					"type":          "dm_delivery",
					"server_msg_id": m.MsgID,
					"delivered":     true,
				})
			}
		}
	}

	// Broadcast to ALL users that this user came online
	if becameOnline {
		// Send to all online users, not just partners
		realtime.DM.SendPresenceToUsers(realtime.DM.ListOnlineUserIDs(), userID, true)
	}

	defer func() {
		becameOffline := realtime.DM.RemoveConn(userID, conn)
		if becameOffline {
			realtime.DM.SendToAll(map[string]any{
				"type": "user_upsert",
				"user": map[string]any{
					"user_id": userID,
					"online":  false,
				},
			})
		}
	}()

	// Heartbeat
	conn.SetReadLimit(64 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(realtime.PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(realtime.PongWait))
	})

	ticker := time.NewTicker(realtime.PingPeriod)
	defer ticker.Stop()

	// Ping loop (separate)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Read loop (blocking)
	for {
		var msg dmClientMsg
		if err := conn.ReadJSON(&msg); err != nil {
			<-done
			return
		}

		switch strings.ToLower(msg.Type) {

		// -----------------------
		// READ RECEIPT
		// -----------------------
		case "dm_read":
			lastID := msg.LastReadMsgID
			if lastID <= 0 || msg.ConversationWith <= 0 {
				continue
			}

			convID, ok, err := database.GetConversationIDIfExists(userID, msg.ConversationWith)
			if err != nil || !ok {
				continue
			}

			_ = database.UpdateLastRead(convID, lastID)

			// mirror read position for per-user unread tracking
			err = database.UpsertUserConversationLastRead(
				userID,
				convID,
				lastID,
			)
			if err != nil {
				// do NOT break read flow
				log.Println("conversation_reads mirror failed:", err)
			}
			// notify the other user
			realtime.DM.SendToUser(msg.ConversationWith, map[string]any{
				"type":              "dm_read",
				"last_read_msg_id":  lastID,
				"conversation_with": userID,
			})

			continue

		// -----------------------
		// SEND MESSAGE
		// -----------------------
		case "dm_send":

			// Validate
			if msg.ToUserID <= 0 || msg.ToUserID == userID {
				sendErr("invalid recipient")
				continue
			}

			body := strings.TrimSpace(msg.Body)
			if body == "" {
				sendErr("empty message")
				continue
			}
			if len(body) > 2000 {
				sendErr("message too long")
				continue
			}
			if msg.ClientMsgID == "" {
				sendErr("missing client_msg_id")
				continue
			}

			exists, err := database.UserExists(msg.ToUserID)
			if err != nil {
				sendErr("server error")
				continue
			}
			if !exists {
				sendErr("recipient not found")
				continue
			}

			// Persist
			convID, _, err := database.GetOrCreateConversation(userID, msg.ToUserID)
			if err != nil {
				sendErr("server error")
				continue
			}

			msgID, createdAt, err := database.InsertDM(convID, userID, body)
			if err != nil {
				sendErr("server error")
				continue
			}

			_ = conn.WriteJSON(map[string]any{
				"type":          "dm_ack",
				"client_msg_id": msg.ClientMsgID,
				"server_msg_id": msgID,
			})

			delivered := realtime.DM.IsOnline(msg.ToUserID)

			if err := database.UpdateConversationLast(convID, msgID, createdAt); err != nil {
				sendErr("server error")
				continue
			}

			// thread bumps
			realtime.DM.SendToUser(msg.ToUserID, map[string]any{
				"type":                "thread_bump",
				"other_user_id":       userID,
				"last_message_body":   body,
				"last_message_at":     createdAt,
				"last_message_sender": userID,
			})
			_ = conn.WriteJSON(map[string]any{
				"type":                "thread_bump",
				"other_user_id":       msg.ToUserID,
				"last_message_body":   body,
				"last_message_at":     createdAt,
				"last_message_sender": userID,
			})

			// send message
			realtime.DM.SendToUser(msg.ToUserID, map[string]any{
				"type":              "dm_new",
				"conversation_with": userID,
				"message": map[string]any{
					"id":              msgID,
					"sender_id":       userID,
					"sender_username": payload.Username,
					"body":            body,
					"created_at":      createdAt,
				},
			})

			if delivered {
				now := time.Now().Unix()
				_ = database.MarkDMDelivered(msgID, now)

				_ = conn.WriteJSON(map[string]any{
					"type":          "dm_delivery",
					"server_msg_id": msgID,
					"delivered":     true,
				})
			}

			_ = conn.WriteJSON(map[string]any{
				"type":              "dm_new",
				"conversation_with": msg.ToUserID,
				"message": map[string]any{
					"id":              msgID,
					"sender_id":       userID,
					"sender_username": payload.Username,
					"body":            body,
					"created_at":      createdAt,
				},
			})

		// -----------------------
		// IGNORE EVERYTHING ELSE
		// -----------------------
		default:
			continue
		}
	}

}
