package agent

import "log"

type Task struct {
	Id             int     `json:"id"`
	Arg1           float64 `json:"arg_1"`
	Arg2           float64 `json:"arg_2"`
	Operation      uint8   `json:"operation"`
	Operation_time int     `json:"operation_time"`
}

type Response struct {
}

type Application struct {
	id uint8
}

func NewApplication(id uint8) *Application {
	return &Application{id}
}

func (app *Application) Run(port string) error {
	log.Printf("Agent %v working", app.id)

	return nil
}
