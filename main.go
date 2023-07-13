package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

func main() {

	tp := flag.String("type", "client", "Type should be `server` or `client`")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	forwardTo := os.Getenv("FORWARD_TO")
	connectionStr := os.Getenv("REDIS_CONNECTION")

	redisClient := createRedisClient(connectionStr)

	switch *tp {
	case "client":
		c := Client{
			Addr:        port,
			RedisClient: redisClient,
			Logger:      log.New(os.Stdout, "[client] ", log.LstdFlags),
			ForwardTo:   forwardTo,
		}
		c.Run()
	case "server":
		s := Server{
			HttpClient: &http.Client{
				Timeout: 5 * time.Second,
			},
			RedisClient: redisClient,
			Logger:      log.New(os.Stdout, "[server] ", log.LstdFlags),
		}
		s.Run()
	default:
		log.Fatalln("invalid type")
	}

	if err := redisClient.Close(); err != nil {
		log.Fatalf("Error closing Redis client: %v", err)
	}
	log.Println("Redis client closed gracefully")
}

func createRedisClient(connStr string) *redis.Client {
	options, err := redis.ParseURL(connStr)
	if err != nil {
		log.Fatalf("Failed to parse Redis connection string: %v", err)
	}

	client := redis.NewClient(options)

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis:", pong)

	return client
}
