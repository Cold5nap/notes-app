# Заметки (Notes App)

Внутренняя система заметок для сотрудников.  
Server-rendered HTML на Go + Gin, PostgreSQL, Alpine.js.

## Развёртывание

### Требования
- Go 1.25+
- Docker и Docker Compose (или свой PostgreSQL)

### 1. Запустить PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Применить миграции

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path ./migrations -database "postgres://notes:notes@localhost:5432/notes_app?sslmode=disable" up
```

### 3. Запустить приложение

```bash
go run ./cmd/server
```

Приложение будет доступно на `http://localhost:8080`.

## Вход

Режим разработки — вход по ID пользователя (1, 2, ...) через форму `/login`.
