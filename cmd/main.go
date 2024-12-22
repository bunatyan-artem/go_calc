package main

import (
	"calculator/internal/application"
	"log"
)

func main() {
	app := application.NewApplication()
	err := app.Run("8080")
	if err != nil {
		log.Fatal(err)
	}
}
