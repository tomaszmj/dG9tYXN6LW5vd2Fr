package api_test

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fetcher/api"
)

// Tests cover only API use cases, for example NewUrl is used as response body in PostNewUrl, so there is only Unmarshal
func TestJsonMarshalling(t *testing.T) {
	t.Run("Marshal UrlId", func(t *testing.T) {
		bytes, err := json.Marshal(api.UrlId{Id: 2})
		require.NoError(t, err)
		assert.Equal(t, `{"id":2}`, string(bytes))
	})

	t.Run("Unmarshal NewUrl", func(t *testing.T) {
		t.Run("with valid json", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60}`)
			expectedUrl, err := url.Parse("https://httpbin.org/range/15")
			require.NoError(t, err)
			var newUrl api.NewUrl
			require.NoError(t, json.Unmarshal(data, &newUrl))
			assert.Equal(t, api.NewUrl{Url: expectedUrl, Interval: 60}, newUrl)
		})
		t.Run("with invalid json syntax", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
		t.Run("with invalid url", func(t *testing.T) {
			data := []byte(`{"url":"xx","interval":60}`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
		t.Run("with non-number interval", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":"60"}`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
		t.Run("with non-integer interval", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60.5}`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
		t.Run("with negative interval", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":-1}`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
		t.Run("with missing json key", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15"}`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
		t.Run("with unexpected json key", func(t *testing.T) {
			data := []byte(`{"url":"https://httpbin.org/range/15","interval":60,"key":"value"}`)
			var newUrl api.NewUrl
			assert.Error(t, json.Unmarshal(data, &newUrl))
		})
	})

	t.Run("Marshal ReturnedUrl", func(t *testing.T) {
		bytes, err := json.Marshal(api.ReturnedUrl{Id: 11, UrlAsString: "https://httpbin.org/range/15", Interval: 60})
		require.NoError(t, err)
		assert.Equal(t, `{"id":11,"url":"https://httpbin.org/range/15","interval":60}`, string(bytes))
	})

	t.Run("Marshal UrlResponse", func(t *testing.T) {
		t.Run("with non-empty response", func(t *testing.T) {
			responseStr := "abcd"
			urlResponse := api.UrlResponse{
				Response:  &responseStr,
				Duration:  time.Duration(int64(0.571 * float64(time.Second))),
				CreatedAt: time.Unix(1559034638, 0),
			}
			bytes, err := json.Marshal(&urlResponse)
			require.NoError(t, err)
			assert.Equal(t, `{"response":"abcd","duration":0.571,"created_at":1559034638}`, string(bytes))
		})
		t.Run("with empty response", func(t *testing.T) {
			urlResponse := api.UrlResponse{
				Response:  nil,
				Duration:  time.Duration(int64(0.571 * float64(time.Second))),
				CreatedAt: time.Unix(1559034638, 0),
			}
			bytes, err := json.Marshal(&urlResponse)
			require.NoError(t, err)
			assert.Equal(t, `{"response":null,"duration":0.571,"created_at":1559034638}`, string(bytes))
		})
	})
}
