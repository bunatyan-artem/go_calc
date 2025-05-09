package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Импорт драйвера SQLite3
	"log"
)

func getExprs() *sql.Rows {
	db, err := sql.Open("sqlite3", "/home/ren/GolandProjects/go_calc/data.db")
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

func print() {
	rows := getExprs()
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
		log.Println(id, login, expression, status, result)
	}
}

func clear() {
	db, err := sql.Open("sqlite3", "/home/ren/GolandProjects/go_calc/data.db")
	if err != nil {
		log.Printf("Ошибка открытия соединения с базой данных: %v", err)
	}
	defer db.Close()

	tableName := "expressions" // Replace with your table name

	// 1. Delete all rows
	_, err = db.Exec("DELETE FROM " + tableName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("All rows deleted from", tableName)

	// 2. Reset the autoincrement counter (if applicable)
	_, err = db.Exec("DELETE FROM sqlite_sequence WHERE name='" + tableName + "'")
	if err != nil {
		log.Println("Error resetting autoincrement:", err) // Log the error, but don't fatal
		// It's okay if this fails if the table doesn't have autoincrement
	} else {
		fmt.Println("Autoincrement counter reset for", tableName)
	}

	fmt.Println("Table truncated (simulated) successfully.")
}

func main() {
	print()
	clear()
	print()
}
