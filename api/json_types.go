package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// Returned by PostNewUrl
type UrlId struct {
	Id uint64 `json:"id"`
}

// Request body in PostNewUrl
type NewUrl struct {
	Url      *url.URL `json:"url"`
	Interval int      `json:"interval"`
}

// Returned by GetAllUrls
type ReturnedUrl struct {
	Id          uint64 `json:"id"`
	UrlAsString string `json:"url"`
	Interval    int    `json:"interval"`
}

// Returned by GetFetcherHistory
type UrlResponse struct {
	Response  *string       `json:"response"`
	Duration  time.Duration `json:"duration"`
	CreatedAt time.Time     `json:"created_at"`
}

func (n *NewUrl) UnmarshalJSON(j []byte) error {
	var rawData map[string]interface{}
	err := json.Unmarshal(j, &rawData)
	if err != nil {
		return err
	}
	if len(rawData) != 2 {
		return fmt.Errorf("expected exactly 2 keys: url, interval, got %d in json %s", len(rawData), j)
	}
	for key, value := range rawData {
		if key == "url" {
			urlStr, ok := value.(string)
			if !ok {
				return fmt.Errorf("unexpected value for key url (expected url as string, got %v as %T) in json %s", value, value, j)
			}
			n.Url, err = url.ParseRequestURI(urlStr)
			if err != nil {
				return err
			}
		} else if key == "interval" {
			intervalFloat, ok := value.(float64)
			if !ok {
				return fmt.Errorf("unexpected value for key interval (expected number, got %v as %T) in json %s", value, value, j)
			}
			n.Interval = int(intervalFloat)
			if float64(n.Interval) != intervalFloat {
				return fmt.Errorf("invalid interval in new url - must be positive integer, got %f", intervalFloat)
			}
			if n.Interval <= 0 {
				return fmt.Errorf("invalid interval in new url - must be positive integer, got %d", n.Interval)
			}
		}
	}
	return nil
}

func (u *UrlResponse) MarshalJSON() ([]byte, error) {
	base := struct {
		Response  *string `json:"response"`
		Duration  float64 `json:"duration"`
		CreatedAt int64   `json:"created_at"`
	}{
		Response:  u.Response,
		Duration:  u.Duration.Seconds(),
		CreatedAt: u.CreatedAt.Unix(),
	}
	return json.Marshal(base)
}
