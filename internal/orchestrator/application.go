package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/fifo.v0"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

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

type Task struct {
	Id            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     uint8   `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Result struct {
	Id     int     `json:"id"`
	Result float64 `json:"result"`
}

var jwtKey = []byte("very_big_secret")

var Trees []*Tree            // all expressions
var Queue *fifo.Queue[*Node] // tasks that might be done
var SentTasks = make(map[int]*Node)
var TaskId = 0

var muTaskId sync.Mutex
var muNodeFlagsRemarking sync.Mutex
var muSentTasks sync.Mutex
var muQueue sync.Mutex

func GenerateJWT(login string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["login"] = login
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

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

func operationTime(op uint8) int {
	switch op {
	case '+':
		res, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
		return res
	case '-':
		res, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
		return res
	case '*':
		res, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
		return res
	case '/':
		res, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
		return res
	}
	panic("invalid operation when attempted to send task")
}

func isDivByZero(task *Node) bool {
	return uint8(reflect.ValueOf(task.Val).Uint()) != '/' ||
		reflect.Indirect(reflect.ValueOf(task.Right.Val)).Convert(reflect.TypeOf(float64(0))).Float() != 0
}

func clean(node *Node) { //marks all nodes that might be skipped beginning from root
	if node == nil {
		return
	}
	node.Flag = 5
	clean(node.Left)
	clean(node.Right)
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
		muQueue.Lock()
		if Queue.Len() == Queue.Cap() {
			_ = Queue.Resize(Queue.Cap() + 1)
		}
		Queue.Enqueue(node)
		muQueue.Unlock()
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

func HandleRequestTask(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var task *Node
	for {
		muQueue.Lock()
		if Queue.Len() == 0 {
			w.WriteHeader(404)
			muQueue.Unlock()
			return
		}

		task, _ = Queue.Dequeue()
		muQueue.Unlock()
		if task.Flag == 5 { //if expression was skipped
			continue
		}

		if isDivByZero(task) {
			break
		}

		//division by zero
		task.Flag = 5
		for task.Parent != nil {
			task = task.Parent
		}

		var tree *Tree
		for _, t := range Trees {
			if t.Root == task {
				tree = t
				break
			}
		}

		if tree == nil {
			panic("problems with finding tree in Trees while solving the division by zero problem")
		}

		tree.Flag = 3
		clean(tree.Root)
	}

	w.WriteHeader(200)

	muTaskId.Lock()

	muSentTasks.Lock()
	SentTasks[TaskId] = task
	muSentTasks.Unlock()

	TaskId++
	task.Flag = 4

	json.NewEncoder(w).Encode(Task{
		TaskId - 1,
		reflect.Indirect(reflect.ValueOf(task.Left.Val)).Convert(reflect.TypeOf(float64(0))).Float(),
		reflect.Indirect(reflect.ValueOf(task.Right.Val)).Convert(reflect.TypeOf(float64(0))).Float(),
		uint8(reflect.ValueOf(task.Val).Uint()),
		operationTime(uint8(reflect.ValueOf(task.Val).Uint())),
	})
	muTaskId.Unlock()
}

func HandleSendTaskAnswer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	result := Result{}
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	muSentTasks.Lock()
	task, containsFlag := SentTasks[result.Id]
	if !containsFlag {
		w.WriteHeader(404)
		return
	}
	delete(SentTasks, result.Id)
	muSentTasks.Unlock()
	w.WriteHeader(200)

	task.Left = nil
	task.Right = nil
	task.Val = result.Result
	task.Flag = 3

	if task.Parent == nil {
		var tree *Tree
		for _, t := range Trees {
			if t.Root == task {
				tree = t
				break
			}
		}

		tree.Flag = 2
		tree.Result = result.Result
		return
	}

	muNodeFlagsRemarking.Lock()
	if task.Parent.Left.Flag == 3 && task.Parent.Right.Flag == 3 {
		task.Parent.Flag = 2

		muQueue.Lock()
		if Queue.Len() == Queue.Cap() {
			_ = Queue.Resize(Queue.Cap() + 1)
		}
		Queue.Enqueue(task.Parent)
		muQueue.Unlock()
	}
	muNodeFlagsRemarking.Unlock()
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if checkExistByLogin(user.Login) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login already exists"})
		return
	}

	registerUser(user)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !checkExistByLoginPassword(user.Login, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}

	tokenString, err := GenerateJWT(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing token"})
			return
		}

		// Удаляем префикс "Bearer "
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			login := claims["login"].(string) // Get login from claims

			// Create a new context with the login value
			ctx := context.WithValue(r.Context(), "login", login)
			// Create a new request with the updated context
			r = r.WithContext(ctx)

			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token claims"})
		}
	}
}

type Orchestrator struct{}

func NewApplication() *Orchestrator {
	return &Orchestrator{}
}

func (app *Orchestrator) Run(port string) error {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	m := mux.NewRouter()
	Queue = fifo.New[*Node](1)

	m.HandleFunc("/api/v1/register", RegisterHandler)
	m.HandleFunc("/api/v1/login", LoginHandler)

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
