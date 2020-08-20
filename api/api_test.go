package api_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fetcher/api"
)

func TestApi(t *testing.T) {
	backend := &fakeBackend{}
	r := chi.NewRouter()
	api.Create(r, backend)
	server := httptest.NewServer(r)

	t.Run("GET on non-covered url returns status 404", func(t *testing.T) {
		response, err := http.Get(server.URL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})

	t.Run("GET on /api/fetcher triggers GetAllUrls", func(t *testing.T) {
		t.Run("with valid request returns status 200", func(t *testing.T) {
			response, err := http.Get(server.URL + "/api/fetcher")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			responseBytes, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, `[{"id":11,"url":"https://httpbin.org/range/15","interval":60}]`, string(responseBytes))
		})
		t.Run("with internal server error returns status 500", func(t *testing.T) {
			backend.SetInternalError()
			defer backend.UnsetInternalError()
			response, err := http.Get(server.URL + "/api/fetcher")
			require.NoError(t, err)
			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	})

	t.Run("GET on /api/fetcher/{id}/history triggers GetFetcherHistory", func(t *testing.T) {
		t.Run("with valid request returns status 200", func(t *testing.T) {
			response, err := http.Get(server.URL + "/api/fetcher/11/history")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			responseBytes, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, `[{"response":null,"duration":0.571,"created_at":1559034638}]`, string(responseBytes))
		})
		t.Run("with internal server error returns status 500", func(t *testing.T) {
			backend.SetInternalError()
			defer backend.UnsetInternalError()
			response, err := http.Get(server.URL + "/api/fetcher/11/history")
			require.NoError(t, err)
			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
		t.Run("with non-integer id returns status 404", func(t *testing.T) {
			response, err := http.Get(server.URL + "/api/fetcher/id/history")
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
		t.Run("with non-existing id returns status 404", func(t *testing.T) {
			response, err := http.Get(server.URL + "/api/fetcher/22/history")
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	})

	t.Run("POST on /api/fetcher triggers PostNewUrl", func(t *testing.T) {
		t.Run("with valid request returns status 200", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60}`)
			response, err := http.Post(server.URL+"/api/fetcher", "application/json", bytes.NewBuffer(data))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			responseBytes, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)
			assert.Equal(t, `{"id":11}`, string(responseBytes))
		})
		t.Run("with internal server error returns status 500", func(t *testing.T) {
			backend.SetInternalError()
			defer backend.UnsetInternalError()
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60}`)
			response, err := http.Post(server.URL+"/api/fetcher", "application/json", bytes.NewBuffer(data))
			require.NoError(t, err)
			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
		t.Run("with invalid json returns status 400", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60`)
			response, err := http.Post(server.URL+"/api/fetcher", "application/json", bytes.NewBuffer(data))
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
		})
		t.Run("with too long request body returns status 413", func(t *testing.T) {
			data := []byte(strings.Repeat(" ", api.MaxPostBodySize+1))
			response, err := http.Post(server.URL+"/api/fetcher", "application/json", bytes.NewBuffer(data))
			require.NoError(t, err)
			require.Equal(t, http.StatusRequestEntityTooLarge, response.StatusCode)
		})
	})

	t.Run("DELETE on /api/fetcher/{id} triggers DeleteUrl", func(t *testing.T) {
		t.Run("with valid request returns status 200", func(t *testing.T) {
			client := &http.Client{}
			request, err := http.NewRequest("DELETE", server.URL+"/api/fetcher/11", nil)
			response, err := client.Do(request)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
		})
		t.Run("with non-integer id returns status 404", func(t *testing.T) {
			client := &http.Client{}
			request, err := http.NewRequest("DELETE", server.URL+"/api/fetcher/i", nil)
			response, err := client.Do(request)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
		t.Run("with non-existing id returns status 404", func(t *testing.T) {
			client := &http.Client{}
			request, err := http.NewRequest("DELETE", server.URL+"/api/fetcher/22", nil)
			response, err := client.Do(request)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
		t.Run("with internal server error returns status 500", func(t *testing.T) {
			backend.SetInternalError()
			defer backend.UnsetInternalError()
			client := &http.Client{}
			request, err := http.NewRequest("DELETE", server.URL+"/api/fetcher/11", nil)
			response, err := client.Do(request)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	})

}

type fakeBackend struct {
	error error
}

func (f *fakeBackend) SetInternalError() {
	f.error = fmt.Errorf("fake error")
}

func (f *fakeBackend) UnsetInternalError() {
	f.error = nil
}

func (f *fakeBackend) GetAllUrls() ([]api.ReturnedUrl, error) {
	returnedUrls := []api.ReturnedUrl{
		{
			Id:          11,
			UrlAsString: "https://httpbin.org/range/15",
			Interval:    60,
		},
	}
	return returnedUrls, f.error
}

func (f *fakeBackend) GetFetcherHistory(urlId uint64) ([]api.UrlResponse, error) {
	urlResponses := []api.UrlResponse{
		{
			Response:  nil,
			Duration:  time.Duration(int64(0.571 * float64(time.Second))),
			CreatedAt: time.Unix(1559034638, 0),
		},
	}
	if urlId == 11 {
		return urlResponses, f.error
	} else {
		return []api.UrlResponse{}, fmt.Errorf(api.BackendErrorNotFound)
	}
}

func (f *fakeBackend) PostNewUrl(url api.NewUrl) (api.UrlId, error) {
	return api.UrlId{Id: 11}, f.error
}

func (f *fakeBackend) DeleteUrl(urlId uint64) error {
	if urlId == 11 {
		return f.error
	} else {
		return fmt.Errorf(api.BackendErrorNotFound)
	}
}
