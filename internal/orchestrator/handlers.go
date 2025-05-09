package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

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

	login, ok := r.Context().Value("login").(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login not found in context"})
		return
	}

	registerExpression(login, expr.Expression)

	tree.Flag = 1
	tree.Login = login
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

	login, ok := r.Context().Value("login").(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login not found in context"})
		return
	}

	var exprs []ExpressionInfo
	for id := range Trees {
		if Trees[id].Login == login {
			exprs = append(exprs, expressionInfo(id))
		}
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

	login, ok := r.Context().Value("login").(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Login not found in context"})
		return
	}

	if id < 0 || id >= len(Trees) || Trees[id].Login != login {
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

		if isNotDivByZero(task) {
			break
		}

		//division by zero
		task.Flag = 5
		for task.Parent != nil {
			task = task.Parent
		}

		var tree *Tree
		var id uint16
		for i, t := range Trees {
			if t.Root == task {
				tree = t
				id = uint16(i)
				break
			}
		}

		if tree == nil {
			panic("problems with finding tree in Trees while solving the division by zero problem")
		}

		tree.Flag = 3
		clean(tree.Root)

		setResult(id, 3, 0)
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
		var id uint16
		for i, t := range Trees {
			if t.Root == task {
				tree = t
				id = uint16(i)
				break
			}
		}

		tree.Flag = 2
		tree.Result = result.Result

		setResult(id, 2, result.Result)

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

func HandleRegister(w http.ResponseWriter, r *http.Request) {
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

func HandleLogin(w http.ResponseWriter, r *http.Request) {
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
