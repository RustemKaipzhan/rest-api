package main

import (
	"log"
	"net/http"
	"rest-api/handlers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	println("Server starting on http://localhost:8080")
	http.HandleFunc("/api/archive/information", handlers.GetArchiveInformation)
	http.HandleFunc("/api/archive/files", handlers.CreateArchive)
	http.HandleFunc("/api/mail/file", handlers.SendFileToEmails)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
