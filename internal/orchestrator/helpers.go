package orchestrator

import (
	"calculator/internal/orchestrator/sql"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/fifo.v0"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
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

func isNotDivByZero(task *Node) bool {
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

func fillTrees() {
	rows := sql.GetExprs()
	if rows == nil {
		return
	}

	for rows.Next() {
		var id int
		var login string
		var expression string
		var status uint8
		var result float64

		err := rows.Scan(&id, &login, &expression, &status, &result)
		if err != nil {
			log.Fatalf("Ошибка сканирования строки: %v", err)
		}
		if id-1 != len(Trees) {
			panic("Expr from db came with wrong id")
		}

		tree := &Tree{Flag: status, Result: result, Login: login}
		if status == 1 {
			tempTree, nodes := NewTree(expression)
			tree.Root = tempTree.Root

			for _, node := range *nodes {
				muQueue.Lock()
				if Queue.Len() == Queue.Cap() {
					_ = Queue.Resize(Queue.Cap() + 1)
				}
				Queue.Enqueue(node)
				muQueue.Unlock()
			}
		}
		Trees = append(Trees, tree)
	}
}
