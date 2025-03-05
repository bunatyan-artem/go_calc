package main

import (
	"calculator/internal/agent"
	"calculator/internal/orchestrator"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := "8080"
	orch := orchestrator.NewApplication()

	go func() {
		time.Sleep(1 * time.Second)
		cnt, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))

		for i := 0; i < cnt; i++ {
			go func(i int) {
				ag := agent.NewApplication(i, port)
				ag.Run()
			}(i)
		}
	}()

	if err := orch.Run(port); err != nil {
		log.Fatal(err)
	}
}
