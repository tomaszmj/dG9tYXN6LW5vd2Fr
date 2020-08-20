package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Backend interface {
	GetAllUrls() ([]ReturnedUrl, error)
	GetFetcherHistory(urlId uint64) ([]UrlResponse, error)
	PostNewUrl(url NewUrl) (UrlId, error)
	DeleteUrl(urlId uint64) error
}

const (
	BackendErrorNotFound = "not found"
	MaxPostBodySize      = 1000000
)

func Create(r chi.Router, backend Backend) {
	a := api{
		backend: backend,
	}
	r.Route("/api/fetcher", func(r chi.Router) {
		r.Get("/", a.handleGetAllUrls)
		r.Get("/{id}/history", a.handleGetFetcherHistory)
		r.Post("/", a.handlePostNewUrl)
		r.Delete("/{id}", a.handleDeleteUrl)
	})
}

type api struct {
	backend Backend
}

func (a *api) handleGetAllUrls(writer http.ResponseWriter, request *http.Request) {
	urls, err := a.backend.GetAllUrls()
	if err != nil {
		writeErrorInHttpResponse(writer, err)
		return
	}
	encodeJsonResponse(writer, urls)
}

func (a *api) handleGetFetcherHistory(writer http.ResponseWriter, request *http.Request) {
	id, err := getIdFromRequest(request)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	history, err := a.backend.GetFetcherHistory(id)
	if err != nil {
		writeErrorInHttpResponse(writer, err)
		return
	}
	encodeJsonResponse(writer, history)
}

func (a *api) handlePostNewUrl(writer http.ResponseWriter, request *http.Request) {
	limitedReader := io.LimitReader(request.Body, MaxPostBodySize)
	data, err := ioutil.ReadAll(limitedReader)
	if err != nil {
		writeErrorInHttpResponse(writer, err)
		return
	}
	if len(data) == MaxPostBodySize {
		http.Error(writer, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	var newUrl NewUrl
	if err := json.Unmarshal(data, &newUrl); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	newUrlId, err := a.backend.PostNewUrl(newUrl)
	if err != nil {
		writeErrorInHttpResponse(writer, err)
		return
	}
	encodeJsonResponse(writer, newUrlId)
}

func (a *api) handleDeleteUrl(writer http.ResponseWriter, request *http.Request) {
	id, err := getIdFromRequest(request)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err := a.backend.DeleteUrl(id); err != nil {
		writeErrorInHttpResponse(writer, err)
		return
	}
}

func getIdFromRequest(request *http.Request) (uint64, error) {
	idStr := chi.URLParam(request, "id")
	idInt, err := strconv.ParseUint(idStr, 10, 64)
	return idInt, err
}

func writeErrorInHttpResponse(writer http.ResponseWriter, err error) {
	if err.Error() == BackendErrorNotFound {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	} else {
		// Internal server errors in theory should not happen - handle it just as a sanity check
		http.Error(writer, fmt.Sprintf("Internal Server error: %s", err), http.StatusInternalServerError)
	}
}

func encodeJsonResponse(writer http.ResponseWriter, data interface{}) {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		writeErrorInHttpResponse(writer, err)
	}
}
