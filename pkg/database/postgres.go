package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error open connecting: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	log.Println("Successfully connected to the database")
	return DB, nil
}

func CloseDB() error {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			return fmt.Errorf("error closing database connection: %w", err)
		}
		log.Println("Database connection closed")
	}
	return nil
}
