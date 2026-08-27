package main

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret string
}

func loadConfig() Config {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
}
