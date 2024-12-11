package core

import (
	"github.com/joho/godotenv"
	"log"
)

func Init() {
	LoggerInit()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
