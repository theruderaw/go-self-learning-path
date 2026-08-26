package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten",handleShorten)
	mux.HandleFunc("GET /{id}",redirectHandle)
	mux.HandleFunc("GET /stats/{id}",statsHandle)

	fmt.Println("Server running on :8080")

	err := http.ListenAndServe(":8080",mux)
	if err != nil {
		fmt.Println(err)
	}


}
