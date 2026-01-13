package realtime

import (
	"encoding/json"
	"sync"
	
	"github.com/gorilla/websocket"
)


type NotificationPayload struct {
    ID       int64       `json:"id"`
    Type     string      `json:"type"`
    ActorID  int64       `json:"actor_id,omitempty"`
    TargetID int64       `json:"target_id,omitempty"`
    Payload  interface{} `json:"payload"`
    Read     bool        `json:"read"`
    Created  string      `json:"created_at"`
}


type NotificationHub struct {
	mu    sync.RWMutex
	conns map[int]map[*websocket.Conn]struct{}
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		conns: make(map[int]map[*websocket.Conn]struct{}),
	}
}

// Add connection
func (h *NotificationHub) AddConn(userID int, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.conns[userID]
	if !ok {
		set = make(map[*websocket.Conn]struct{})
		h.conns[userID] = set
	}
	set[c] = struct{}{}
}

// Remove connection
func (h *NotificationHub) RemoveConn(userID int, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.conns[userID]
	if !ok {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.conns, userID)
	}
}

// Check if user is online
func (h *NotificationHub) IsOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	set, ok := h.conns[userID]
	return ok && len(set) > 0
}

// Send a notification to a specific user (all connections)
func (h *NotificationHub) SendToUser(userID int, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	set := h.conns[userID]
	conns := make([]*websocket.Conn, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, b)
	}
}
