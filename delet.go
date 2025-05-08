package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Импорт драйвера SQLite3
	"log"
)

func main() {
	db, err := sql.Open("sqlite3", "data.db") // Открываем или создаем базу данных file.db
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем таблицу users
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			login TEXT PRIMARY KEY,
			password TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Таблица 'users' успешно создана.")

	// Создаем таблицу expressions
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT,
			expression TEXT NOT NULL,
			status INTEGER DEFAULT 1,
			result REAL DEFAULT 0,
			FOREIGN KEY (login) REFERENCES users(login)
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Таблица 'expressions' успешно создана.")

	fmt.Println("База данных и таблицы созданы успешно!")
}
