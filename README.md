# Заметки (Notes App)

Внутренняя система заметок для сотрудников. Server-rendered HTML — никакого SPA, формы и редиректы на сервере, точечная реактивность через Alpine.js.

## Стек

| Компонент | Технология |
|---|---|
| Язык | Go 1.25+ |
| HTTP-фреймворк | Gin |
| База данных | PostgreSQL 16 |
| Шаблонизатор | Go `html/template` |
| UI-реактивность | Alpine.js (pin toggle, выбор цвета) |
| Миграции | golang-migrate/migrate |
| Контейнеризация | Docker Compose |

## Функции

- **CRUD заметок** — создание, просмотр, редактирование, удаление
- **Пагинация** — 20 заметок на страницу, pinned всегда сверху
- **Поиск** — фильтрация по заголовку (GET-параметр `q`)
- **Pin/unpin** — без перезагрузки страницы через Alpine.js + fetch
- **Палитра цветов** — 6 предустановленных цветов для заметки
- **Flash-сообщения** — уведомления после создания/обновления/удаления
- **CSRF-защита** — double-submit cookie + заголовок `X-CSRF-Token` для AJAX
- **Авторизация** — проверка владельца заметки (каждый пользователь видит только свои)

## Структура проекта

```
cmd/server/main.go        # Точка входа, роутинг
internal/
├── config/               # Конфигурация из env
├── handler/              # CRUD-обработчики
├── middleware/           # Auth + CSRF
├── model/                # Note, NoteForm, Pagination
├── repository/           # SQL-запросы (prepared statements)
├── service/              # Бизнес-логика
└── view/                 # Рендер шаблонов, flash-сообщения
migrations/               # SQL-миграции
templates/                # HTML-шаблоны (layout, notes)
static/                   # CSS, JS
```

## Развёртывание

### Требования
- Go 1.25+
- Docker и Docker Compose (или свой PostgreSQL)
- Make (опционально, для кратких команд)

### 1. Запустить PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Применить миграции

```bash
# через make
make migrate-up

# или напрямую
migrate -path ./migrations -database "postgres://notes:notes@localhost:5432/notes_app?sslmode=disable" up
```

> Если `migrate` не установлен: `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### 3. Запустить приложение

```bash
# через make
make run

# или напрямую
go run ./cmd/server
```

Приложение будет доступно на `http://localhost:8080`.

## Вход

Режим разработки — вход по ID пользователя (1, 2, ...) через форму `/login`.
