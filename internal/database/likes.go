package database

import "database/sql"

type ReactionResult int

const (
    ReactionNone ReactionResult = iota // removed / dislike / no notif
    ReactionLiked                      // NEW like created
)

func InsertorUpdateReaction(
    userID int,
    targetType string,
    targetID int,
    isLike bool,
) (ReactionResult, error) {

    var existing bool
    err := DB.QueryRow(`
        SELECT is_like FROM likes
        WHERE user_id = ? AND target_type = ? AND target_id = ?
    `, userID, targetType, targetID).Scan(&existing)

    switch {
    case err == sql.ErrNoRows:
        // New reaction
        _, err = DB.Exec(`
            INSERT INTO likes (user_id, target_type, target_id, is_like)
            VALUES (?, ?, ?, ?)
        `, userID, targetType, targetID, isLike)

        if err != nil {
            return ReactionNone, err
        }
        if isLike {
            return ReactionLiked, nil
        }
        return ReactionNone, nil

    case err != nil:
        return ReactionNone, err

    case existing == isLike:
        // Toggle off
        _, err = DB.Exec(`
            DELETE FROM likes
            WHERE user_id = ? AND target_type = ? AND target_id = ?
        `, userID, targetType, targetID)
        return ReactionNone, err

    default:
        // Switch reaction
        _, err = DB.Exec(`
            UPDATE likes SET is_like = ?
            WHERE user_id = ? AND target_type = ? AND target_id = ?
        `, isLike, userID, targetType, targetID)

        if err != nil {
            return ReactionNone, err
        }
        if isLike {
            return ReactionLiked, nil
        }
        return ReactionNone, nil
    }
}
