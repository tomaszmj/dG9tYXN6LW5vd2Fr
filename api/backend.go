package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

func DefaultBackend() Backend {
	return &defaultBackend{}
}

type defaultBackend struct {
	urls      sync.Map // map[uint64]NewUrl
	responses sync.Map // map[uint64]UrlResponse
	maxId     uint64
	urlsCount uint64
	mutex     sync.Mutex
}

func (d *defaultBackend) GetAllUrls(writer http.ResponseWriter) {
	returnedUrls := make([]ReturnedUrl, 0, d.urlsCount)
	d.urls.Range(func(key, value interface{}) bool {
		id := key.(uint64)
		url := value.(*NewUrl)
		returnedUrls = append(returnedUrls, ReturnedUrl{
			Id:          id,
			UrlAsString: url.Url.String(),
			Interval:    url.Interval,
		})
		return true
	})
	jsonData, err := json.Marshal(returnedUrls)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if _, err := writer.Write(jsonData); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (d *defaultBackend) GetFetcherHistory(writer http.ResponseWriter, urlId uint64) {
	if _, err := writer.Write([]byte(fmt.Sprintf("get history from id %d\n", urlId))); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (d *defaultBackend) PostNewUrl(writer http.ResponseWriter, requestData io.ReadCloser) {
	data, err := ioutil.ReadAll(requestData)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := requestData.Close(); err != nil { //TODO limit data size
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	var newUrl NewUrl
	if err := json.Unmarshal(data, &newUrl); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	newId := d.newId()
	d.urls.Store(newId, &newUrl)
	jsonData, err := json.Marshal(&UrlId{Id: newId})
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if _, err := writer.Write(jsonData); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (d *defaultBackend) DeleteUrl(writer http.ResponseWriter, urlId uint64) {
	if _, err := writer.Write([]byte(fmt.Sprintf("delete url id %d\n", urlId))); err != nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (d *defaultBackend) newId() uint64 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.urlsCount++
	d.maxId++
	return d.maxId
}
