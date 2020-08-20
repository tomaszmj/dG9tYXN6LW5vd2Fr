package api

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Backend interface {
	GetAllUrls(writer http.ResponseWriter)
	GetFetcherHistory(writer http.ResponseWriter, urlId uint64)
	PostNewUrl(writer http.ResponseWriter, requestData io.ReadCloser)
	DeleteUrl(writer http.ResponseWriter, urlId uint64)
}

func CreateRoutes(r chi.Router, backend Backend) {
	r.Route("/api/fetcher", func(r chi.Router) {
		r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
			backend.GetAllUrls(writer)
		})
		r.Get("/{id}/history", func(writer http.ResponseWriter, request *http.Request) {
			id, err := getIdFromRequest(request)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			backend.GetFetcherHistory(writer, id)
		})
		r.Post("/", func(writer http.ResponseWriter, request *http.Request) {
			backend.PostNewUrl(writer, request.Body)
		})
		r.Delete("/{id}", func(writer http.ResponseWriter, request *http.Request) {
			id, err := getIdFromRequest(request)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			backend.DeleteUrl(writer, id)
		})
	})
}

func getIdFromRequest(request *http.Request) (uint64, error) {
	idStr := chi.URLParam(request, "id")
	idInt, err := strconv.ParseUint(idStr, 10, 64)
	return idInt, err
}
