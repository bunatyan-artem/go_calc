package orchestrator

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/fifo.v0"
	"log"
	"net/http"
	"strconv"
)

type Expression struct {
	Expression string `json:"expression"`
}

type Id struct {
	Id int `json:"id"`
}

type ExpressionInfo struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result,omitempty"`
}

type Expressions struct {
	Expressions []ExpressionInfo `json:"expressions"`
}

var Trees []*Tree
var Queue *fifo.Queue[*Node]

func expressionInfo(id int) ExpressionInfo {
	status := ""
	switch Trees[id].Flag {
	case 1:
		status = "processing"
	case 2:
		status = "done"
	case 3:
		status = "error: division by zero"
	}
	return ExpressionInfo{id, status, Trees[id].Result}
}

func HandleSendExpr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	expr := Expression{}
	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	tree, nodes := NewTree(expr.Expression)

	if tree.Flag == 15 {
		w.WriteHeader(422)
		return
	}

	tree.Flag = 1
	Trees = append(Trees, tree)
	for _, node := range *nodes {
		if Queue.Len() == Queue.Cap() {
			Queue.Resize(Queue.Cap() + 1)
		}
		Queue.Enqueue(node)
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(Id{
		Id: len(Trees) - 1,
	})
}

func HandleGetExprs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(500)
		return
	}

	var exprs []ExpressionInfo
	for id := range Trees {
		exprs = append(exprs, expressionInfo(id))
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(Expressions{
		Expressions: exprs,
	})
}

func HandleGetExpr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(500)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(vars["id"])
		return
	}

	if id >= len(Trees) {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(expressionInfo(id))
}

func HandleInternal(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		HandleSendTaskAnswer(w, r)
		return
	} else if r.Method == http.MethodGet {
		HandleRequestTask(w, r)
		return
	}
	w.WriteHeader(500)
}

func HandleRequestTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

func HandleSendTaskAnswer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

type Orchestrator struct{}

func NewApplication() *Orchestrator {
	return &Orchestrator{}
}

func (app *Orchestrator) Run(port string) error {
	m := mux.NewRouter()
	Queue = fifo.New[*Node](1)

	m.HandleFunc("/api/v1/calculate", HandleSendExpr)
	m.HandleFunc("/api/v1/expressions", HandleGetExprs)
	m.HandleFunc("/api/v1/expressions/{id:[0-9]+}", HandleGetExpr)
	m.HandleFunc("/internal/task", HandleInternal)

	log.Printf("Runnig on port %s", port)

	err := http.ListenAndServe(":"+port, m)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
