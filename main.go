package main

import (
	"log"
	"net/http"
	"rest-api/handlers"
)

func main() {
	println("Server starting on http://localhost:8080")
	http.HandleFunc("/api/archive/information", handlers.GetArchiveInformation)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
