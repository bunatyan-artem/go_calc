package application

import (
	"calculator/pkg/calc"
	"encoding/json"
	"log"
	"net/http"
)

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		_ = json.NewEncoder(w).Encode(Response{
			Error: "Internal server error",
		})
		return
	}

	request := Request{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(500)
		_ = json.NewEncoder(w).Encode(Response{
			Error: "Internal server error",
		})
		return
	}

	result, err := calc.Calc(request.Expression)
	if err != nil {
		w.WriteHeader(422)
		_ = json.NewEncoder(w).Encode(Response{
			Error: "Expression is not valid",
		})
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(Response{
		Result: result,
	})
}

type Application struct{}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Run(port string) error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	log.Printf("Runnig on port %s", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
