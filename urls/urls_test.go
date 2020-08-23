package urls_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fetcher/api"
	"fetcher/urls"
)

func TestUrls(t *testing.T) {
	urlsBackend := urls.New()
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
	t.Run("GetFetcherHistory returns no error on existing url", func(t *testing.T) {
		_, err := urlsBackend.GetFetcherHistory(0)
		assert.NoError(t, err)
	})
}
