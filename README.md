# reqctl
reqctl is a Go package that provides enhanced control over HTTP requests, including retry mechanisms and asynchronous execution.

## Features

* Simple and exponential retry strategies
* Custom retry checkers
* Request timeouts
* Asynchronous parallel requests
* Works with the standard http.Client

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
ctrl := reqctl.Request(ctx, req).
    SetSimpleRetry(time.Second, 3)
resp, err := ctrl.Do()
```

With Exponential Backoff
```go
ctrl := reqctl.Request(ctx, req).
    SetExponentialRetry(time.Second, 3)
resp, err := ctrl.Do()
```

With Custom Retry Checker
```go
customChecker := func(resp *http.Response, err error) bool {
    return err != nil || resp.StatusCode >= 500
}

ctrl := reqctl.Request(ctx, req).
    SetSimpleRetryWithChecker(time.Second, 3, customChecker)
resp, err := ctrl.Do()
```

With Timeout
```go
ctrl := reqctl.Request(ctx, req).
    SetTimeout(5 * time.Second)
resp, err := ctrl.Do()
```

Asynchronous Parallel Requests
```go
ctrl := reqctl.Request(ctx, req).
    SetDelayedParallelCall(2 * time.Second)
resp, err := ctrl.Do()
```

## License
This project is licensed under the Open Source MIT License - see the LICENSE file for details.



## Contributing
Feel free to contribute or request any feature. Reporting issues can also help.
