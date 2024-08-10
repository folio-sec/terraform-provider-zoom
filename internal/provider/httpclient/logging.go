package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type LoggingRoundTripper struct {
	ctx context.Context
	rt  http.RoundTripper
}

func NewLoggingRoundTripper(ctx context.Context, rt http.RoundTripper) http.RoundTripper {
	return &LoggingRoundTripper{ctx: ctx, rt: rt}
}

func (t LoggingRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.rt.RoundTrip(req)
	if err != nil {
		tflog.Info(
			t.ctx,
			fmt.Sprintf("http err %s", err),
			map[string]interface{}{
				"method":      req.Method,
				"request_uri": req.URL.RequestURI(),
			},
		)
		return
	}
	if resp != nil {
		buf, _ := io.ReadAll(resp.Body)
		loggingBody := io.NopCloser(bytes.NewBuffer(buf))
		rawBody := io.NopCloser(bytes.NewBuffer(buf))
		resp.Body = rawBody

		bodyBytes, err := io.ReadAll(loggingBody)
		if err != nil {
			return nil, fmt.Errorf("http failed to read response body: %w", err)
		}
		bodyString := string(bodyBytes)
		if resp.StatusCode >= 400 {
			tflog.Info(t.ctx,
				fmt.Sprintf("http response %d", resp.StatusCode),
				map[string]interface{}{
					"method":      req.Method,
					"request_uri": req.URL.RequestURI(),
					"body":        bodyString,
				},
			)
		} else {
			tflog.Debug(t.ctx,
				fmt.Sprintf("http response %d", resp.StatusCode),
				map[string]interface{}{
					"method":      req.Method,
					"request_uri": req.URL.RequestURI(),
					"body":        bodyString,
				},
			)
		}
	}
	return
}
