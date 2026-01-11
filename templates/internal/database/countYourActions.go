package database

import (
	"database/sql"
	"fmt"
	"forum/internal/models"
)

func GetUserStats(userID int) (models.UserStats, error) {
	var stats models.UserStats

	// Count posts
	err := DB.QueryRow(`SELECT COUNT(*) FROM posts WHERE user_id = ?`, userID).Scan(&stats.PostCount)
	if err != nil {
		return stats, err
	}

	// Count likes
	err = DB.QueryRow(`
		SELECT COUNT(*) FROM likes 
		WHERE user_id = ? AND target_type='post' AND is_like = 1
	`, userID).Scan(&stats.LikesGiven)
	if err != nil {
		return stats, err
	}

	// Count dislikes
	err = DB.QueryRow(`
		SELECT COUNT(*) FROM likes 
		WHERE user_id = ? AND target_type='post' AND is_like = 0
	`, userID).Scan(&stats.DislikesGiven)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func GetReactionStatsAndUserReaction(userID int, targetType string, targetID int) (likes int, dislikes int, userReaction *string, err error) {
	// 1. Get total likes and dislikes for the target
	err = DB.QueryRow(`
		SELECT 
			COUNT(CASE WHEN is_like THEN 1 END) AS likes,
			COUNT(CASE WHEN NOT is_like THEN 1 END) AS dislikes
		FROM likes
		WHERE target_type = ? AND target_id = ?
	`, targetType, targetID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to get likes/dislikes count: %w", err)
	}

	// 2. Get the user's reaction to this target (if any)
	var isLike sql.NullBool
	err = DB.QueryRow(`
		SELECT is_like FROM likes
		WHERE user_id = ? AND target_type = ? AND target_id = ?
	`, userID, targetType, targetID).Scan(&isLike)

	if err != nil {
		if err == sql.ErrNoRows {
			// User has not reacted
			return likes, dislikes, nil, nil
		}
		return 0, 0, nil, fmt.Errorf("failed to get user reaction: %w", err)
	}

	if isLike.Valid {
		if isLike.Bool {
			s := "like"
			userReaction = &s
		} else {
			s := "dislike"
			userReaction = &s
		}
	} else {
		userReaction = nil
	}

	return likes, dislikes, userReaction, nil
}

//test

func CountComments(post_id int) (NumComments int, err error) {
	err = DB.QueryRow(`
			SELECT 
				COUNT(*)
			FROM comments
			WHERE post_id = ?
		`, post_id).Scan(&NumComments)

	if err != nil {
		return 0, fmt.Errorf("failed to get likes/dislikes count: %w", err)
	}
	return NumComments, err
}
