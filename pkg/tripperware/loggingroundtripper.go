package tripperware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sethgrid/pester"
	"go.uber.org/zap"
)

type LoggingRoundTripper struct {
	next   http.RoundTripper
	logger *zap.Logger
}

func NewLoggingRoundTripper(
	next http.RoundTripper,
	logger *zap.Logger,
) *LoggingRoundTripper {
	return &LoggingRoundTripper{
		next:   next,
		logger: logger,
	}
}

func (rt *LoggingRoundTripper) RoundTrip(
	req *http.Request,
) (resp *http.Response, err error) {
	defer func(begin time.Time) {
		var msg string
		if resp != nil {
			msg = fmt.Sprintf(
				"method=%s host=%s path=%s status_code=%d took=%s\n",
				req.Method,
				req.URL.Host,
				req.URL.Path,
				resp.StatusCode,
				time.Since(begin),
			)
		} else {
			msg = fmt.Sprintf(
				"method=%s host=%s path=%s status_code=nil took=%s\n",
				req.Method, req.URL.Host, req.URL.Path, time.Since(begin),
			)
		}

		if err != nil {
			rt.logger.Error(
				msg,
				zap.Bool("performance", true),
				zap.Error(err),
			)
		} else {
			rt.logger.Debug(msg, zap.Bool("performance", true))
		}
	}(time.Now())

	return rt.next.RoundTrip(req)
}

func NewLoggedRetryHTTPClient(logger *zap.Logger) *pester.Client {
	pesterClient := pester.New()
	pesterClient.Backoff = pester.ExponentialJitterBackoff
	pesterClient.MaxRetries = 8
	pesterClient.KeepLog = false // Cannot both retain logs and have loghook
	pesterClient.LogHook = func(e pester.ErrEntry) {
		// e.Err nil when Retry on HTTP 429
		message := fmt.Sprintf("%d %s [%s] %s request-%d retry-%d",
			e.Time.Unix(), e.Method, e.Verb, e.URL, e.Request, e.Retry)
		if e.Err != nil {
			logger.Error(message, zap.Error(e.Err))
		} else {
			logger.Error(message)
		}
	}
	pesterClient.KeepLog = false
	rt := NewLoggingRoundTripper(http.DefaultTransport, logger)
	pesterClient.Transport = rt
	return pesterClient
}
