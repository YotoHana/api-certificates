package main

import (
	h "api-certificates/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/check", h.Check)
	http.HandleFunc("/grants", h.Grants)
	http.HandleFunc("/grants/{id}", h.GrantsId)
	http.HandleFunc("/grants/{id}/filters", h.GrantsFilters)
	log.Println("Server started at localhost:8080")
	http.ListenAndServe(":8080", nil)
}