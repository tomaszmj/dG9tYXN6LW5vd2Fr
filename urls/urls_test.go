package urls_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fetcher/api"
	"fetcher/urls"
)

func TestUrls(t *testing.T) {
	worker := &fakeWorker{}
	urlsBackend := urls.New(worker)
	assert.NotNil(t, urlsBackend)
	t.Run("PostNewUrl assigns new id", func(t *testing.T) {
		u, err := url.Parse("https://httpbin.org/range/15")
		require.NoError(t, err)
		for i := 0; i < 5; i++ {
			newUrl := api.NewUrl{Url: u, Interval: 5 + i}
			id, err := urlsBackend.PostNewUrl(newUrl)
			require.NoError(t, err)
			assert.Equal(t, id, api.UrlId{Id: uint64(i)})
		}
	})
	t.Run("DeleteUrl returns error on non-existing id", func(t *testing.T) {
		assert.Error(t, urlsBackend.DeleteUrl(9))
	})
	t.Run("DeleteUrl returns no error on existing url", func(t *testing.T) {
		assert.NoError(t, urlsBackend.DeleteUrl(1))
	})
	t.Run("GetAllUrls returns all urlsBackend excluding deleted ones", func(t *testing.T) {
		listedUrls, err := urlsBackend.GetAllUrls()
		require.NoError(t, err)
		expectedUrls := []api.ReturnedUrl{
			{
				Id:          0,
				UrlAsString: "https://httpbin.org/range/15",
				Interval:    5,
			},
			{
				Id:          2,
				UrlAsString: "https://httpbin.org/range/15",
				Interval:    7,
			},
			{
				Id:          3,
				UrlAsString: "https://httpbin.org/range/15",
				Interval:    8,
			},
			{
				Id:          4,
				UrlAsString: "https://httpbin.org/range/15",
				Interval:    9,
			},
		}
		assert.Equal(t, expectedUrls, listedUrls)
	})
	t.Run("GetFetcherHistory returns error on non-existing url", func(t *testing.T) {
		_, err := urlsBackend.GetFetcherHistory(9)
		assert.Error(t, err)
	})
	t.Run("GetFetcherHistory returns fetcher history", func(t *testing.T) {
		history0, err := urlsBackend.GetFetcherHistory(0)
		assert.NoError(t, err)
		assert.Equal(t, []api.UrlResponse{}, history0)
		abcString := "abc"
		responses := []api.UrlResponse{
			{
				Response:  &abcString,
				Duration:  1,
				CreatedAt: time.Unix(1500000000, 0),
			},
			{
				Response:  nil,
				Duration:  5,
				CreatedAt: time.Unix(1500000006, 0),
			},
		}
		worker.Fetch(0, responses[0])
		worker.Fetch(0, responses[1])
		history1, err := urlsBackend.GetFetcherHistory(0)
		assert.Equal(t, responses, history1)
	})
}

type fakeWorker struct {
	handlers []func(response api.UrlResponse)
}

func (f *fakeWorker) NewFetchRoutine(newUrl api.NewUrl, onFetch func(response api.UrlResponse), stopChan chan struct{}) {
	// normally it should create fetcher goroutine - here we just emulate fetching in "Fetch" method in the same goroutine
	f.handlers = append(f.handlers, onFetch)
}

func (f *fakeWorker) Fetch(handlerIndex int, response api.UrlResponse) {
	f.handlers[handlerIndex](response)
}
