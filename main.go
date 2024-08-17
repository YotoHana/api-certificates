package main

import (
	"api-certificates/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/check", handlers.Check)
	http.HandleFunc("/grants", handlers.Grants)
	log.Println("Server started at localhost:8080")
	http.ListenAndServe(":8080", nil)
}