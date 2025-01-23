package main

import (
	"fmt"
	"net/http"

	"server/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Setup CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://10.0.2.2:8000", "http://192.168.46.146:8000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	// Routes
	r.Get("/", handlers.HandleHome)      // Updated to use exported function
	r.Post("/chat", handlers.HandleChat) // Updated to use exported function

	fmt.Println("Server starting on port 8000...")
	http.ListenAndServe(":8000", r)
}
