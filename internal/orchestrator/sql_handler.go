package orchestrator

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

var muSQL sync.Mutex

func checkExistByLogin(login string) bool {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", "/home/ren/GolandProjects/go_calc/data.db")
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

func checkExistByLoginPassword(login, password string) bool {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", "/home/ren/GolandProjects/go_calc/data.db")
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

func registerUser(user User) {
	muSQL.Lock()
	defer muSQL.Unlock()

	db, err := sql.Open("sqlite3", "/home/ren/GolandProjects/go_calc/data.db")
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
