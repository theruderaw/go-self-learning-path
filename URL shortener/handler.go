package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
)

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

type ShortenRequest struct {
	URL string `json:"url"`
}

func generateID(length int) string {
	id := make([]byte, length)

	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}

	return string(id)
}

func HandleShorten(w http.ResponseWriter, r *http.Request) {
	id := generateID(8)

	var req ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	url := URL{
		ID: id,
		OriginalURL:req.URL,
	}

	urls[id] = url
	json.NewEncoder(w).Encode(url)
}

func RedirectHandle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	originalURL,ok := urls[id]

	if !ok {
		http.NotFound(w, r)
		return
	}
	originalURL.AccessCount++
	urls[id] = originalURL
	http.Redirect(w,r,originalURL.OriginalURL, http.StatusFound)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	originalURL,ok := urls[id]

	if !ok {
		http.NotFound(w, r)
		return
	}

	json.NewEncoder(w).Encode(originalURL)
	
}


