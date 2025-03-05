package main

import (
	"calculator/internal/orchestrator"
	"log"
)

func main() {
	port := "8080"
	orch := orchestrator.NewApplication()
	err := orch.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}
