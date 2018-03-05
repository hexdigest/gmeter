# GMeter [![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](https://github.com/hexdigest/gmeter/blob/master/LICENSE) [![Build Status](https://travis-ci.org/hexdigest/gmeter.svg?branch=master)](https://travis-ci.org/hexdigest/gmeter) [![Coverage Status](https://coveralls.io/repos/github/hexdigest/gmeter/badge.svg?branch=master)](https://coveralls.io/github/hexdigest/gmeter?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/hexdigest/gmeter)](https://goreportcard.com/report/github.com/hexdigest/gmeter) [![GoDoc](https://godoc.org/github.com/hexdigest/gmeter?status.svg)](http://godoc.org/github.com/hexdigest/gmeter)

A simple reverse proxy server that records HTTP requests and replays HTTP responses.
GMeter can be used for testing programs that make HTTP requests to third-party services.

## Installation

```
go get github.com/hexdigest/gmeter/cmd/gmeter
```

Alternatively you can download one of the pre-built executables:

* [gmeter-windows-i386-v1.0.zip](https://github.com/hexdigest/gmeter/files/1778923/gmeter-windows-i386-v1.0.zip)  
* [gmeter-windows-amd64-v1.0.zip](https://github.com/hexdigest/gmeter/files/1778924/gmeter-windows-amd64-v1.0.zip)
* [gmeter-linux-i386-v1.0.gz](https://github.com/hexdigest/gmeter/files/1778925/gmeter-linux-i386-v1.0.gz)
* [gmeter-linux-amd64-v1.0.gz](https://github.com/hexdigest/gmeter/files/1778926/gmeter-linux-amd64-v1.0.gz)
* [gmeter-darwin-amd64-v1.0.zip](https://github.com/hexdigest/gmeter/files/1778927/gmeter-darwin-amd64-v1.0.zip)

## Usage

```
  -d string
    	cassettes dir (default ".")
  -h	display this help text and exit
  -insecure
    	skip HTTPs checks
  -l string
    	listen address (default "localhost:8080")
  -t string
    	target base URL
```

Start gmeter:

```
$ gmeter -t http://github.com 
2018/03/05 00:06:19 started proxy localhost:8080 -> http://github.com
```

Put gmeter in recording mode:
```
$ curl -X POST http://localhost:8080/gmeter/record -d'{"cassette": "github_test"}'
```

Once you see something like:
```
2018/03/05 00:08:14 started recording of the cassette: github_test
```

Now you can start recording requests/responses. Lets request github index page:

```
$ curl http://localhost:8080/ -v
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 8080 (#0)
> GET / HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.47.0
> Accept: */*
> 
< HTTP/1.1 301 Moved Permanently
< Content-Length: 0
< Location: https://github.com/
< Date: Sun, 04 Mar 2018 21:11:25 GMT
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host localhost left intact
```

Both request and response now recorded and present in the github_test.cassette file that can be replayed.
In order to replay a cassette you have to put gmeter in play mode by making following request:

```
$ curl -X POST http://localhost:8080/gmeter/play -d'{"cassette": "github_test"}'
```

Now you can request github index page again and get a recorded response from the cassette.
