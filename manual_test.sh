set -e
set -x


echo 'create urls which will fetch "abcdefghijklmno"'
curl -si 127.0.0.1:8080/api/fetcher -X POST -d '{"url":"https://httpbin.org/range/15","interval":4}'
curl -si 127.0.0.1:8080/api/fetcher -X POST -d '{"url":"https://httpbin.org/range/15","interval":2}'

echo 'get all urls (rerun it later with more urls)'
curl -si 127.0.0.1:8080/api/fetcher

echo 'create url which will timeout'
curl -si 127.0.0.1:8080/api/fetcher -X POST -d '{"url":"https://httpbin.org/delay/10","interval":5}'

echo 'create url which will be unreachable'
curl -si 127.0.0.1:8080/api/fetcher -X POST -d '{"url":"http://nonexisting-url.com","interval":6}'

echo 'delete first url'
curl -s 127.0.0.1:8080/api/fetcher/0 -X DELETE

echo 'try to delete nonexisiting url'
curl -s 127.0.0.1:8080/api/fetcher/0 -X DELETE

echo 'get history (try it also with other urls)'
curl -s 127.0.0.1:8080/api/fetcher/1/history
