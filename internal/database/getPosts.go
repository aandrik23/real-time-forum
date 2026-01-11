package database

import (
	"fmt"
	"forum/internal/models"
)

func GetPostByID(postID int) (models.Post, error) {
	var post models.Post

	// base post + author username
	err := DB.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`, postID).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt, &post.Author)
	if err != nil {
		return models.Post{}, err
	}

	// snippet (optional)
	if len(post.Content) > 100 {
		post.Snippet = post.Content[:100] + "..."
	} else {
		post.Snippet = post.Content
	}

	// categories
	cats, err := GetCategoriesByPostID(post.ID)
	if err != nil {
		return models.Post{}, err
	}
	post.Categories = cats

	// likes/dislikes (same approach you use elsewhere)
	err = DB.QueryRow(`
		SELECT 
			COUNT(CASE WHEN is_like THEN 1 END) AS likes,
			COUNT(CASE WHEN NOT is_like THEN 1 END) AS dislikes
		FROM likes
		WHERE target_type = 'post' AND target_id = ?
	`, post.ID).Scan(&post.Likes, &post.Dislikes)
	if err != nil {
		return models.Post{}, err
	}

	// comments
	comments, err := GetCommentsForPost(post.ID)
	if err != nil {
		return models.Post{}, err
	}
	post.Comments = comments
	post.NumComments = len(comments)

	return post, nil
}

func GetPostsByAuthorID(authorID int) ([]models.Post, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt, &username); err != nil {
			return nil, err
		}
		//Generate snippet
		if len(post.Content) > 100 {
			post.Snippet = post.Content[:100] + "..."
		} else {
			post.Snippet = post.Content
		}
		post.Author = username

		// Load categories
		categories, err := GetCategoriesByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		// Load reaction stats (likes/dislikes)
		likes, dislikes, _, err := GetReactionStatsAndUserReaction(post.AuthorID, "post", post.ID)
		if err != nil {
			return nil, err
		}
		post.Likes = likes
		post.Dislikes = dislikes

		// Load comments for the post
		comments, err := GetCommentsForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments

		// Optional: number of comments can be length of comments slice
		post.NumComments = len(comments)

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetLikedPostsByUserID(userID int, likeType string) ([]models.Post, error) {
	// Turn your string into the tinyint(1) flag the DB expects
	var isLikeFlag int
	switch likeType {
	case "like", "liked", "1":
		isLikeFlag = 1
	case "dislike", "disliked", "0":
		isLikeFlag = 0
	default:
		return nil, fmt.Errorf("invalid likeType %q; must be \"liked\" or \"disliked\"", likeType)
	}

	rows, err := DB.Query(`
SELECT p.id, p.user_id, p.title, p.content, p.created_at, u.username
	FROM posts p
	JOIN likes l ON p.id = l.target_id
	JOIN users u ON p.user_id = u.id
	WHERE l.user_id = ? 
	  AND l.target_type = 'post' 
	  AND l.is_like = ?
	ORDER BY p.created_at DESC`, userID, isLikeFlag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var username string
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt, &username); err != nil {
			return nil, err
		}
		//Generate snippet
		if len(post.Content) > 100 {
			post.Snippet = post.Content[:100] + "..."
		} else {
			post.Snippet = post.Content
		}
		post.Author = username

		// Load categories
		categories, err := GetCategoriesByPostID(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories

		// Load reaction stats (likes/dislikes)
		likes, dislikes, _, err := GetReactionStatsAndUserReaction(post.AuthorID, "post", post.ID)
		if err != nil {
			return nil, err
		}
		post.Likes = likes
		post.Dislikes = dislikes

		// Load comments for the post
		comments, err := GetCommentsForPost(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments

		// Optional: number of comments can be length of comments slice
		post.NumComments = len(comments)

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetAllPosts() ([]models.Post, error) {

	rows, err := DB.Query("SELECT id, user_id, content, created_at,title FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Content, &post.CreatedAt, &post.Title); err != nil {
			return nil, err
		}

		// Fetch author username using post.AuthorID
		err := DB.QueryRow("SELECT username FROM users WHERE id = ?", post.AuthorID).Scan(&post.Author)
		if err != nil {
			return nil, err
		}

		// Fetch categories for this post
		catRows, err := DB.Query(`
            SELECT c.id, c.name
            FROM categories c
            JOIN post_categories pc ON pc.category_id = c.id
            WHERE pc.post_id = ?
        `, post.ID)
		if err != nil {
			return nil, err
		}
		defer catRows.Close()

		var categories []models.Category
		for catRows.Next() {
			var cat models.Category
			if err := catRows.Scan(&cat.ID, &cat.Name); err != nil {
				return nil, err
			}
			categories = append(categories, cat)
		}
		catRows.Close()
		post.Categories = categories

		err = DB.QueryRow(`
			SELECT 
				COUNT(CASE WHEN is_like THEN 1 END) AS likes,
				COUNT(CASE WHEN NOT is_like THEN 1 END) AS dislikes
			FROM likes
			WHERE target_type = 'post' AND target_id = ?
		`, post.ID).Scan(&post.Likes, &post.Dislikes)

		if err != nil {
			return nil, err
		}

		err = DB.QueryRow(`
			SELECT 
				COUNT(*)
			FROM comments
			WHERE post_id = ?
		`, post.ID).Scan(&post.NumComments)

		if err != nil {
			return nil, err
		}

		//Generate snippet
		if len(post.Content) > 100 {
			post.Snippet = post.Content[:100] + "..."
		} else {
			post.Snippet = post.Content
		}
		post.Comments, err = GetCommentsForPost(post.ID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)

	}

	return posts, nil
}

func GetPostsByCategoryID(categoryID int) ([]models.Post, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, p.title, p.content, p.created_at
		FROM posts p
		JOIN post_categories pc ON p.id = pc.post_id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			return nil, err
		}
		// Optionally load categories, likes, comments...
		post.Categories, _ = GetCategoriesByPostID(post.ID)
		post.Comments, _ = GetCommentsForPost(post.ID)
		// Load reaction stats (likes/dislikes)
		likes, dislikes, _, err := GetReactionStatsAndUserReaction(post.AuthorID, "post", post.ID)
		if err != nil {
			return nil, err
		}
		post.Likes = likes
		post.Dislikes = dislikes
		post.NumComments = len(post.Comments)

		//Generate snippet
		if len(post.Content) > 100 {
			post.Snippet = post.Content[:100] + "..."
		} else {
			post.Snippet = post.Content
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func InsertPostWithCategories(post models.Post) (int, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO posts (user_id, title, content, created_at)
		VALUES (?, ?, ?, ?)`,
		post.AuthorID, post.Title, post.Content, post.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, category := range post.Categories {
		_, err = tx.Exec(
			"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			postID, category.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to link post to category ID %d: %w", category.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(postID), nil
}

func DeletePost(postID int) error {
	stmt := `DELETE FROM posts WHERE id = ?`
	_, err := DB.Exec(stmt, postID)
	return err
}
