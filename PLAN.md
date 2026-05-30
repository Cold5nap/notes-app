# План приложения "Заметки" (Notes App)

## 1. Стек технологий

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.22+ |
| База данных | PostgreSQL 16 |
| HTTP-фреймворк | Gin |
| Шаблонизатор | Go `html/template` |
| Миграции | golang-migrate/migrate |
| UI | Alpine.js + CSS (без бандлера) |
| Контейнеризация | Docker Compose |

**Почему Go + HTML-шаблоны, а не JSON API + SPA?**
Задание требует server-rendered HTML: формы, redirect после сохранения, flash-сообщения, CSRF-токены. SPA здесь избыточен. Alpine.js на клиенте для точечной реактивности (pin toggle, выбор цвета).

---

## 2. Структура проекта

```
notes-app/
├── cmd/
│   └── server/
│       └── main.go            # Точка входа: запуск сервера
├── internal/
│   ├── config/
│   │   └── config.go          # Чтение конфига из env
│   ├── handler/
│   │   ├── notes.go           # CRUD-обработчики
│   │   └── notes_test.go      # Тесты хендлеров
│   ├── middleware/
│   │   ├── auth.go            # Проверка авторизации (stub для тестового)
│   │   └── csrf.go            # CSRF-защита
│   ├── model/
│   │   └── note.go            # Структура Note, валидация
│   ├── repository/
│   │   └── note_repo.go       # SQL-запросы к БД (prepared statements)
│   ├── service/
│   │   └── note_service.go    # Бизнес-логика
│   └── view/
│       ├── layout.go          # Обёртка для рендера шаблонов
│       └── flash.go           # Flash-сообщения (session-based)
├── migrations/
│   ├── 000001_create_notes_table.up.sql
│   └── 000001_create_notes_table.down.sql
├── templates/
│   ├── layout/
│   │   └── base.html          # Основной layout (head, header, flash)
│   ├── notes/
│   │   ├── index.html         # Список заметок (grid карточек)
│   │   ├── form.html          # Форма создания/редактирования
│   │   └── pagination.html    # Компонент пагинации
│   └── components/
│       └── color_picker.html  # Палитра цветов
├── static/
│   ├── css/
│   │   └── app.css            # Стили (grid, цветная полоска, булавка)
│   └── js/
│       └── app.js             # Alpine.js-компоненты (pin toggle, color picker)
├── docker-compose.yml         # Postgres + приложение
├── Dockerfile                 # Multi-stage сборка Go
├── .env.example               # Пример переменных окружения
├── go.mod
├── Makefile                   # Сборка, миграции, тесты
└── README.md
```

---

## 3. Модель данных (PostgreSQL)

```sql
CREATE TABLE notes (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL,
    title      VARCHAR(255) NOT NULL,
    content    TEXT,
    color      VARCHAR(7) DEFAULT '#6366f1',
    is_pinned  BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_user_pinned ON notes(user_id, is_pinned DESC);
```

**Изменения относительно задания:**
- `SERIAL` вместо `INT AUTO_INCREMENT` (постгрес-синтаксис)
- `BOOLEAN` вместо `TINYINT(1)` (идиоматичнее для PG)
- `TIMESTAMPTZ` вместо `DATETIME` (храним с часовым поясом)
- `updated_at` обновляется через триггер или на уровне приложения

---

## 4. Маршруты (Gin)

| Метод | Путь | Handler | Описание |
|-------|------|---------|----------|
| GET | /notes | `NotesIndex` | Список + пагинация + поиск |
| GET | /notes/create | `NotesCreate` | Форма создания |
| POST | /notes | `NotesStore` | Сохранение новой |
| GET | /notes/:id/edit | `NotesEdit` | Форма редактирования |
| POST | /notes/:id | `NotesUpdate` | Обновление |
| POST | /notes/:id/delete | `NotesDestroy` | Удаление |
| POST | /notes/:id/toggle-pin | `NotesTogglePin` | JSON-ответ |

Группа `/notes` обёрнута в middleware `AuthRequired()`.

---

## 5. Ключевые решения

### 5.1 Пагинация с закреплёнными сверху
```sql
SELECT * FROM notes
WHERE user_id = $1
  AND ($2 = '' OR title ILIKE '%' || $2 || '%')
ORDER BY is_pinned DESC, updated_at DESC
LIMIT $3 OFFSET $4;
```

### 5.2 Безопасность
- **Prepared statements** — все запросы через `db.PrepareContext()` или параметризованные запросы `$1, $2, ...`
- **User ownership** — каждое обращение проверяет `user_id` из сессии
- **CSRF** — double-submit cookie или скрытое поле с токеном
- **Экранирование** — Go `html/template` автоматически экранирует вывод

### 5.3 Pin toggle (AJAX)
Alpine.js отправляет `POST /notes/{id}/toggle-pin` → сервер возвращает `{"ok": true, "is_pinned": true}` → Alpine обновляет иконку булавки без перезагрузки.

### 5.4 Палитра цветов
6 предустановленных цветов на выбор. Хранятся hex-строкой в БД. На фронте — Alpine.js переключает active-класс кружку, скрытое поле хранит выбранный hex.

---

## 6. Docker Compose

```yaml
services:
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: notes
      POSTGRES_PASSWORD: notes
      POSTGRES_DB: notes_app
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U notes -d notes_app"]
      interval: 5s
      timeout: 3s
      retries: 5

  app:
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: notes
      DB_PASSWORD: notes
      DB_NAME: notes_app
      DB_SSLMODE: disable
      SESSION_SECRET: change-me-in-production
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  pg_data:
```

---

## 7. Очерёдность разработки

1. **Настройка проекта** — go mod, docker-compose, Makefile, структура папок
2. **Миграция БД** — SQL для notes + golang-migrate
3. **Config + DB connection** — чтение env, подключение к PG
4. **Модель + Repository** — Note struct, CRUD-запросы
5. **Service** — бизнес-логика (проверка владельца, пагинация)
6. **Middleware** — auth (stub), CSRF
7. **Handler + Templates** — все 7 endpoint + html-шаблоны
8. **Static (CSS + Alpine.js)** — сетка карточек, цветная полоска, pin, поиск
9. **Dockerfile** — multi-stage сборка Go-бинарника
10. **Тесты** — handler tests с httptest
