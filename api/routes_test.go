package api_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"fetcher/api"
)

func TestCreateRoutes(t *testing.T) {
	backend := &fakeBackend{}
	r := chi.NewRouter()
	api.CreateRoutes(r, backend)
	server := httptest.NewServer(r)

	t.Run("GET on non-covered url returns error 404", func(t *testing.T) {
		backend.Reset()
		response, err := http.Get(server.URL)
		require.NoError(t, err)
		assert.Equal(t, response.StatusCode, http.StatusNotFound)
	})

	t.Run("GET on /api/fetcher triggers GetAllUrls", func(t *testing.T) {
		backend.Reset()
		response, err := http.Get(server.URL + "/api/fetcher")
		require.NoError(t, err)
		require.Equal(t, response.StatusCode, http.StatusOK)
		assert.True(t, backend.GetAllUrlsCalled)
	})

	t.Run("GET on /api/fetcher/{id}/history triggers GetFetcherHistory", func(t *testing.T) {
		backend.Reset()
		response, err := http.Get(server.URL + "/api/fetcher/11/history")
		require.NoError(t, err)
		require.Equal(t, response.StatusCode, http.StatusOK)
		assert.Equal(t, backend.GetFetcherHistoryUrlId, 11)
	})

	t.Run("GET on /api/fetcher/{id}/history with incorrect id returns error 404", func(t *testing.T) {
		backend.Reset()
		response, err := http.Get(server.URL + "/api/fetcher/9999999999999999999999999999999999999999999999999/history")
		require.NoError(t, err)
		assert.Equal(t, response.StatusCode, http.StatusNotFound)
	})

	t.Run("POST on /api/fetcher triggers PostNewUrl", func(t *testing.T) {
		backend.Reset()
		data := []byte(`{"url":"https://httpbin.org/range/15","interval":60}'`)
		response, err := http.Post(server.URL+"/api/fetcher", "application/json", bytes.NewBuffer(data))
		require.NoError(t, err)
		require.Equal(t, response.StatusCode, http.StatusOK)
		assert.Equal(t, backend.PostNewUrlRequestData, data)
	})

	t.Run("DELETE on /api/fetcher/{id} triggers DeleteUrl", func(t *testing.T) {
		backend.Reset()
		client := &http.Client{}
		request, err := http.NewRequest("DELETE", server.URL+"/api/fetcher/2", nil)
		response, err := client.Do(request)
		require.NoError(t, err)
		require.Equal(t, response.StatusCode, http.StatusOK)
		assert.Equal(t, backend.DeleteUrlUrlId, 2)
	})
}

type fakeBackend struct {
	GetAllUrlsCalled       bool
	GetFetcherHistoryUrlId int
	PostNewUrlRequestData  []byte
	DeleteUrlUrlId         int
	Error                  error
}

func (f *fakeBackend) Reset() {
	f.GetAllUrlsCalled = false
	f.GetFetcherHistoryUrlId = -1
	f.PostNewUrlRequestData = []byte{}
	f.DeleteUrlUrlId = -1
	f.Error = nil
}

func (f *fakeBackend) GetAllUrls(writer http.ResponseWriter) {
	f.GetAllUrlsCalled = true
}

func (f *fakeBackend) GetFetcherHistory(writer http.ResponseWriter, urlId int) {
	f.GetFetcherHistoryUrlId = urlId
}

func (f *fakeBackend) PostNewUrl(writer http.ResponseWriter, requestData io.ReadCloser) {
	f.PostNewUrlRequestData, f.Error = ioutil.ReadAll(requestData)
}

func (f *fakeBackend) DeleteUrl(writer http.ResponseWriter, urlId int) {
	f.DeleteUrlUrlId = urlId
}
