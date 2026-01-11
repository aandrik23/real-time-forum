package database

import (
	"forum/internal/models"
)

func GetCommentsForPost(postID int) ([]models.Comments, error) {
	rows, err := DB.Query(`
		SELECT 
			c.id, c.user_id, u.username, c.content, c.created_at,
			COUNT(CASE WHEN l.is_like THEN 1 END) AS likes,
			COUNT(CASE WHEN NOT l.is_like THEN 1 END) AS dislikes
		FROM comments c
		JOIN users u ON u.id = c.user_id
		LEFT JOIN likes l ON l.target_type = 'comment' AND l.target_id = c.id
		WHERE c.post_id = ?
		GROUP BY c.id
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comments
	for rows.Next() {
		var c models.Comments
		err := rows.Scan(&c.ID, &c.AuthorID, &c.Author, &c.Content, &c.CreatedAt, &c.Likes, &c.Dislikes)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}
