package handlers

import (
	"net/http"
	"strings"
	"time"

	authutils "forum/internal/authUtils"
	"forum/internal/realtime"

	"github.com/gorilla/websocket"
)

// Upgrader for Notification WS
var notifUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		return strings.Contains(origin, r.Host)
	},
}

func NotificationWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	payload := authutils.GetJWTFromContext(r.Context())
	if payload == nil || payload.Role == authutils.RoleAnonymous {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := notifUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	userID := payload.UserID
	realtime.Notif.AddConn(userID, conn)
	defer realtime.Notif.RemoveConn(userID, conn)

	// Optional: send initial connected message
	_ = conn.WriteJSON(map[string]any{
		"type":    "connected",
		"user_id": userID,
	})

	// Heartbeat / ping
	conn.SetReadLimit(64 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(realtime.PongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(realtime.PongWait))
	})

	ticker := time.NewTicker(realtime.PingPeriod)
	defer ticker.Stop()
	done := make(chan struct{})

	// Ping loop
	go func() {
		defer close(done)
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Read loop (just to keep the connection open)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			<-done
			return
		}
	}
}
