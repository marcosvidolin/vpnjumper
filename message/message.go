package message

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	Method        string      `json:"method"`
	URL           *url.URL    `json:"url"`
	Header        http.Header `json:"header"`
	Host          string      `json:"host"`
	Body          string      `json:"body"`
	ContentLength int64       `json:"contentLength"`
}

func (r Request) HttpRequest() *http.Request {
	b := io.NopCloser(bytes.NewBuffer([]byte(r.Body)))
	return &http.Request{
		Method:        r.Method,
		URL:           r.URL,
		Header:        r.Header,
		Body:          b,
		ContentLength: r.ContentLength,
		Host:          r.Host,
	}
}

type Response struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"statusCode"`
	Header     http.Header `json:"header"`
	Body       string      `json:"body"`
}

func (r Response) HttpResponse() *http.Response {
	b := io.NopCloser(bytes.NewBuffer([]byte(r.Body)))
	return &http.Response{
		Status:     r.Status,
		StatusCode: r.StatusCode,
		Header:     r.Header,
		Body:       b,
	}
}
