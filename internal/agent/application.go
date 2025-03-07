package agent

import (
	"log"
	"net/http"
	"time"
)

type Application struct {
	id     int
	port   string
	client *http.Client
}

func NewApplication(id int, port string) *Application {
	return &Application{id, port, &http.Client{}}
}

func (agent *Application) Run() {
	log.Printf("Agent %v started", agent.id)
	agent.work()
}

func (agent *Application) work() {
	task, err := agent.getTask()
	if err != nil {
		log.Println(err)
	}

	if task == nil {
		time.Sleep(1 * time.Second)
		go agent.work()
		return
	}

	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	if err := agent.sendResult(task); err != nil {
		panic(err)
	}

	go agent.work()
}
