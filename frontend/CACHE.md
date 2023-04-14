# Initial request

GET xxx
Cache-control: no-cache
Pragma: no-cache
// Interesting: Sec-Fetch-*: ...

Response: 200 OK
Age: 1234
cache-control: public,max-age=3600
date: Thu, 13 Apr 2023 11:53:05 GMT
etag: "237e491b7c5faf7d3e7fa0b1c9cbf216"
last-modified: Wed, 12 Apr 2023 14:50:57 GMT


GET xxx
If-Modified-Since
	Wed, 12 Apr 2023 14:50:57 GMT
If-None-Match
	"237e491b7c5faf7d3e7fa0b1c9cbf216"

304 NOT MODIFIED
cache-control
	public,max-age=3600
date
	Thu, 13 Apr 2023 11:53:05 GMT
etag
	"237e491b7c5faf7d3e7fa0b1c9cbf216"



## Ideas from https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#use_cases

You can add a long max-age value and immutable because the content will never change.

```
# /assets/*
Cache-Control: max-age=31536000, immutable
```


for /index.html:
Cache-Control: no-cache





<IfModule mod_headers.c>
  <filesmatch "\.(ico|flv|jpg|jpeg|png|gif|css|swf)$">
  Header set Cache-Control "max-age=2678400, public"
  </filesmatch>
  <filesmatch "\.(html|htm)$">
  Header set Cache-Control "max-age=7200, private, must-revalidate"
  </filesmatch>
  <filesmatch "\.(pdf)$">
  Header set Cache-Control "max-age=86400, public"
  </filesmatch>
  <filesmatch "\.(js)$">
  Header set Cache-Control "max-age=2678400, private"
  </filesmatch>
</IfModule>




####

https://trains.jo-m.ch/data/blobs/train_20230413_113303.366_+01:00.jpg
https://trains.jo-m.ch/data/db.sqlite3?ts=1681390968283
