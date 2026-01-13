package database

import (
	"database/sql"
)

func GetPostAuthorID(postID int) (int, error) {
	var userID sql.NullInt64
	err := DB.QueryRow(`
		SELECT user_id
		FROM posts
		WHERE id = ?
	`, postID).Scan(&userID)
	if err != nil {
		return 0, err
	}

	if !userID.Valid {
		return 0, nil // post has no author (user_id NULL)
	}

	return int(userID.Int64), nil
}
func GetCommentAuthorID(commentID int) (int, error) {
	var userID int
	err := DB.QueryRow(`
		SELECT user_id
		FROM comments
		WHERE id = ?
	`, commentID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}



