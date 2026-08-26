package main

import "sync"

type URL struct {
	ID			string
	OriginalURL	string
	AccessCount	int
}

var (
	urls = make(map[string]URL)
	mu    sync.RWMutex
)
