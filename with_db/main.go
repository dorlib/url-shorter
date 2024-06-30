package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

type URLStore struct {
	client *redis.Client
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

var ctx = context.Background()

func NewURLStore() *URLStore {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	return &URLStore{client: client}
}

func (store *URLStore) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := fmt.Sprintf("%d", time.Now().UnixNano())
	err := store.client.Set(ctx, shortURL, req.URL, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{ShortURL: shortURL}
	json.NewEncoder(w).Encode(resp)
}

func (store *URLStore) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	longURL, err := store.client.Get(ctx, shortURL).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func main() {
	store := NewURLStore()
	r := mux.NewRouter()
	r.HandleFunc("/shorten", store.ShortenURL).Methods("POST")
	r.HandleFunc("/{shortURL}", store.Redirect).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", r))
}
