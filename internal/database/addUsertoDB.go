package database

import (
	"errors"
	"strings"
)

func AddUserToDb(
	username,
	firstName,
	lastName,
	age,
	gender,
	email,
	hashedPassword,
	status string,
) error {
	stmt := `
		INSERT INTO users
		(username, first_name, last_name, age, gender, email, password, role, bio, avatar, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := DB.Exec(
		stmt,
		username,
		firstName,
		lastName,
		age,
		gender,
		email,
		hashedPassword,
		"user",
		"",
		"",
		status,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return errors.New("username already exists")
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return errors.New("email already exists")
		}
		return err
	}

	return nil
}

func ChangeUserDataFromDb(oldUsername int, newUsername, newbio, newavatar string) error {
	stmt := `UPDATE users SET username = ?, bio = ?, avatar = ? WHERE id = ?	`
	_, err := DB.Exec(stmt, newUsername, newbio, newavatar, oldUsername)
	return err
}
