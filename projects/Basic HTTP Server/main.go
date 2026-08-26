package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type AddPayload struct{
	Ints []int `json:"ints"`
}

func hello(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Hello from Go")
}

func add(w http.ResponseWriter, r *http.Request) {
	var payload AddPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w,"Invalid JSON",http.StatusBadRequest)	
	}

	sum := 0
	for _,item := range payload.Ints {
		sum += item
	}
	fmt.Fprintln(w, sum)

}

func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("GET /",hello)
	mux.HandleFunc("POST /",add)

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10*time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}


