package main

import (
	"fetcher/api"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	api.SetRoutes(r)
	fmt.Println("Starting server on port 8080 ...")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println("Failed to start server on port 8080:", err)
		return
	}
}
