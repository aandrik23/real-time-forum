package database

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

// -------------------------
// Types
// -------------------------

type DMThread struct {
	OtherUserID       int    `json:"other_user_id"`
	OtherUsername     string `json:"other_username"`
	OtherAvatar       string `json:"other_avatar"`
	LastMessageBody   string `json:"last_message_body"`
	LastMessageAt     int64  `json:"last_message_at"`
	LastMessageSender int    `json:"last_message_sender"`
}

type DMMessage struct {
	ID             int    `json:"id"`
	SenderID       int    `json:"sender_id"`
	SenderUsername string `json:"sender_username"`
	Body           string `json:"body"`
	CreatedAt      int64  `json:"created_at"`
}

// Suggested users when no threads exist
type DMUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
}

// -------------------------
// Helpers
// -------------------------
func MarkDMDelivered(msgID int, deliveredAt int64) error {
	_, err := DB.Exec(
		`UPDATE messages SET delivered_at = ? WHERE id = ? AND delivered_at IS NULL`,
		deliveredAt, msgID,
	)
	return err
}

func UpdateLastRead(convID int, msgID int) error {
	_, err := DB.Exec(`
		UPDATE conversations
		SET last_read_msg_id = ?
		WHERE id = ?
	`, msgID, convID)
	return err
}

func normalizePair(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func UserExists(userID int) (bool, error) {
	var exists bool
	err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, userID).Scan(&exists)
	return exists, err
}

// -------------------------
// Delivery Status
// -------------------------

type UndeliveredDM struct {
	MsgID    int
	SenderID int
}

func GetUndeliveredDMsForUser(userID int) ([]UndeliveredDM, error) {
	rows, err := DB.Query(`
		SELECT m.id, m.sender_id
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE m.delivered_at IS NULL
		  AND m.sender_id != ?
		  AND (c.user1_id = ? OR c.user2_id = ?)
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UndeliveredDM
	for rows.Next() {
		var r UndeliveredDM
		if err := rows.Scan(&r.MsgID, &r.SenderID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// -------------------------
// Conversation + message ops
// -------------------------

func GetOrCreateConversation(userA, userB int) (int, bool, error) {
	u1, u2 := normalizePair(userA, userB)

	// Try existing
	var convID int
	err := DB.QueryRow(
		`SELECT id FROM conversations WHERE user1_id = ? AND user2_id = ?`,
		u1, u2,
	).Scan(&convID)

	if err == nil {
		return convID, false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	// Create new
	res, err := DB.Exec(
		`INSERT INTO conversations (user1_id, user2_id, last_message_id, last_message_at) VALUES (?, ?, NULL, NULL)`,
		u1, u2,
	)
	if err != nil {
		// race safety: if someone else created it, fetch it
		var id2 int
		err2 := DB.QueryRow(
			`SELECT id FROM conversations WHERE user1_id = ? AND user2_id = ?`,
			u1, u2,
		).Scan(&id2)
		if err2 == nil {
			return id2, false, nil
		}
		return 0, false, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, false, err
	}
	return int(id), true, nil
}

func InsertDM(conversationID, senderID int, body string) (msgID int, createdAt int64, err error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return 0, 0, errors.New("empty message")
	}

	createdAt = time.Now().Unix()
	res, err := DB.Exec(
		`INSERT INTO messages (conversation_id, sender_id, body, created_at) VALUES (?, ?, ?, ?)`,
		conversationID, senderID, body, createdAt,
	)
	if err != nil {
		return 0, 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	return int(id), createdAt, nil
}

func UpdateConversationLast(conversationID, lastMsgID int, lastMsgAt int64) error {
	_, err := DB.Exec(
		`UPDATE conversations SET last_message_id = ?, last_message_at = ? WHERE id = ?`,
		lastMsgID, lastMsgAt, conversationID,
	)
	return err
}

// -------------------------
// Threads + message loading
// -------------------------

func GetDMThreads(userID int) ([]DMThread, error) {
	// Returns only existing conversations, sorted by last_message_at desc (NULLS last)
	rows, err := DB.Query(`
		SELECT
			CASE
				WHEN c.user1_id = ? THEN c.user2_id
				ELSE c.user1_id
			END AS other_user_id,
			u.username,
			COALESCE(u.avatar, '') AS avatar,
			COALESCE(m.body, '') AS last_body,
			COALESCE(c.last_message_at, 0) AS last_at,
			COALESCE(m.sender_id, 0) AS last_sender
		FROM conversations c
		JOIN users u ON u.id = (
			CASE
				WHEN c.user1_id = ? THEN c.user2_id
				ELSE c.user1_id
			END
		)
		LEFT JOIN messages m ON m.id = c.last_message_id
		WHERE (c.user1_id = ? OR c.user2_id = ?)
  			AND c.last_message_at IS NOT NULL
			  ORDER BY c.last_message_at DESC, c.last_message_id DESC
		`, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DMThread
	for rows.Next() {
		var t DMThread
		if err := rows.Scan(
			&t.OtherUserID,
			&t.OtherUsername,
			&t.OtherAvatar,
			&t.LastMessageBody,
			&t.LastMessageAt,
			&t.LastMessageSender,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func GetSuggestedDMUsers(selfUserID int) ([]DMUser, error) {
	rows, err := DB.Query(`
		SELECT id, username, COALESCE(avatar, ''), status
		FROM users
		WHERE id != ?
		ORDER BY username COLLATE NOCASE ASC
	`, selfUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DMUser
	for rows.Next() {
		var u DMUser
		if err := rows.Scan(&u.UserID, &u.Username, &u.Avatar, &u.Status); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func GetDMMessagesWithUser(selfUserID, otherUserID int, beforeID int, limit int) ([]DMMessage, int, int, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Validate other exists
	exists, err := UserExists(otherUserID)

	if err != nil {
		return nil, 0, 0, err
	}

	if !exists {
		return nil, 0, 0, errors.New("recipient not found")
	}

	convID, ok, err := GetConversationIDIfExists(selfUserID, otherUserID)
	if err != nil {
		return nil, 0, 0, err
	}

	if !ok {
		return []DMMessage{}, 0, 0, nil
	}

	var lastRead int
	_ = DB.QueryRow(
		`SELECT last_read_msg_id FROM conversations WHERE id = ?`,
		convID,
	).Scan(&lastRead)

	// Base query: newest first, then frontend can reverse if it wants
	q := `
		SELECT m.id, m.sender_id, u.username, m.body, m.created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = ?
	`
	args := []any{convID}

	if beforeID > 0 {
		q += ` AND m.id < ? `
		args = append(args, beforeID)
	}

	q += ` ORDER BY m.id DESC LIMIT ? `
	args = append(args, limit)

	rows, err := DB.Query(q, args...)
	if err != nil {
		return nil, 0, lastRead, err
	}
	defer rows.Close()

	var out []DMMessage
	for rows.Next() {
		var m DMMessage
		if err := rows.Scan(&m.ID, &m.SenderID, &m.SenderUsername, &m.Body, &m.CreatedAt); err != nil {
			return nil, 0, lastRead, err
		}
		out = append(out, m)
	}
	// reverse to oldest -> newest
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return out, convID, lastRead, rows.Err()
}

func GetConversationIDIfExists(userA, userB int) (int, bool, error) {
	u1, u2 := normalizePair(userA, userB)

	var convID int
	err := DB.QueryRow(
		`SELECT id FROM conversations WHERE user1_id = ? AND user2_id = ?`,
		u1, u2,
	).Scan(&convID)

	if err == nil {
		return convID, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, err
}

func GetDMPartnerIDs(userID int) ([]int, error) {
	rows, err := DB.Query(`
		SELECT
			CASE
				WHEN user1_id = ? THEN user2_id
				ELSE user1_id
			END AS other_id
		FROM conversations
		WHERE (user1_id = ? OR user2_id = ?)
		  AND last_message_at IS NOT NULL
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
