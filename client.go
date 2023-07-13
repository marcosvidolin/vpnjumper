package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/marcosvidolin/vpnjumper/message"
)

type Client struct {
	Addr        string
	RedisClient *redis.Client
	Logger      *log.Logger
	ForwardTo   string
}

func (c *Client) Run() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			c.Logger.Printf("failed to read the body: %v", err)
			return
		}
		defer r.Body.Close()

		req := message.Request{
			Method:        r.Method,
			URL:           r.URL,
			Header:        r.Header,
			Host:          c.ForwardTo,
			Body:          string(body),
			ContentLength: r.ContentLength,
		}

		reqJson, err := json.Marshal(req)
		if err != nil {
			c.Logger.Println(err)
		}
		c.publish("requests", string(reqJson))

		respCh := make(chan string)
		go c.subscribe("responses", respCh)
		for msg := range respCh {
			c.Logger.Println("response received:", msg)
			var resp message.Response
			if err := json.Unmarshal([]byte(msg), &resp); err != nil {
				c.Logger.Printf("failed to unmarshal response: %v", err)
				continue
			}

			r := message.Response{
				Status:     resp.Status,
				StatusCode: resp.StatusCode,
				Header:     resp.Header,
			}
			bresp, err := json.Marshal(r)
			if err != nil {
				c.Logger.Println(err)
			}

			c.Logger.Println(string(bresp))
			break
		}
	})

	return http.ListenAndServe(c.Addr, nil)
}

func (c *Client) subscribe(channel string, respCh chan<- string) {
	pubsub := c.RedisClient.Subscribe(channel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			c.Logger.Printf("failed to receive message: %v", err)
			continue
		}
		respCh <- msg.Payload
	}
}

func (c *Client) publish(channel, message string) {
	err := c.RedisClient.Publish(channel, message).Err()
	if err != nil {
		c.Logger.Printf("failed to publish message: %v", err)
	}
}
