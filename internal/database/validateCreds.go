package database

import (
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Check credentials from DB
func ValidateCredsByIdentifier(identifier, password string) (
	username, role, bio, avatar string,
	userID int,
	status string,
	err error,
) {
	var storedHash string

	identifier = strings.ToLower(strings.TrimSpace(identifier))

	err = DB.QueryRow(`
		SELECT password, username, role, bio, avatar, id, status
		FROM users
		WHERE email = ? OR username = ?
	`, identifier, identifier).
		Scan(&storedHash, &username, &role, &bio, &avatar, &userID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", "", 0, status, errors.New("user not found")
		}
		return "", "", "", "", 0, status, err
	}

	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) != nil {
		return "", "", "", "", 0, status, errors.New("invalid password")
	}

	return username, role, bio, avatar, userID, status, nil
}

func UpdateUserStatus(userID int, newStatus string) error {
	stmt, err := DB.Prepare("UPDATE users SET status = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newStatus, userID)
	return err
}
