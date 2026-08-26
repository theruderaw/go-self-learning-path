package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
)

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

type ShortenRequest struct {
	URL string `json"url"`
}

func generateID(length int) string {
	id := make([]byte, length)

	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}

	return string(id)
}

func shortenhandler(w http.ResponseWriter, r *http.Request) {
	id := generateID(8)

	var req ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
	}

	url := URL{
		ID: id,
		OriginalURL:req.URL,
	}

	urls[id] = url
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}


