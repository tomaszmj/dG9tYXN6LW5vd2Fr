package worker_test

//This is not a unit test and this test will take some time.
//It tests integration of worker and urls packages.
//Api is not tested here just because it is easier to operate
//directly on Go data structures instead of sending https requests and parsing json responses.
//Besides, api is well-tested on unit level (unlike worker, which has no tests)

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fetcher/api"
	"fetcher/urls"
	"fetcher/worker"
)

func TestWorkerAndUrlsIntergation(t *testing.T) {
	fmt.Println("running TestWorkerAndUrlsIntergration - it will take 7-8s because timeouts are tested")

	urlsBackend := urls.New(worker.New())

	http.HandleFunc("/ok", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		_, err := writer.Write([]byte("abcde"))
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/error", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	http.HandleFunc("/timeout", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(6 * time.Second)
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	server := httptest.NewServer(nil)

	t.Run("Create 100 new urls ...", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			u, err := url.Parse(server.URL + urlOfIteration(i))
			require.NoError(t, err)
			newUrl := api.NewUrl{Url: u, IntervalSeconds: 1}
			id, err := urlsBackend.PostNewUrl(newUrl)
			require.NoError(t, err)
			assert.Equal(t, id, api.UrlId{Id: uint64(i)})
		}
	})

	t.Run("Wait for 7s so that all fetchers are run", func(t *testing.T) {
		time.Sleep(7 * time.Second)
	})

	t.Run("Assert that valid history is returned for each url", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			responses, err := urlsBackend.GetFetcherHistory(uint64(i))
			assert.NoError(t, err)
			assert.Greater(t, len(responses), 0)
			for _, r := range responses {
				if i%3 == 0 {
					require.NotNil(t, r.Response)
					assert.Equal(t, "abcde", *r.Response)
				} else {
					assert.Nil(t, r.Response)
				}
			}
		}
	})

	t.Run("Delete first 50 urls", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			require.NoError(t, urlsBackend.DeleteUrl(uint64(i)))
		}
	})

	t.Run("Assert that GetAllUrls returns correct number of urls", func(t *testing.T) {
		urls, err := urlsBackend.GetAllUrls()
		assert.NoError(t, err)
		assert.Equal(t, 50, len(urls))
	})

	t.Run("Assert that valid history is returned for each url excluding deleted ones", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			_, err := urlsBackend.GetFetcherHistory(uint64(i))
			assert.Error(t, err)
		}
		for i := 50; i < 100; i++ {
			responses, err := urlsBackend.GetFetcherHistory(uint64(i))
			assert.NoError(t, err)
			assert.Greater(t, len(responses), 0)
			for _, r := range responses {
				if i%3 == 0 {
					require.NotNil(t, r.Response)
					assert.Equal(t, "abcde", *r.Response)
				} else {
					assert.Nil(t, r.Response)
				}
			}
		}
	})

	t.Run("Assert that urls were fetched according to interval 1s with 0.5s precision", func(t *testing.T) {
		for i := 50; i < 100; i++ {
			responses, err := urlsBackend.GetFetcherHistory(uint64(i))
			assert.NoError(t, err)
			if len(responses) > 1 {
				assert.LessOrEqual(t, math.Abs(1.0-responses[1].CreatedAt.Sub(responses[0].CreatedAt).Seconds()), 0.5)
			}
		}
	})

	t.Run("Create 50 another urls", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			u, err := url.Parse(server.URL + urlOfIteration(i))
			require.NoError(t, err)
			newUrl := api.NewUrl{Url: u, IntervalSeconds: 1}
			id, err := urlsBackend.PostNewUrl(newUrl)
			require.NoError(t, err)
			assert.Equal(t, id, api.UrlId{Id: uint64(i) + 100})
		}
	})

	t.Run("Assert that GetAllUrls returns correct number of urls (including new ones)", func(t *testing.T) {
		urls, err := urlsBackend.GetAllUrls()
		assert.NoError(t, err)
		assert.Equal(t, 100, len(urls))
	})
}

func urlOfIteration(iteration int) string {
	if iteration%3 == 0 {
		return "/ok"
	}
	if iteration%3 == 1 {
		return "/error"
	}
	return "/timeout"
}
