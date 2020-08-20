package urls

import (
	"fmt"
	"net/url"
	"sort"
	"sync"

	"fetcher/api"
)

func New() *Urls {
	return &Urls{}
}

type Urls struct {
	data      sync.Map // map[uint64]*urlData
	idManager urlIdManager
}

type urlData struct {
	Url       *url.URL
	Interval  int
	Responses []api.UrlResponse
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
	returnedUrls := make([]api.ReturnedUrl, 0)
	u.data.Range(func(key, value interface{}) bool {
		urlId := key.(uint64)
		urlVal := value.(*urlData)
		returnedUrls = append(returnedUrls, api.ReturnedUrl{
			Id:          urlId,
			UrlAsString: urlVal.Url.String(),
			Interval:    urlVal.Interval,
		})
		return true
	})
	sort.Slice(returnedUrls, func(i, j int) bool {
		return returnedUrls[i].Id < returnedUrls[j].Id
	})
	return returnedUrls, nil
}

func (u *Urls) GetFetcherHistory(urlId uint64) ([]api.UrlResponse, error) {
	value, ok := u.data.Load(urlId)
	if !ok {
		return []api.UrlResponse{}, fmt.Errorf(api.BackendErrorNotFound)
	}
	urlVal := value.(*urlData)
	return urlVal.Responses, nil
}

func (u *Urls) PostNewUrl(url api.NewUrl) (api.UrlId, error) {
	newId := u.idManager.NextId()
	newUrlData := &urlData{
		Url:       url.Url,
		Interval:  url.Interval,
		Responses: make([]api.UrlResponse, 0),
	}
	u.data.Store(newId, newUrlData)
	return api.UrlId{Id: newId}, nil
}

func (u *Urls) DeleteUrl(urlId uint64) error {
	_, ok := u.data.Load(urlId)
	if !ok {
		return fmt.Errorf(api.BackendErrorNotFound)
	}
	u.data.Delete(urlId)
	return nil
}
