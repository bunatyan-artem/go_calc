# Сервис подсчёта арифметических выражений

## Описание

Веб-сервис. Пользователь отправляет арифметическое выражение по HTTP и может получить ответ. Есть регистрация пользователя и авторизация по логину/паролю (JWT). Выражение может содержать в себе скобки приоритета, не поддерживается унарный минус. Состоит из двух частей: оркестратор и агент.

Оркестратор принимает запросы от пользователей и агентов. При получении выражения оно разбивается на AST-дерево для дальнейшего выполнения агентами элементарных операций. Если при выполнении задачи оркестратор перезагрузится, выражения загрузятся из бд и начнут выполняться заново.

Агент периодически делает запрос оркестратору на получение задачи. Если задача есть - выполняет ее и передает ответ.

---

## Запуск

go version 1.23.1 (64bit)

Установка зависимостей
```
go mod tidy
```

Переменные среды можно поменять в [.env](.env) файле

Запуск

```
go run cmd/main.go
```

Тесты

```
go test ./...
```

Порт 8080 (можно поменять в [main.go](cmd/main.go))

## Эндпоинты

| Эндпоинт                | Метод  | Описание                                                     |
|-------------------------|--------|--------------------------------------------------------------|
| /api/v1/register        | POST   | Регистрация нового пользователя                              |
| /api/v1/login           | POST   | Вход в аккаунт пользователя и получение JWT                  |
| /api/v1/calculate       | POST   | Отправка выражения и получение id                            |
| /api/v1/expressions     | GET    | Получение списка всех выражений и статусов (ответ если есть) |
| /api/v1/expressions/:id | GET    | Получение выражения по id и его статуса (ответ если есть)    |
| /internal/task          | GET    | Взаимодействие оркестратора и агента - отправка задачи       |
| /internal/task          | POST   | Взаимодействие оркестратора и агента - получение результата  |

## Примеры использования


| /api/v1/register                                                                                                                                             | Ответ                                            | Status |
|--------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|--------|
| ```curl --location 'localhost:8080/api/v1/register' --header 'Content-Type: application/json' --data '{"login": "user", "password": "secret"}'```            | ```{"message":"User registered successfully"}``` | 200    |
| ```curl --location 'localhost:8080/api/v1/register' --header 'Content-Type: application/json' --data '{"login": "user", "password": "secret"}'``` (повторно) |  ```{"message":"Login already exists"}```        | 409    |
| ```curl -X GET --location 'localhost:8080/api/v1/register' --header 'Content-Type: application/json' --data '{"login": "user", "password": "secret"}'```     |                                                  | 500    |

| /api/v1/login                                                                                                                                         | Ответ                                   | Status |
|-------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------|--------|
| ```curl --location 'localhost:8080/api/v1/login' --header 'Content-Type: application/json' --data '{"login": "user", "password": "secret"}'```        | ```{"token":"<token>"}```               | 200    |
| ```curl --location 'localhost:8080/api/v1/login' --header 'Content-Type: application/json' --data '{"login": "wrongUser", "password": "secret"}'```   | ```{"error":"Invalid credentials"}```   | 401    |
| ```curl --location 'localhost:8080/api/v1/login' --header 'Content-Type: application/json' --data '{"login": "user", "password": "notSecret"}'```     | ```{"error":"Invalid credentials"}```   | 401    |
| ```curl -X GET --location 'localhost:8080/api/v1/login' --header 'Content-Type: application/json' --data '{"login": "user", "password": "secret"}'``` |                                         | 500    |


| /api/v1/calculate                                                                                                                                                        | Ответ                           | Status |
|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------|--------|
| ```curl --location localhost:8080/api/v1/calculate -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json' -d '{"expression": "2 + 2 * 2 - 2"}'```        | ```{"id":"0"}```                | 201    |
| ```curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression": "2 + 2 * 2 - 2"}'```                               | ```{"error":"Missing token"}``` | 401    |
| ```curl --location localhost:8080/api/v1/calculate -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json' -d '{"expresion": "2 + 2 * 2 - 2"}'```         |                                 | 422    |
| ```curl --location localhost:8080/api/v1/calculate -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json' -d '{"expression": "2 + 2 * 2 - "}'```         |                                 | 422    |
| ```curl -X GET --location localhost:8080/api/v1/calculate -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json' -d '{"expression": "2 + 2 * 2 - 2"}'``` |                                 | 500    |

| /api/v1/expressions                                                                                                                    | Ответ                                                             | Status |
|----------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------|--------|
| ```curl --location localhost:8080/api/v1/expressions -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json'```         | ```{"expressions": [{"id": 0, "status": "done", "result": 4}]}``` | 200    |
| ```curl --location localhost:8080/api/v1/expressions -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json'```         | ```{"error":"Missing token"}```                                   | 401    |
| ```curl -X POST --location localhost:8080/api/v1/expressions -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json'``` |                                                                   | 500    |

| /api/v1/expressions/:id                                                                                                                                                                                   | Ответ                                          | Status |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|--------|
| ```curl --location localhost:8080/api/v1/expressions/0 -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json'```                                                                          | ```{"id": 0, "status": "done", "result": 4}``` | 200    |
| ```curl --location localhost:8080/api/v1/expressions/1 -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json'``` (выражение с таким id не существует или записан за другим пользователем) |                                                | 404    |
| ```curl -X POST --location localhost:8080/api/v1/expressions/0 -H 'Authorization: Bearer <токен>' -H 'Content-Type: application/json'```                                                                  |                                                | 500    |
