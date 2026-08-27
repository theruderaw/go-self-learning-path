package main

import (
	"log"
	"net/http"

)



func main(){

	cfg := loadConfig()

	db,err := connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Database connected")

	repo := &UserRepository{
		db: db,
	}


	authService := NewAuthService(
		repo,
		cfg.JWTSecret,
	)

	authHandler := NewAuthHandler(authService) 

	authMiddleware := NewAuthMiddleware(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)


	mux.Handle("POST /update",authMiddleware.Authenticate(
		http.HandlerFunc(authHandler.UpdateUser),
	))

	mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./swagger-ui/"))))
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	log.Println("Server running on port :8080")

	err = http.ListenAndServe(":8080",mux)

	if err != nil {
		log.Fatal(err)
	}
}
