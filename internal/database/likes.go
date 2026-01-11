package database

import "database/sql"

func InsertorUpdateReaction(userID int, targetType string, targetID int, isLike bool) error {
	var existing bool
	err := DB.QueryRow(`
		SELECT is_like FROM likes 
		WHERE user_id = ? AND target_type = ? AND target_id = ?
	`, userID, targetType, targetID).Scan(&existing)

	switch {
	case err == sql.ErrNoRows:
		// No existing reaction → insert new one
		_, err = DB.Exec(`
			INSERT INTO likes (user_id, target_type, target_id, is_like)
			VALUES (?, ?, ?, ?)
		`, userID, targetType, targetID, isLike)
		return err

	case err != nil:
		// Some other error
		return err

	case existing == isLike:
		// Same reaction clicked again → remove it (toggle off)
		_, err = DB.Exec(`
			DELETE FROM likes
			WHERE user_id = ? AND target_type = ? AND target_id = ?
		`, userID, targetType, targetID)
		return err

	default:
		// Different reaction → update
		_, err = DB.Exec(`
			UPDATE likes SET is_like = ?
			WHERE user_id = ? AND target_type = ? AND target_id = ?
		`, isLike, userID, targetType, targetID)
		return err
	}
}
