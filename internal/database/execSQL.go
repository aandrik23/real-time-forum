package database

import (
	"fmt"
	"os"
	"strings"
)

func execSQLFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", filename, err)
	}

	queries := strings.Split(string(content), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("error executing query in %s: %w", filename, err)
		}
	}
	return nil
}
