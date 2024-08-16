package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type NoJsonResponseRoundTripper struct {
	ctx context.Context
	rt  http.RoundTripper
}

func NewNoJsonResponseRoundTripper(ctx context.Context, rt http.RoundTripper) http.RoundTripper {
	return &NoJsonResponseRoundTripper{ctx: ctx, rt: rt}
}

func (t NoJsonResponseRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.rt.RoundTrip(req)
	if resp == nil {
		return
	}

	// zoom api sometimes return following response
	// - no Content-Type header
	// - empty response body
	// at that time, ogen cannot parse the response body as json
	// so we need to set Content-Type header to application/json and response body.
	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.Header.Get("Content-Length")
	if contentType == "" && contentLength == "0" {
		resp.Header.Set("Content-Type", "application/json")
		resp.Body = io.NopCloser(bytes.NewBufferString("{}"))
	}
	return
}
