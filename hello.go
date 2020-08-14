package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello"))
		panicOnError(err)
	})
	panicOnError(http.ListenAndServe(":8080", r))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
