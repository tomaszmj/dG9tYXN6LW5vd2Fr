
# Simple http server with background url fetcher

## Building and testing
Run unit tests: ``go test ./api/... ./urls/...``  
Run integration test: ``go test -v -race worker/worker_integration_test.go``  
Build and run main program: ``cd cmd/server``, ``go build``, ``./server``
Run e2e tests manually with curl - while the server is running, execute some commands from the script ``./manual_test.sh``  

## API
#### Create new URL: POST/api/fetcher {"url":(string),"interval":(int)}
``$ curl -si 127.0.0.1:8080/api/fetcher -X POST -d '{"url":"https://httpbin.org/range/15","interval":4}``
``{ "id": 1 }``


#### Delete URL: DELETE /api/fetcher/(id)
``$ curl -s 127.0.0.1:8080/api/fetcher/0 -X DELETE``
Response: http 200 if url was deleted, http 404 if url did not exist


#### Get all urls: GET /api/fetcher
``$ curl -si 127.0.0.1:8080/api/fetcher``
```HTTP/1.1 200 OK
Date: Mon, 24 Aug 2020 05:31:09 GMT
Content-Length: 169
Content-Type: text/plain; charset=utf-8

[
  {
    "id": 0,
    "url": "https://httpbin.org/range/15",
    "interval": 4
  },
  {
    "id": 1,
    "url": "https://httpbin.org/range/15",
    "interval": 2
  }
]
```


#### Get fetched responses: GET/api/fetcher/(id)/history
``$ curl -s 127.0.0.1:8080/api/fetcher/1/history``
```
[
  {
    "response": "abcdefghijklmno",
    "duration": 1.994200221,
    "created_at": 1598247071
  },
  {
    "response": "abcdefghijklmno",
    "duration": 0.19730229,
    "created_at": 1598247073
  }
]
```
