package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"gitlab.unanet.io/devops/go/pkg/log"
)

// Transport implements http.RoundTripper. When set as Transport of http.Client, it executes HTTP requests with logging.
// No field is mandatory.
type Transport struct {
	Transport   http.RoundTripper
	LogRequest  func(req *http.Request)
	LogResponse func(resp *http.Response)
}

// THe default logging transport that wraps http.DefaultTransport.
var LoggingTransport = &Transport{
	Transport: http.DefaultTransport,
}

// Used if transport.LogRequest is not set.
var DefaultLogRequest = func(req *http.Request) {
	ctx := req.Context()
	fields := []zap.Field{
		zap.String("host", req.Host),
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
	}

	if reqID := log.GetReqID(ctx); len(reqID) > 0 {
		fields = append(fields, zap.String("req_id", reqID))
	}

	l := log.Logger
	if l.Core().Enabled(zap.DebugLevel) {
		buf, bodyErr := ioutil.ReadAll(req.Body)
		if bodyErr != nil {
			fields = append(fields, zap.Error(bodyErr))
		} else {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			fields = append(fields, log.DecodeBody(ioutil.NopCloser(bytes.NewBuffer(buf))))
			fields = append(fields, log.DecodeHeader(req.Header))
		}
	}

	l.Info("Outgoing HTTP Request", fields...)
}

// Used if transport.LogResponse is not set.
var DefaultLogResponse = func(resp *http.Response) {
	ctx := resp.Request.Context()
	l := log.Logger
	fields := []zap.Field{
		zap.Int("status", resp.StatusCode),
		zap.String("uri", resp.Request.URL.String()),
		zap.String("host", resp.Request.Host),
	}

	if reqID := log.GetReqID(ctx); len(reqID) > 0 {
		fields = append(fields, zap.String("req_id", reqID))
	}

	if l.Core().Enabled(zap.DebugLevel) {
		buf, bodyErr := ioutil.ReadAll(resp.Body)
		if bodyErr != nil {
			fields = append(fields, zap.Error(bodyErr))
		} else {
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			fields = append(fields, log.DecodeBody(ioutil.NopCloser(bytes.NewBuffer(buf))))
			fields = append(fields, log.DecodeHeader(resp.Header))
		}
	}

	l.Info("Incoming HTTP Response", fields...)
}

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP request with request/response logging.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	reqID := log.GetReqID(ctx)

	if len(reqID) == 0 {
		reqID = log.GetNextRequestID()
		req = req.WithContext(context.WithValue(ctx, log.RequestIDKey, reqID))
	}
	req.Header.Add(log.RequestIDHeader, reqID)

	t.logRequest(req)
	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp)

	return resp, err
}

func (t *Transport) logRequest(req *http.Request) {
	if t.LogRequest != nil {
		t.LogRequest(req)
	} else {
		DefaultLogRequest(req)
	}
}

func (t *Transport) logResponse(resp *http.Response) {
	if t.LogResponse != nil {
		t.LogResponse(resp)
	} else {
		DefaultLogResponse(resp)
	}
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}
