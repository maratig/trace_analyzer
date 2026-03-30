# Trace analyzer

Trace analyzer is a fast light-weight tool for analyzing Go applications profiles produced by the `pprof` package. It can collect and process traces and heap profiles from a specified HTTP/HTTPS endpoint.
At any point of time one can get application statistics like:
- top 10 most idling goroutines
- heap profiles collected with specified time interval

> **Note:** `source_path` must be an HTTP or HTTPS URL

## Usage

### 1. As a standalone application
`trace_analyzer` can be used as a standalone application controlled by requests to its RESTful API endpoint. In order to do that `trace_analyzer` should be compiled and run. By default, it listens port 10000
#### Example 1: listen an HTTP endpoint streaming trace events

_Request:_

```curl -X POST <analyzer_host>:<analyzer_port>/trace-events/listen -d 'source_path=http://example.com/debug/pprof/trace'```

_Response:_

```{"id": 0}```

The `id` above is a unique identifier of a process listening trace events. You can use that `id` to get a summary about collected data.

If `source_path` is not a valid HTTP/HTTPS URL, the server responds with HTTP 400:

```{"error": invalid source path}```

#### Example 2: get top 10 idling goroutines using the `id` above ####
_Request_:
```curl <analyzer_host>:<analyzer_port>/trace-events/0/top-idling-goroutines```

_Response:_

```
[
  {
    "id": 3222, <-- Goroutine ID
    "transition-stack": "\tnet/http.(*Transport).dialConn @ 0x741e84\n\t\t/usr/local/go/src/net/http/transport.go:1800\n\tnet/http.(*Transport).dialConnFor @ 0x73fb2c\n\t\t/usr/local/go/src/net/http/transport.go:1485\n",
    "stack": "\tnet/http.(*persistConn).writeLoop @ 0x745d40\n\t\t/usr/local/go/src/net/http/transport.go:2441\n",
    "execution-duration": 0,
    "idle-duration": 304,
    "invoked-by": {
      "id": 322, <-- Goroutine ID
      "transition-stack": "\tnet/http.(*Transport).dialConn @ 0x741e84\n\t\t/usr/local/go/src/net/http/transport.go:1800\n\tnet/http.(*Transport).dialConnFor @ 0x73fb2c\n\t\t/usr/local/go/src/net/http/transport.go:1485\n",
      "stack": "\tnet/http.(*persistConn).writeLoop @ 0x745d40\n\t\t/usr/local/go/src/net/http/transport.go:2441\n",
      "execution-duration": 0,
      "idle-duration": 10,
    }
  }
]
```

If the connection to the source endpoint is still being established or has temporarily failed and is retrying, the response is:

```{"status": "processing"}```

#### Example 3: collecting heap profiles every 5 seconds
Request:

```curl -X POST <analyzer_host>:<analyzer_port>/heap-profiles/listen -d 'source_path=http://example.com/debug/pprof/heap```

Response:
```{"id": 0}```

The `id` above is a unique identifier of a process collecting heap profiles. You can use that `id` to get collected heap profiles.

#### Example 4: get collected heap profiles

Request:

```curl <analyzer_host>:<analyzer_port>/heap-profiles/0/profiles```

Response is a JSON-encoded string of `[][]*Profile` from `github.com/google/pprof/profile` package.

If profiles are still being collected (e.g. the endpoint is not yet reachable), the response is:

```{"status": "processing"}```

### 2. As an external package in your application ###

In order to use `trace_analyzer` as a package you need to download it

```go get github.com/maratig/trace_analyzer```

And create an instance of App from `github.com/maratig/trace_analyzer/app` package. It has useful methods to listen trace events and get statistics similar to what the RESTful API provides.

The `TopIdlingGoroutines` and `HeapProfilesSummary` methods return a status string as their second return value. When the value is `"processing"`, the data collection is still in progress (e.g. retrying a failed connection) and no results are available yet. Callers should poll again later.