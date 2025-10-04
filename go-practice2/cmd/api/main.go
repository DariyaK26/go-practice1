package main

import (
	"log"
	"net/http"

	"go-practice2/internal/handlers"
	"go-practice2/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/user", handlers.UserHandler)

	handlerWithMiddleware := middleware.AuthMiddleware(mux)

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", handlerWithMiddleware)
	if err != nil {
		log.Fatal(err)
	}
}
