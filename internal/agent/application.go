package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Task struct {
	Id             int     `json:"id"`
	Arg1           float64 `json:"arg_1"`
	Arg2           float64 `json:"arg_2"`
	Operation      uint8   `json:"operation"`
	Operation_time int     `json:"operation_time"`
}

type Response struct {
	Id     int     `json:"id"`
	Result float64 `json:"result"`
}

func calc(task *Task) float64 {
	switch task.Operation {
	case '+':
		return task.Arg1 + task.Arg2
	case '-':
		return task.Arg1 - task.Arg2
	case '*':
		return task.Arg1 * task.Arg2
	case '/':
		return task.Arg1 / task.Arg2
	}
	panic("invalid operation in agent-calc")
}

func (agent *Application) getTask() (*Task, error) {
	resp, err := agent.client.Get("http://localhost:" + agent.port + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d when attempted to get task", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (agent *Application) sendResult(task *Task) error {
	data := Response{task.Id, calc(task)}
	response, _ := json.Marshal(data)

	url := "http://localhost:" + agent.port + "/internal/task"
	resp, err := agent.client.Post(url, "application/json", bytes.NewReader(response))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d when attempted to send result", resp.StatusCode)
	}
	return nil
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

	time.Sleep(time.Duration(task.Operation_time) * time.Millisecond)
	if err := agent.sendResult(task); err != nil {
		panic(err)
	}

	go agent.work()
}

type Application struct {
	id     uint8
	port   string
	client *http.Client
}

func NewApplication(id uint8, port string) *Application {
	return &Application{id, port, &http.Client{}}
}

func (agent *Application) Run(port string) error {
	log.Printf("Agent %v started", agent.id)
	agent.work()

	return nil
}
