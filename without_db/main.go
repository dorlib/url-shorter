package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type URLStore struct {
	sync.RWMutex
	store map[string]string
}

var urlStore = &URLStore{store: make(map[string]string)}

func generateShortID() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortID := generateShortID()
	urlStore.Lock()
	urlStore.store[shortID] = req.URL
	urlStore.Unlock()

	response := map[string]string{"shortenedURL": "http://localhost:8081/" + shortID}
	json.NewEncoder(w).Encode(response)
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortID := vars["shortID"]

	urlStore.RLock()
	originalURL, exists := urlStore.store[shortID]
	urlStore.RUnlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenURL).Methods("POST")
	r.HandleFunc("/{shortID}", redirectURL).Methods("GET")

	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
