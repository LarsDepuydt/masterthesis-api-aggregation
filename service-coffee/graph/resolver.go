package graph

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// go:generate go run github.com/99designs/gqlgen generate
//
// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB *sql.DB
}

// ConnectDB establishes a database connection using environment variables.
func (r *Resolver) initCofeeDBConnection() error {
	host := os.Getenv("COFFEE_DB_HOST")
	dbname := os.Getenv("COFFEE_DB_NAME")
	user := os.Getenv("COFFEE_DB_USER")
	password := os.Getenv("COFFEE_DB_PASSWORD")
	port := os.Getenv("COFFEE_DB_PORT")

	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, dbname, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	r.DB = db
	log.Println("Successfully connected to the database")
	return nil
}
