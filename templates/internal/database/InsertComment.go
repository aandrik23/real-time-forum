package database

import "forum/internal/models"

func InsertComment(userID int, postID int, content string) (models.Comments, error) {
	var c models.Comments

	err := DB.QueryRow(`
		INSERT INTO comments (post_id, user_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		RETURNING id, user_id, content, created_at
	`, postID, userID, content).Scan(
		&c.ID,
		&c.AuthorID,
		&c.Content,
		&c.CreatedAt,
	)

	if err != nil {
		return c, err
	}

	return c, nil
}
