package main

import (
	"log"

	"github.com/jamalkaksouri/DigiOrder/internal/db"
	"github.com/jamalkaksouri/DigiOrder/internal/server" // Import the new server package
)

func main() {
	// Database connection is the only setup needed here.
	database, err := db.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Create and start the server.
	srv := server.New(database)
	log.Println("Starting server on :5582")
	if err := srv.Start(":5582"); err != nil {
		log.Fatal(err)
	}
}