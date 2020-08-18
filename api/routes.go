package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func SetRoutes(r chi.Router) {
	r.Route("/api/fetcher", func(r chi.Router) {
		r.Get("/", onGetAll)
		r.Get("/{id}/history", onGetHistory)
		r.Post("/", onPost)
		r.Delete("/{id}", onDelete)
	})
}

func onGetAll(writer http.ResponseWriter, request *http.Request) {
	if _, err := writer.Write([]byte("get all urls\n")); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
func onGetHistory(writer http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if _, err := writer.Write([]byte(fmt.Sprintf("get history from id %d\n", idInt))); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func onPost(writer http.ResponseWriter, request *http.Request) {
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	defer request.Body.Close()
	if _, err := writer.Write([]byte(fmt.Sprintf("create: %s\n", data))); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func onDelete(writer http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if _, err := writer.Write([]byte(fmt.Sprintf("delete url id %d\n", idInt))); err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}
