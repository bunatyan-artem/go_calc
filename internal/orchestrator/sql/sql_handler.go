package sql

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

var dbPath string = "data.db"

var muSQL sync.Mutex

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func CheckExistByLogin(login string) bool {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
		return false
	}
	defer db.Close()

	query := "SELECT COUNT(*) FROM users WHERE login = ?"
	var count int

	err = db.QueryRow(query, login).Scan(&count)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return false
	}

	return count > 0
}

func CheckExistByLoginPassword(login, password string) bool {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
		return false
	}
	defer db.Close()

	query := "SELECT COUNT(*) FROM users WHERE login = ? AND password = ?"
	var count int

	err = db.QueryRow(query, login, password).Scan(&count)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return false
	}

	return count > 0
}

func RegisterUser(user User) {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
		return
	}
	defer db.Close()

	query := "INSERT INTO users (login, password) VALUES (?, ?)"
	_, err = db.Exec(query, user.Login, user.Password)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return
	}
}

func RegisterExpression(login, expression string) {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
	}
	defer db.Close()

	query := "INSERT INTO expressions (login, expression) VALUES (?, ?)"

	_, err = db.Exec(query, login, expression)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return
	}
}

func SetResult(id uint16, status uint8, result float64) {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
	}
	defer db.Close()

	query := "UPDATE expressions SET status = ?, result = ? WHERE id = ?"

	_, err = db.Exec(query, status, result, id+1)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return
	}
}

func GetExprs() *sql.Rows {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
	}
	defer db.Close()

	query := "SELECT * FROM expressions ORDER BY id ASC"

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Ошибка выполнения запроса: %v", err)
		return nil
	}
	return rows
}
