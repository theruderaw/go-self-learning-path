package main

import (
	"fmt"
	"net/http"
	"os"

	swagger "github.com/myyrakle/go-swagger-ui/nethttp"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", HandleShorten)
	mux.HandleFunc("GET /{id}", RedirectHandle)
	mux.HandleFunc("GET /stats/{id}", StatsHandler)

	spec, err := os.ReadFile("openapi.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}

	swaggerHandler := swagger.Handler("/swagger", spec)

	mux.Handle("GET /swagger", swaggerHandler)
	mux.Handle("GET /swagger/", swaggerHandler)

	fmt.Println("Server running on :8080")
	fmt.Println("Swagger UI: http://localhost:8080/swagger/")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err)
	}
}
