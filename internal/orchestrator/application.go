package orchestrator

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/fifo.v0"
	"log"
	"net/http"
)

type Orchestrator struct{}

func NewApplication() *Orchestrator {
	return &Orchestrator{}
}

func (app *Orchestrator) Run(port string) error {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	Queue = fifo.New[*Node](1)
	fillTrees()

	m := mux.NewRouter()

	m.HandleFunc("/api/v1/register", HandleRegister)
	m.HandleFunc("/api/v1/login", HandleLogin)

	m.HandleFunc("/api/v1/calculate", AuthMiddleware(HandleSendExpr))
	m.HandleFunc("/api/v1/expressions", AuthMiddleware(HandleGetExprs))
	m.HandleFunc("/api/v1/expressions/{id:[0-9]+}", AuthMiddleware(HandleGetExpr))

	m.HandleFunc("/internal/task", HandleInternal)

	log.Printf("Runnig on port %s", port)

	err := http.ListenAndServe(":"+port, m)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
