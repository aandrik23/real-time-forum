package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type DMHub struct {
	mu sync.RWMutex

	// userID -> set of conns
	conns map[int]map[*websocket.Conn]struct{}
}

func (h *DMHub) ListOnlineUserIDs() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]int, 0, len(h.conns))
	for userID, set := range h.conns {
		if len(set) > 0 {
			out = append(out, userID)
		}
	}
	return out
}

func NewDMHub() *DMHub {
	return &DMHub{
		conns: make(map[int]map[*websocket.Conn]struct{}),
	}
}

func (h *DMHub) AddConn(userID int, c *websocket.Conn) (becameOnline bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.conns[userID]
	if !ok {
		set = make(map[*websocket.Conn]struct{})
		h.conns[userID] = set
	}
	wasEmpty := len(set) == 0
	set[c] = struct{}{}
	return wasEmpty // if it was empty and now has 1 => became online
}

func (h *DMHub) RemoveConn(userID int, c *websocket.Conn) (becameOffline bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.conns[userID]
	if !ok {
		return false
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.conns, userID)
		return true
	}
	return false
}

func (h *DMHub) IsOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	set, ok := h.conns[userID]
	return ok && len(set) > 0
}

func (h *DMHub) OnlineMap(userIDs []int) map[int]bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[int]bool, len(userIDs))
	for _, id := range userIDs {
		set, ok := h.conns[id]
		out[id] = ok && len(set) > 0
	}
	return out
}

func (h *DMHub) SendPresenceToUsers(targetUserIDs []int, userID int, online bool) {
	msg := map[string]any{
		"type":    "presence",
		"user_id": userID,
		"online":  online,
	}
	for _, t := range targetUserIDs {
		h.SendToUser(t, msg)
	}
}

func (h *DMHub) SendToUser(userID int, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	set := h.conns[userID]
	// copy keys to avoid holding lock while writing
	conns := make([]*websocket.Conn, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, b)
	}
}

func (h *DMHub) BroadcastPresence(userID int, online bool) {
	msg := map[string]any{
		"type":    "presence",
		"user_id": userID,
		"online":  online,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	// broadcast to everyone connected
	var all []*websocket.Conn
	for _, set := range h.conns {
		for c := range set {
			all = append(all, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range all {
		_ = c.WriteMessage(websocket.TextMessage, b)
	}
}

func (h *DMHub) ForceDisconnectUser(userID int) {
	h.mu.Lock()
	set, ok := h.conns[userID]
	if !ok {
		h.mu.Unlock()
		return
	}

	// copy conns before mutating map
	conns := make([]*websocket.Conn, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}

	// authoritative removal
	delete(h.conns, userID)
	h.mu.Unlock()

	// close sockets outside lock
	for _, c := range conns {
		_ = c.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.ClosePolicyViolation,
				"session expired",
			),
		)
		_ = c.Close()
	}

	// single authoritative presence update
	h.BroadcastPresence(userID, false)
}

// Heartbeat helpers (used by WS handler)
const (
	WriteWait  = 10 * time.Second
	PongWait   = 60 * time.Second
	PingPeriod = (PongWait * 9) / 10
)
