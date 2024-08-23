# reqctl
reqctl is a Go package that provides enhanced control over HTTP requests, including retry mechanisms and asynchronous execution.

## Features

* Simple and exponential retry strategies
* Custom retry checkers
* Request timeouts
* Asynchronous parallel requests
* No third party dependencies

## Installation
To install the reqctl package, use go get:
```bash
go get github.com/RohanPoojary/reqctl
```

## Usage
Basic Request
```go
import (
    "context"
    "net/http"
    "github.com/yourusername/reqctl"
)

func main() {
    ctx := context.Background()
    req, _ := http.NewRequest("GET", "https://api.example.com", nil)
    
    ctrl := reqctl.Request(ctx, req)
    resp, err := ctrl.Do()
    if err != nil {
        // Handle error
    }
    defer resp.Body.Close()
    // Process response
}
```

With Retry
```go
// The request gets retried on failure with a delay of 10ms upto 3 times.
// After which it returns error.
//
// By default request failure is determined based on `error` received while sending request.
// You can add custom checker using `SetSimpleRetryWithChecker` instead of `SetSimpleRetry`
ctrl := reqctl.Request(ctx, req).
    SetSimpleRetry(10 * time.Millisecond, 3)
resp, err := ctrl.Do()
```

With Exponential Backoff
```go
// The request gets retried on failure with an exponential delay staring with 10ms upto 3 times.
// Hence the delay would be 10ms, 20ms & 40ms for 1st, 2nd & 3rd request respectively.
// After which it returns error.
//
// By default request failure is determined based on `error` received while sending request.
// You can add custom checker using `SetExponentialRetryWithChecker` instead of `SetExponentialRetry`
ctrl := reqctl.Request(ctx, req).
    SetExponentialRetry(10*time.Milliesecond, 3)
resp, err := ctrl.Do()
```

With Custom Retry Checker
```go
// Retry with a custom checker, which retries only if there are any errors in api call or api returns 5xx.
customChecker := func(resp *http.Response, err error) bool {
    return err != nil || resp.StatusCode >= 500
}

ctrl := reqctl.Request(ctx, req).
    SetSimpleRetryWithChecker(time.Second, 3, customChecker)
resp, err := ctrl.Do()
```

With Timeout
```go
// Every request shall have a timeout of 1s, after which it returns error.
ctrl := reqctl.Request(ctx, req).
    SetTimeout( time.Second)
resp, err := ctrl.Do()
```

Asynchronous Parallel Requests
```go
// A parallel request shall be initiated if no response is obtained in 100ms ( doesnt matter if its failure or successful ).
// The response & error of the fastest request between the two will be returned.
ctrl := reqctl.Request(ctx, req).
    SetParallelCallWithDelay(100 * time.Millisecond)
resp, err := ctrl.Do()
```

Complex Request
```go
// A complex request handling, where every request shall have timeout of 1s,
// if the request fails, a timeout will be triggered with a delay of 100ms upto 3 times.
// If the overall response time ( with retry ) takes more than 500ms,
// a parallel call shall be fired with same timeout & retry policy.
ctlr := reqctl.Request(context.TODO(), request).
    SetTimeout(time.Second).
    SetSimpleRetry(100*time.Millisecond, 3).
    SetParallelCallWithDelay(500 * time.Millisecond)

resp, err := ctrl.Do()
```

## TODO
- [ ] Support all request methods of default httpClient.
- [ ] Use `net/http/httptest` module for test cases.

## License
This project is licensed under the Open Source MIT License - see the LICENSE file for details.

## Contributing
Feel free to contribute or request any feature. Reporting issues can also help.
