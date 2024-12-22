# Сервис подсчёта арифметических выражений

## Описание

Веб-сервис. Пользователь отправляет арифметическое выражение по HTTP и получает в ответ его результат. Выражение может содержать в себе скобки приоритета, не поддерживается унарный минус и нецелые числа во входных данных.

---

## Запуск

go version 1.23.0

```
go run cmd/main.go
```

Порт 8080 (можно поменять в [main.go](cmd/main.go))

## Примеры использования

Запрос без ошибок
```
curl -X POST http://localhost:8080/api/v1/calculate \
-H "Content-Type: application/json" \
-d '{"expression":"2 * (2 + 2)"}'
```
code 200
```
{
    "result": 8
}
```

Запрос с плохим expression
```
curl -X POST http://localhost:8080/api/v1/calculate \
-H "Content-Type: application/json" \
-d '{"expression":"2 * (2 + 2"}'
```
code 422
```
{
    "error": "Expression is not valid"
}
```
Неправильный запрос
```
curl -X Get http://localhost:8080/api/v1/calculate \
-H "Content-Type: application/json" \
-d '{"expression":"2 * (2 + 2)"}'
```
code 500
```
{
    "error": "Internal server error"
}
```
