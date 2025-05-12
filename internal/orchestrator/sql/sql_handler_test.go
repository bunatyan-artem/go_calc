package sql

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"testing"
)

var db *sql.DB

func setupTestDB() {
	DbPath = "test.db"

	var err error
	db, err = sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		login TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);
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
		log.Fatalf("Ошибка создания таблиц: %v", err)
	}
}

func teardownTestDB() {
	_, err := db.Exec("DROP TABLE IF EXISTS users; DROP TABLE IF EXISTS expressions;")
	if err != nil {
		log.Fatalf("Ошибка удаления таблиц: %v", err)
	}
}

func TestCheckExistByLogin(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	_, err := db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", "testuser", "password123")
	if err != nil {
		t.Fatalf("Ошибка вставки тестового пользователя: %v", err)
	}

	exists := CheckExistByLogin("testuser")
	if !exists {
		t.Fatalf("Ожидалось, что пользователь существует")
	}

	exists = CheckExistByLogin("nonexistent")
	if exists {
		t.Fatalf("Ожидалось, что пользователь не существует")
	}
}

func TestCheckExistByLoginPassword(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	_, err := db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", "testuser", "password123")
	if err != nil {
		t.Fatalf("Ошибка вставки тестового пользователя: %v", err)
	}

	exists := CheckExistByLoginPassword("testuser", "password123")
	if !exists {
		t.Fatalf("Ожидалось, что логин и пароль правильные")
	}

	exists = CheckExistByLoginPassword("testuser", "wrongpassword")
	if exists {
		t.Fatalf("Ожидалось, что неправильный пароль не пройдет")
	}
}

func TestRegisterUser(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	user := User{Login: "newuser", Password: "newpassword123"}
	RegisterUser(user)

	exists := CheckExistByLogin("newuser")
	if !exists {
		t.Fatalf("Ожидалось, что новый пользователь был зарегистрирован")
	}
}

func TestRegisterExpression(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	user := User{Login: "testuser", Password: "password123"}
	RegisterUser(user)

	RegisterExpression("testuser", "2+2")

	rows := GetExprs()
	defer rows.Close()

	var id int
	var login string
	var expression string
	var status uint8
	var result float64

	found := false
	for rows.Next() {
		err := rows.Scan(&id, &login, &expression, &status, &result)
		if err != nil {
			t.Fatalf("Ошибка сканирования строки: %v", err)
		}

		if login == "testuser" && expression == "2+2" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Ожидалось, что выражение будет зарегистрировано")
	}
}

func TestSetResult(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	user := User{Login: "testuser", Password: "password123"}
	RegisterUser(user)

	RegisterExpression("testuser", "2+2")

	rows := GetExprs()

	var id int
	var login string
	var expression string
	var status uint8
	var result float64

	for rows.Next() {
		err := rows.Scan(&id, &login, &expression, &status, &result)
		if err != nil {
			t.Fatalf("Ошибка сканирования строки: %v", err)
		}
	}
	rows.Close()

	err := db.Close()
	if err != nil {
		log.Fatalf("не удалось закрыть бд: %v", err)
	}
	SetResult(uint16(id-1), 2, 4.0)
	db, err = sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}

	rows = GetExprs()
	defer rows.Close()

	found := false
	for rows.Next() {
		err = rows.Scan(&id, &login, &expression, &status, &result)
		if err != nil {
			t.Fatalf("Ошибка сканирования строки: %v", err)
		}

		if id == 1 && status == 2 && result == 4.0 {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Ожидалось, что результат и статус будут обновлены, получено id=%v, status=%v, result=%v", id, status, result)
	}
}

func TestGetExprs(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	user := User{Login: "testuser", Password: "password123"}
	RegisterUser(user)

	RegisterExpression("testuser", "2+2")
	RegisterExpression("testuser", "3*3")

	rows := GetExprs()
	defer rows.Close()

	var id int
	var login string
	var expression string
	var status uint8
	var result float64

	foundExpressions := 0
	for rows.Next() {
		err := rows.Scan(&id, &login, &expression, &status, &result)
		if err != nil {
			t.Fatalf("Ошибка сканирования строки: %v", err)
		}

		if login == "testuser" && (expression == "2+2" || expression == "3*3") {
			foundExpressions++
		}
	}

	if foundExpressions != 2 {
		t.Fatalf("Ожидалось, что будут найдены оба выражения")
	}
}
