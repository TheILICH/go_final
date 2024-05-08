package main

import (
	"github.com/joho/godotenv"
	"go_final/route"
	"log"
)

func loadenv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error while loading .env file: " + err.Error())
	}
}

func main() {
	// loadenv()
	log.Fatal(route.RunAPI(":8080"))
}
