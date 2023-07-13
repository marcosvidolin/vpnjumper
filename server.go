package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/go-redis/redis"
	"github.com/marcosvidolin/vpnjumper/message"
)

type Server struct {
	HttpClient  *http.Client
	RedisClient *redis.Client
	Logger      *log.Logger
}

func (f *Server) Run() error {
	respCh := make(chan string)

	go f.subscribe("requests", respCh)

	go func() {
		for msg := range respCh {
			f.Logger.Println("request received:", msg)
			var req message.Request
			if err := json.Unmarshal([]byte(msg), &req); err != nil {
				f.Logger.Printf("failed to unmarshal request: %v", err)
				continue
			}
			resp, err := f.processRequest(req.HttpRequest())
			if err != nil {
				f.Logger.Println(err)
				continue
			}

			r := message.Response{
				Status:     resp.Status,
				StatusCode: resp.StatusCode,
				Header:     resp.Header,
			}
			bresp, err := json.Marshal(r)
			if err != nil {
				f.Logger.Println(err)
				continue
			}

			f.Logger.Println(r)
			go f.publish("responses", string(bresp))
		}
	}()

	// Wait for a termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	return nil
}

func (f *Server) processRequest(r *http.Request) (*http.Response, error) {
	targetURL := fmt.Sprintf("%s%s", r.Host, r.URL)

	f.Logger.Printf("target: %s", targetURL)

	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: r.Method,
		URL:    target,
		Header: make(http.Header),
		Host:   r.Host,
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for _, cookie := range r.Cookies() {
		req.AddCookie(cookie)
	}

	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (f *Server) subscribe(channel string, respCh chan<- string) {
	pubsub := f.RedisClient.Subscribe(channel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			f.Logger.Printf("failed to receive message: %v", err)
			continue
		}
		respCh <- msg.Payload
	}
}

func (f *Server) publish(channel, message string) {
	err := f.RedisClient.Publish(channel, message).Err()
	if err != nil {
		f.Logger.Printf("failed to publish message: %v", err)
	}
}
