package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is our database connection pool
// A "pool" manages multiple connections efficiently - you don't open/close for each query
var DB *pgxpool.Pool

// ConnectDB establishes the database connection
func ConnectDB() {
	// Get the connection string from environment variable
	// This is why we created the .env file!
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Create a connection pool
	var err error
	DB, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Test the connection
	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	log.Println("Connected to database successfully!")
}

// CloseDB cleanly shuts down the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
