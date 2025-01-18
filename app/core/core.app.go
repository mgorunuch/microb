package core

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
)

func Init() {
	flag.Parse()

	LoggerInit()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
