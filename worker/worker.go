package worker

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"fetcher/api"
)

type Worker struct {
}

const httpRequestTimeout = 5
const abortSendingResponseTimeout = 10

type urlResponseWithError struct {
	Response api.UrlResponse
	Error    error
}

func New() *Worker {
	return &Worker{}
}

func (w *Worker) NewFetchRoutine(url api.NewUrl, onFetch func(response api.UrlResponse), stopChan chan struct{}) {
	go func() {
		responseChan := make(chan urlResponseWithError)
		ticker := time.NewTicker(time.Duration(url.IntervalSeconds) * time.Second)
		for {
			select {
			case <-stopChan:
				ticker.Stop()
				return
			case response := <-responseChan:
				if response.Error != nil {
					log.Println(response.Error)
				} else {
					onFetch(response.Response)
				}
			case <-ticker.C:
				go makeRequestAndSaveResponse(url, responseChan)
			}
		}
	}()
}

func makeRequestAndSaveResponse(url api.NewUrl, responseChan chan urlResponseWithError) {
	response, err := makeHttpRequest(url.Url)
	// Set timeout on writing response to avoid leaking goroutines in case url gets deleted, i.e.
	// fetcher routine reads from stopChan and exits without reading response from this function
	timer := time.NewTimer(abortSendingResponseTimeout * time.Second)
	select {
	case responseChan <- urlResponseWithError{Response: response, Error: err}:
		timer.Stop()
		return
	case <-timer.C:
		return
	}
}

func makeHttpRequest(url *url.URL) (api.UrlResponse, error) {
	createdAt := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, httpRequestTimeout*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		e := fmt.Errorf("could not create request for url %s: %s", url.String(), err)
		return api.UrlResponse{}, e
	}
	t1 := time.Now()
	response, err := http.DefaultClient.Do(request)
	t2 := time.Now()
	duration := t2.Sub(t1)
	if err != nil {
		//errors.Is(err, context.DeadlineExceeded) or any other error - all are treated the same way
		return api.UrlResponse{Response: nil, Duration: duration, CreatedAt: createdAt}, nil
	}
	if response.StatusCode != http.StatusOK {
		return api.UrlResponse{Response: nil, Duration: duration, CreatedAt: createdAt}, nil
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return api.UrlResponse{Response: nil, Duration: duration, CreatedAt: createdAt}, nil
	}
	responseStr := string(bytes)
	return api.UrlResponse{Response: &responseStr, Duration: duration, CreatedAt: createdAt}, nil
}
