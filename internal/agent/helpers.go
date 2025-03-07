package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

type Task struct {
	Id            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     uint8   `json:"operation"`
	OperationTime int     `json:"operation_time"`
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

	log.Printf("agent %v get task with id = %v", agent.id, task.Id)
	return &task, nil
}

func (agent *Application) sendResult(task *Task) error {
	data := Response{task.Id, calc(task)}
	log.Printf("agent %v send result with id = %v", agent.id, task.Id)
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
