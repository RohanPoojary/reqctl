package reqctl

import (
	"context"
	"math"
	"net/http"
	"sync"
	"time"
)

// retryType defines the type of retry strategy
type retryType string

const (
	noRetry          = retryType("none")
	simpleRetry      = retryType("simple")
	exponentialRetry = retryType("exponential")
)

// RetryCheckFunc is a function type that determines if a retry should be attempted
type RetryCheckFunc func(*http.Response, error) bool

// retryConfig holds the configuration for retry attempts
type retryConfig struct {
	MaxCount       int
	RetryType      retryType
	RetryInterval  time.Duration
	RetryCheckFunc RetryCheckFunc
}

// asyncRetryConfig holds the configuration for asynchronous retry
type asyncRetryConfig struct {
	Delay time.Duration
}

// ctrl is the internal controller that maintains the state of the request
type ctrl struct {
	ctx    context.Context
	req    *http.Request
	config struct {
		retryCfg *retryConfig
		asyncCfg *asyncRetryConfig
		timeout  time.Duration
	}
}

// Request creates a new ctrl instance with the given context and request
func Request(ctx context.Context, req *http.Request) *ctrl {
	c := ctrl{
		ctx: ctx,
		req: req.Clone(ctx),
	}

	c.config.retryCfg = &retryConfig{
		RetryType: noRetry,
	}

	return &c
}

// SetSimpleRetry configures simple retry with default checker
func (c ctrl) SetSimpleRetry(interval time.Duration, times int) ctrl {
	return c.setRetryWithChecker(simpleRetry, interval, times, DefaultRetryChecker)
}

// SetSimpleRetryWithChecker configures simple retry with custom checker
func (c ctrl) SetSimpleRetryWithChecker(interval time.Duration, times int, checker RetryCheckFunc) ctrl {
	return c.setRetryWithChecker(simpleRetry, interval, times, checker)
}

// SetExponentialRetry configures exponential retry with default checker
func (c ctrl) SetExponentialRetry(interval time.Duration, times int) ctrl {
	return c.setRetryWithChecker(exponentialRetry, interval, times, DefaultRetryChecker)
}

// SetExponentialRetryWithChecker configures exponential retry with custom checker
func (c ctrl) SetExponentialRetryWithChecker(interval time.Duration, times int, checker RetryCheckFunc) ctrl {
	return c.setRetryWithChecker(exponentialRetry, interval, times, checker)
}

// setRetryWithChecker is a helper function to set retry configuration
func (c ctrl) setRetryWithChecker(rt retryType, interval time.Duration, times int, checker RetryCheckFunc) ctrl {
	cfg := retryConfig{
		RetryType:      rt,
		RetryInterval:  interval,
		MaxCount:       times,
		RetryCheckFunc: checker,
	}

	c.config.retryCfg = &cfg
	return c
}

// SetTimeout sets the timeout for the request
func (c ctrl) SetTimeout(timeout time.Duration) ctrl {
	c.config.timeout = timeout
	return c
}

// SetParallelCallWithDelay configures asynchronous retry
func (c ctrl) SetParallelCallWithDelay(delay time.Duration) ctrl {
	c.config.asyncCfg = &asyncRetryConfig{
		Delay: delay,
	}

	return c
}

// Do executes the request with the default HTTP client
func (c ctrl) Do() (*http.Response, error) {
	return c.do(http.DefaultClient)
}

// DoWithClient executes the request with the provided HTTP client
func (c *ctrl) DoWithClient(client *http.Client) (*http.Response, error) {
	return c.do(client)
}

// DefaultRetryChecker is the default retry function that retries on network errors
func DefaultRetryChecker(resp *http.Response, err error) bool {
	return err != nil
}

// Clone creates a deep copy of the ctrl instance
func (c *ctrl) Clone() ctrl {
	retryCfg := *c.config.retryCfg
	asyncCfg := *c.config.asyncCfg

	res := *c
	res.config.retryCfg = &retryCfg
	res.config.asyncCfg = &asyncCfg
	return res
}

// do is the main function that handles the request execution
func (c *ctrl) do(client *http.Client) (*http.Response, error) {
	if c.config.asyncCfg != nil {
		return c.doAsync(client)
	} else {
		return c.doRetry(client)
	}
}

// doAsync handles asynchronous retry
func (c *ctrl) doAsync(client *http.Client) (*http.Response, error) {
	var result *http.Response
	var resErr error

	once := sync.Once{}
	doneCh := make(chan struct{})

	aCtx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	runFunc := func(timeout time.Duration) {

		asyncCtrl := c.Clone()
		asyncCtrl.ctx = aCtx

		if timeout > 0 {
			select {
			// Either wait till one of the routine is closed or until timeout
			case <-doneCh:
				return
			case <-time.After(timeout):
			}
		}

		// Validate if the context is still active
		if aCtx.Err() == nil {
			res, err := asyncCtrl.doRetry(client)
			once.Do(func() {
				result = res
				resErr = err
				close(doneCh)
			})
		}
	}

	go runFunc(0)                       // The first request
	go runFunc(c.config.asyncCfg.Delay) // Delayed request

	<-doneCh

	return result, resErr
}

// doRequest executes a single HTTP request
func (c *ctrl) doRequest(client *http.Client) (*http.Response, error) {
	req := c.req.Clone(c.ctx)
	if c.config.timeout > 0 {
		ctx, cancel := context.WithTimeout(c.ctx, c.config.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	return client.Do(req)
}

// doRetry handles the retry logic
func (c *ctrl) doRetry(client *http.Client) (*http.Response, error) {
	retryCfg := c.config.retryCfg

	var resultErr error
	var resultResp *http.Response

	// Check if the first request succeeds
	if resultResp, resultErr = c.doRequest(client); retryCfg.RetryType == noRetry ||
		!retryCfg.RetryCheckFunc(resultResp, resultErr) {
		return resultResp, resultErr
	}

	// Initiate retry logic with delay
	for i := 0; i < retryCfg.MaxCount; i++ {

		// Calculate waiting duration for next execution
		var waitDuration time.Duration
		if retryCfg.RetryType == simpleRetry {
			waitDuration = retryCfg.RetryInterval
		} else if retryCfg.RetryType == exponentialRetry {
			waitDuration = retryCfg.RetryInterval * time.Duration(math.Exp2(float64(i)))
		}

		if waitDuration > 0 {
			time.Sleep(waitDuration)
		}

		if resultResp, resultErr = c.doRequest(client); !retryCfg.RetryCheckFunc(resultResp, resultErr) {
			break
		}
	}

	return resultResp, resultErr
}
