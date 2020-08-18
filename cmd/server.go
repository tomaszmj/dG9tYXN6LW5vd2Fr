package main

import (
	"fetcher/api"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	api.CreateRoutes(r, &fakeBackend{})
	fmt.Println("Starting server on port 8080 ...")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println("Failed to start server on port 8080:", err)
		return
	}
}

type fakeBackend struct {
}

func (f *fakeBackend) GetAllUrls(writer http.ResponseWriter) {
	if _, err := writer.Write([]byte("get all urls\n")); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (f *fakeBackend) GetFetcherHistory(writer http.ResponseWriter, urlId int) {
	if _, err := writer.Write([]byte(fmt.Sprintf("get history from id %d\n", urlId))); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (f *fakeBackend) PostNewUrl(writer http.ResponseWriter, requestData io.ReadCloser) {
	defer requestData.Close()
	data, err := ioutil.ReadAll(requestData)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	if _, err := writer.Write([]byte(fmt.Sprintf("create: %s\n", data))); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (f *fakeBackend) DeleteUrl(writer http.ResponseWriter, urlId int) {
	if _, err := writer.Write([]byte(fmt.Sprintf("delete url id %d\n", urlId))); err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}
