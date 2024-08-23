package reqctl_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RohanPoojary/reqctl"
)

func Example_retry() {
	// Create a new request
	request, err := http.NewRequest("GET", "https://httpbin.org/status/200", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create a new request controller
	ctlr := reqctl.Request(ctx, request).
		SetExponentialRetry(100*time.Millisecond, 3)

	// Execute the request
	httpResp, err := ctlr.Do()
	if err != nil {
		fmt.Printf("Request failed: %v", err)
	}

	fmt.Printf("Response Code: %v", httpResp.StatusCode)
	// Output:
	//
	// Response Code: 200
}

func Example_timeout() {
	// Create a new request
	request, err := http.NewRequest("GET", "https://httpbin.org/delay/1", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create a new request controller
	ctlr := reqctl.Request(ctx, request).
		SetTimeout(100 * time.Millisecond)

	// Request should fail as the api takes 1 second to respond
	_, err = ctlr.Do()
	fmt.Println("Request failed:", err)

	// Output:
	//
	// Request failed: Get "https://httpbin.org/delay/1": context deadline exceeded
}

func Example_fastestFirst() {
	// Create a new request
	request, err := http.NewRequest("GET", "https://httpbin.org/status/200", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create a new request controller
	ctlr := reqctl.Request(ctx, request).
		SetParallelCallWithDelay(100 * time.Millisecond)

	// Request should respond with 200 as the fastest request is successful
	httpResp, err := ctlr.Do()
	if err != nil {
		fmt.Printf("Request failed: %v", err)
	}

	fmt.Printf("Response Code: %v", httpResp.StatusCode)
	// Output:
	//
	// Response Code: 200
}

func Example_advanced() {
	// Create a new request
	request, err := http.NewRequest("GET", "https://httpbin.org/status/200", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// A complex request handling, where every request shall have timeout of 100ms,
	// if the request fails, a timeout will be triggered with a delay of 50ms upto 3 times.
	// If the overall response time ( with retry ) takes more than 150ms,
	// a parallel call shall be fired with same timeout & retry policy.
	ctlr := reqctl.Request(ctx, request).
		SetTimeout(100*time.Millisecond).
		SetSimpleRetry(50*time.Millisecond, 3).
		SetParallelCallWithDelay(150 * time.Millisecond)

	// Request should respond with 200 as the fastest request is successful
	httpResp, err := ctlr.Do()
	if err != nil {
		fmt.Printf("Request failed: %v", err)
	}

	fmt.Printf("Response Code: %v", httpResp.StatusCode)
	// Output:
	//
	// Response Code: 200
}
