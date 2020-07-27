package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

func indexHandler(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Welcome! Please hit the '/qod' endpoint to get the quote of the day."))
}

func quoteOfTheDayHandler(client *redis.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		currentTime := time.Now()
		date := currentTime.Format("2006-01-02")

		val, err := client.Get(date).Result()
		if err == redis.Nil {
			log.Println("Cache miss for date", date)
			quoteRes, err := getQuoteFromAPI()
			if err != nil {
				res.Write([]byte("Sorry! We could not the Quote of the Day. Please try again later."))

				return
			}

			quote := quoteRes.Contents.Quotes[0].Quote
			client.Set(date, quote, 24*time.Hour)
			res.Write([]byte(quote))
		} else {
			log.Println("Cache hit for date", date)
			res.Write([]byte(val))
		}
	}
}

func getQuoteFromAPI() (*QuoteResponse, error) {
	API_URL := "http://quotes.rest/qod.json"
	res, err := http.Get(API_URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	log.Println("Quote API returned: ", res.StatusCode, http.StatusText(res.StatusCode))

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		quoteRes := &QuoteResponse{}
		json.NewDecoder(res.Body).Decode(quoteRes)

		return quoteRes, nil
	} else {
		return nil, errors.New("Could not get quote from API")
	}
}

func main() {
	// Create Redis Client
	var (
		host     = getEnv("REDIS_HOST", "localhost")
		port     = string(getEnv("REDIS_PORT", "6379"))
		password = getEnv("REDIS_PASSWORD", "")
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	// Create Server and Route Handlers
	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/qod", quoteOfTheDayHandler(client))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal
	<-interruptChan

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
