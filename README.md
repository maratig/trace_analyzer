# Trace analyzer

Package `trace_analyzer` is a fast light-weight utility for analyzing Go applications. It can collect and process `runtime/trace` metrics either from the corresponding endpoint or from file.
At any point of time one can get application statistics, for example a top 10 most idling goroutines which potentially can be a sign of leaking goroutines.

## Usage

### 1. As a standalone application
`trace_analyzer` can be used as a standalone application controlled by requests to its RESTful API endpoint. To do that `trace_analyzer` should be compiled and run. By default it listens port 8080
#### Example 1: listen an HTTP endpoint streaming trace events

_Request:_

```curl -X POST <analyzer_host>:<analyzer_port>/trace-events/listen -d source_path=http://example.com/debug/pprof/trace```

_Response:_

```{"id": 0}```

The `id` above is a unique identifier of a process listening trace events. You can use that `id` to get a summary about collected data.

_Note: instead of URL you can provide path to a file containing trace events_

#### Example 2: get a top 10 idling gorotines using the `id` above ####
_Request_:
```curl -X GET <analyzer_host>:<analyzer_host>/trace-events/top-idling-goroutines?id=0```

_Response:_

```
[
  {
    "id": 3222, <-- Goroutine ID
    "parent-stack": "\tnet/http.(*Transport).dialConn @ 0x741e84\n\t\t/usr/local/go/src/net/http/transport.go:1800\n\tnet/http.(*Transport).dialConnFor @ 0x73fb2c\n\t\t/usr/local/go/src/net/http/transport.go:1485\n",
    "stack": "\tnet/http.(*persistConn).writeLoop @ 0x745d40\n\t\t/usr/local/go/src/net/http/transport.go:2441\n",
    "execution-duration": 0,
    "live-duration": 2099000246336
  }
]
```

### 2. As an external package in your application ###

In order to use `trace_analyzer` as a package you need to download it

```go get github.com/maratig/trace_analyze```

And create an instance of App from `github.com/maratig/trace_analyzer/app` package. It has useful methods to listen trace event and get statistics similar to what RESTful API provides