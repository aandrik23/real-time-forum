package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var Seed = false

const (
	databasePath = "./internal/database/migrations/"
	seedPath     = "./internal/database/seeds/"
)

func InitDB() {
	var err error

	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	initializeTables()

	// Restrict permissions for safety (read/write owner only)
	if err := os.Chmod("forum.db", 0777); err != nil {
		log.Printf("Warning: failed to set file permissions: %v", err)
	}

	if Seed == true {
		// Run seed file
		if err := execSQLFile(seedPath + "seedUsers.sql"); err != nil {
			log.Fatalf("Failed to execute seed.sql: %v", err)
		}
	}

	log.Println("Database initialized successfully.")
}

func initializeTables() {
	// Run schema files
	if err := execSQLFile(databasePath + "createUser.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createCookies.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createCategories.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}

	if err := execSQLFile(databasePath + "createComments.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createLikes.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createPost_Categories.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createPost.sql"); err != nil {
		log.Fatalf("Failed to execute user.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createConversations.sql"); err != nil {
		log.Fatalf("Failed to execute createConversations.sql: %v", err)
	}
	if err := execSQLFile(databasePath + "createMessages.sql"); err != nil {
		log.Fatalf("Failed to execute createMessages.sql: %v", err)
	}
	//seed categories
	if err := execSQLFile(seedPath + "seedCategories.sql"); err != nil {
		log.Fatalf("Failed to execute seed.sql: %v", err)
	}
}
