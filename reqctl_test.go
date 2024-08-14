package reqctl

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestSimpleRetry(t *testing.T) {
	failureURL := "https://httpbin.org/status/500"
	request, err := http.NewRequest("GET", failureURL, nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctlr := Request(ctx, request).
		SetSimpleRetry(10*time.Millisecond, 3)

	_, err = ctlr.Do()
	if err == nil {
		t.Errorf("Request should have failed due to retry")
	}
}

func TestRetryWithCustomFunc(t *testing.T) {
	failureURL := "https://httpbin.org/status/500"
	request, err := http.NewRequest("GET", failureURL, nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	execCount := 0
	customChecker := func(*http.Response, error) bool {
		// Returns success on 3rd retry
		execCount++
		return execCount < 3
	}

	_, err = Request(context.Background(), request).
		SetSimpleRetryWithChecker(100*time.Millisecond, 3, customChecker).
		Do()

	if err != nil {
		t.Errorf("Request should have succeeded after 3 retries %v", err)
	}
}

func TestExponentialRetry(t *testing.T) {
	failureURL := "https://httpbin.org/status/500"
	request, err := http.NewRequest("GET", failureURL, nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	customChecker := func(resp *http.Response, err error) bool {
		return resp.StatusCode == 500
	}

	start := time.Now()
	resp, err := Request(context.Background(), request).
		SetExponentialRetryWithChecker(1*time.Second, 3, customChecker).
		Do()

	if err != nil {
		t.Errorf("Shouldnt have failed via error: %v", err)
		return
	}

	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
		return
	}

	if time.Since(start) < 6*time.Second {
		t.Errorf("Expected atleast 6s of total delay, got %v", time.Since(start))
	}
}

func TestTimeout(t *testing.T) {
	delayURL := "https://httpbin.org/delay/1"
	request, err := http.NewRequest("GET", delayURL, nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	_, err = Request(request.Context(), request).
		SetTimeout(10 * time.Millisecond).
		Do()

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

}

func TestDelayedParallelCall(t *testing.T) {
	delayURL := "https://httpbin.org/status/200"
	request, err := http.NewRequest("GET", delayURL, nil)
	if err != nil {
		t.Errorf("Error creating request: %v", err)
		return
	}

	_, err = Request(request.Context(), request).
		SetDelayedParallelCall(100 * time.Millisecond).
		Do()

	if err != nil {
		t.Errorf("Obtained error: %v", err)
	}
}
