package urls

import (
	"fmt"
	"net/url"
	"sort"
	"sync"

	"fetcher/api"
)

type Worker interface {
	NewFetchRoutine(newUrl api.NewUrl, onFetch func(response api.UrlResponse), stopChan chan struct{})
}

func New(w Worker) *Urls {
	u := &Urls{
		worker: w,
		urlMap: make(map[uint64]*urlData),
	}
	return u
}

type Urls struct {
	worker      Worker
	urlMap      map[uint64]*urlData
	urlMapMutex sync.RWMutex
	idManager   urlIdManager
}

type urlData struct {
	Url                *url.URL
	Interval           int
	Responses          []api.UrlResponse
	stopFetcherChannel chan struct{}
}

type urlIdManager struct {
	mutex sync.Mutex
	maxId uint64
}

// we can assume that ids will never overflow (with one id per microsecond, it would take ~599730 years to use 2^64 ids)
func (i *urlIdManager) NextId() uint64 {
	i.mutex.Lock()
	id := i.maxId
	i.maxId++
	i.mutex.Unlock()
	return id
}

func (u *Urls) GetAllUrls() ([]api.ReturnedUrl, error) {
	u.urlMapMutex.RLock()
	returnedUrls := make([]api.ReturnedUrl, 0, len(u.urlMap))
	for id, urlData := range u.urlMap {
		returnedUrls = append(returnedUrls, api.ReturnedUrl{
			Id:          id,
			UrlAsString: urlData.Url.String(),
			Interval:    urlData.Interval,
		})
	}
	u.urlMapMutex.RUnlock()
	sort.Slice(returnedUrls, func(i, j int) bool {
		return returnedUrls[i].Id < returnedUrls[j].Id
	})
	return returnedUrls, nil
}

func (u *Urls) GetFetcherHistory(urlId uint64) ([]api.UrlResponse, error) {
	u.urlMapMutex.RLock()
	urlData, ok := u.urlMap[urlId]
	if !ok {
		u.urlMapMutex.RUnlock()
		return []api.UrlResponse{}, fmt.Errorf(api.BackendErrorNotFound)
	}
	returnedResponses := make([]api.UrlResponse, 0, len(urlData.Responses))
	// deep-copy all responses to avoid races (urlData may be modified later)
	for _, response := range urlData.Responses {
		var responsePtr *string
		if response.Response == nil {
			responsePtr = nil
		} else {
			responseAsString := *response.Response
			responsePtr = &responseAsString
		}
		returnedResponses = append(returnedResponses, api.UrlResponse{
			Response:  responsePtr,
			Duration:  response.Duration,
			CreatedAt: response.CreatedAt,
		})
	}
	u.urlMapMutex.RUnlock()
	sort.Slice(returnedResponses, func(i, j int) bool {
		return returnedResponses[i].CreatedAt.Before(returnedResponses[j].CreatedAt)
	})
	return returnedResponses, nil
}

func (u *Urls) PostNewUrl(url api.NewUrl) (api.UrlId, error) {
	newId := u.idManager.NextId()
	newUrlMapEntry := &urlData{
		Url:                url.Url,
		Interval:           url.IntervalSeconds,
		Responses:          []api.UrlResponse{},
		stopFetcherChannel: make(chan struct{}, 1),
	}
	u.urlMapMutex.Lock()
	defer u.urlMapMutex.Unlock()
	u.urlMap[newId] = newUrlMapEntry
	onFetch := func(response api.UrlResponse) {
		u.urlMapMutex.Lock()
		defer u.urlMapMutex.Unlock()
		urlEntry, ok := u.urlMap[newId]
		if !ok {
			return // this may happen because stopFetcherChannel is buffered (DeleteUrl may exit before worker goroutine ends)
		}
		urlEntry.Responses = append(urlEntry.Responses, response)
	}
	u.worker.NewFetchRoutine(url, onFetch, newUrlMapEntry.stopFetcherChannel) // this should run worker in new goroutine
	return api.UrlId{Id: newId}, nil
}

func (u *Urls) DeleteUrl(urlId uint64) error {
	u.urlMapMutex.Lock()
	defer u.urlMapMutex.Unlock()
	deletedUrlData, ok := u.urlMap[urlId]
	if !ok {
		return fmt.Errorf(api.BackendErrorNotFound)
	}
	deletedUrlData.stopFetcherChannel <- struct{}{}
	delete(u.urlMap, urlId)
	return nil
}
