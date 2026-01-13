package handlers

import (
	"context"
	"database/sql"
	"time"
	 "encoding/json"
	"forum/internal/realtime"
	"github.com/gorilla/websocket"
	
)


type NotificationService struct {
	DB  *sql.DB
	Hub *realtime.NotificationHub
}

// CreateNotification saves to DB and pushes to user if online
func (s *NotificationService) CreateNotification(
	ctx context.Context,
	userID int64,
	actorID int64,
	typ string,
	targetID int64,
	payload interface{},
) error {

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Insert into DB
	var notifID int64
	err = s.DB.QueryRowContext(ctx, `
		INSERT INTO notifications (user_id, actor_id, type, target_id, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, actorID, typ, targetID, payloadJSON).Scan(&notifID)
	if err != nil {
		return err
	}

	// Build payload to send
	notifPayload := realtime.NotificationPayload{
		ID:       notifID,
		Type:     typ,
		ActorID:  actorID,
		TargetID: targetID,
		Payload:  payload,
		Read:     false,
		Created:  time.Now().Format(time.RFC3339),
	}

	// Send realtime if user is online
	if s.Hub.IsOnline(int(userID)) {
		s.Hub.SendToUser(int(userID), notifPayload)
	}

	return nil
}


func wsHandler(conn *websocket.Conn, userID int) {
    realtime.Notif.AddConn(userID, conn)
    defer realtime.Notif.RemoveConn(userID, conn)

    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            break
        }
    }
}

