package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"fetcher/api"
	"fetcher/urls"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	api.Create(r, urls.New(&fakeWorker{}))
	fmt.Println("Starting server on port 8080 ...")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println("Failed to start server on port 8080:", err)
		return
	}
}

type fakeWorker struct {
}

func (f *fakeWorker) NewFetchRoutine(newUrl api.NewUrl, onFetch func(response api.UrlResponse), stopChan chan struct{}) {
}
