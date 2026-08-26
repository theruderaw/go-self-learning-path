package main

import (
	"fmt"
	"net/http"
	"os"

	swagger "github.com/myyrakle/go-swagger-ui/nethttp"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", hello)

	mux.HandleFunc("GET /tasks", getTasksHandler)
	mux.HandleFunc("POST /tasks", createTaskHandler)

	mux.HandleFunc("GET /tasks/{id}", getTaskHandler)
	mux.HandleFunc("PATCH /tasks/{id}", updateTaskHandler)
	mux.HandleFunc("DELETE /tasks/{id}", deleteTaskHandler)

	mux.HandleFunc("PATCH /tasks/{id}/complete", completeTaskHandler)

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
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
	}
}
