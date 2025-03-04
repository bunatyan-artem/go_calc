package main

import (
	"calculator/internal/orchestrator"
	"log"
)

func main() {
	orch := orchestrator.NewApplication()
	err := orch.Run("8080")
	if err != nil {
		log.Fatal(err)
	}
}
