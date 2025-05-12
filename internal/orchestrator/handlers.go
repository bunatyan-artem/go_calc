package orchestrator

import (
	"calculator/internal/orchestrator/Tree"
	"calculator/internal/orchestrator/helpers"
	"calculator/internal/orchestrator/sql"
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

	expr := helpers.Expression{}
	err := json.NewDecoder(r.Body).Decode(&expr)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	tree, nodes := Tree.NewTree(expr.Expression)

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

	sql.RegisterExpression(login, expr.Expression)

	tree.Login = login
	helpers.Trees = append(helpers.Trees, tree)
	for _, node := range *nodes {
		helpers.MuQueue.Lock()
		if helpers.Queue.Len() == helpers.Queue.Cap() {
			_ = helpers.Queue.Resize(helpers.Queue.Cap() + 1)
		}
		helpers.Queue.Enqueue(node)
		helpers.MuQueue.Unlock()
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(helpers.Id{
		Id: len(helpers.Trees) - 1,
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

	var exprs []helpers.ExpressionInfo_
	for id := range helpers.Trees {
		if helpers.Trees[id].Login == login {
			exprs = append(exprs, helpers.ExpressionInfo(id))
		}
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(helpers.Expressions{
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

	if id < 0 || id >= len(helpers.Trees) || helpers.Trees[id].Login != login {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(helpers.ExpressionInfo(id))
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

	var task *Tree.Node
	for {
		helpers.MuQueue.Lock()
		if helpers.Queue.Len() == 0 {
			w.WriteHeader(404)
			helpers.MuQueue.Unlock()
			return
		}

		task, _ = helpers.Queue.Dequeue()
		helpers.MuQueue.Unlock()
		if task.Flag == 5 { //if expression was skipped
			continue
		}

		if helpers.IsNotDivByZero(task) {
			break
		}

		//division by zero
		task.Flag = 5
		for task.Parent != nil {
			task = task.Parent
		}

		var tree *Tree.Tree
		var id uint16
		for i, t := range helpers.Trees {
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
		helpers.Clean(tree.Root)

		sql.SetResult(id, 3, 0)
	}

	w.WriteHeader(200)

	helpers.MuTaskId.Lock()

	helpers.MuSentTasks.Lock()
	helpers.SentTasks[helpers.TaskId] = task
	helpers.MuSentTasks.Unlock()

	helpers.TaskId++
	task.Flag = 4

	json.NewEncoder(w).Encode(helpers.Task{
		helpers.TaskId - 1,
		reflect.Indirect(reflect.ValueOf(task.Left.Val)).Convert(reflect.TypeOf(float64(0))).Float(),
		reflect.Indirect(reflect.ValueOf(task.Right.Val)).Convert(reflect.TypeOf(float64(0))).Float(),
		uint8(reflect.ValueOf(task.Val).Uint()),
		helpers.OperationTime(uint8(reflect.ValueOf(task.Val).Uint())),
	})
	helpers.MuTaskId.Unlock()
}

func HandleSendTaskAnswer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	result := helpers.Result{}
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	helpers.MuSentTasks.Lock()
	task, containsFlag := helpers.SentTasks[result.Id]
	if !containsFlag {
		w.WriteHeader(404)
		return
	}
	delete(helpers.SentTasks, result.Id)
	helpers.MuSentTasks.Unlock()
	w.WriteHeader(200)

	task.Left = nil
	task.Right = nil
	task.Val = result.Result
	task.Flag = 3

	if task.Parent == nil {
		var tree *Tree.Tree
		var id uint16
		for i, t := range helpers.Trees {
			if t.Root == task {
				tree = t
				id = uint16(i)
				break
			}
		}

		tree.Flag = 2
		tree.Result = result.Result

		sql.SetResult(id, 2, result.Result)

		return
	}

	helpers.MuNodeFlagsRemarking.Lock()
	if task.Parent.Left.Flag == 3 && task.Parent.Right.Flag == 3 {
		task.Parent.Flag = 2

		helpers.MuQueue.Lock()
		if helpers.Queue.Len() == helpers.Queue.Cap() {
			_ = helpers.Queue.Resize(helpers.Queue.Cap() + 1)
		}
		helpers.Queue.Enqueue(task.Parent)
		helpers.MuQueue.Unlock()
	}
	helpers.MuNodeFlagsRemarking.Unlock()
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	var user sql.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if sql.CheckExistByLogin(user.Login) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login already exists"})
		return
	}

	sql.RegisterUser(user)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	var user sql.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !sql.CheckExistByLoginPassword(user.Login, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}

	tokenString, err := helpers.GenerateJWT(user.Login)
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
			return helpers.JwtKey, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			login := claims["login"].(string)

			ctx := context.WithValue(r.Context(), "login", login)
			r = r.WithContext(ctx)

			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token claims"})
		}
	}
}
