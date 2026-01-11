package database

func SaveToken(uuid, tokenID, tokenType string, expiresAt int64) error {
	stmt := `INSERT INTO token_store (jti, uuid, token_type, expires_at) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(stmt, tokenID, uuid, tokenType, expiresAt)
	return err
}

func DeleteToken(tokenID string) error {
	stmt := `DELETE FROM token_store WHERE jti = ?`
	_, err := DB.Exec(stmt, tokenID)
	return err
}

func TokenExists(tokenID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM token_store WHERE jti = ?)`
	err := DB.QueryRow(query, tokenID).Scan(&exists)
	return exists, err
}

func ClearAllTokens() error {
	stmt := `DELETE FROM token_store`
	_, err := DB.Exec(stmt)

	if err != nil {
		return err
	}

	// SET EVERYONE INACTIVE
	upd := `UPDATE users SET status = 'inactive';`

	_, err = DB.Exec(upd)

	if err != nil {
		return err
	}

	return nil
}
