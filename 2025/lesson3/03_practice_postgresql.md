# Часть 3: Практика — PostgreSQL и Docker

## Запускаем PostgreSQL через Docker

### 1. Создаем docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: lesson3_db
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: myapp_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d

volumes:
  postgres_data:
```

### 2. Создаем миграции

**Файл:** `migrations/001_init.sql`

```sql
-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу задач
CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);

-- Тестовые данные
INSERT INTO users (email, password_hash) VALUES
    ('alice@example.com', '$2a$10$hash1'),
    ('bob@example.com', '$2a$10$hash2');

INSERT INTO todos (user_id, title, description, completed) VALUES
    (1, 'Купить молоко', 'Обезжиренное, 1 литр', false),
    (1, 'Сделать ДЗ', 'Лекция 3 по Go', false),
    (2, 'Прочитать книгу', 'Clean Code', true);
```

### 3. Запускаем

```bash
# Запуск
docker-compose up -d

# Проверка
docker-compose ps

# Логи
docker-compose logs postgres

# Подключение к БД
docker exec -it lesson3_db psql -U postgres -d myapp_db
```

### Внутри psql:

```sql
\dt                    -- Список таблиц
\d todos               -- Описание таблицы
SELECT * FROM users;   -- Проверка данных
\q                     -- Выход
```

### Остановка

```bash
docker-compose stop         # Остановить (данные сохраняются)
docker-compose start        # Запустить снова
docker-compose down -v      # Удалить контейнер И данные
```

---

## CRUD операции в Go

### CREATE — Добавление задачи

```go
func (r *TodoRepository) Create(ctx context.Context, todo *model.Todo) (int64, error) {
    query := `
        INSERT INTO todos (user_id, title, description, completed)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `

    err := r.db.QueryRowContext(
        ctx,
        query,
        todo.UserID,
        todo.Title,
        todo.Description,
        todo.Completed,
    ).Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt)

    return todo.ID, err
}
```

**Ключевые моменты:**
- `$1, $2, $3` — параметры (защита от SQL-инъекций)
- `RETURNING` — PostgreSQL возвращает сгенерированные значения
- `QueryRowContext` — для запросов с одной строкой результата

---

### READ — Получение задачи по ID

```go
func (r *TodoRepository) GetByID(ctx context.Context, id int64) (*model.Todo, error) {
    query := `
        SELECT id, user_id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE id = $1
    `

    todo := &model.Todo{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &todo.ID,
        &todo.UserID,
        &todo.Title,
        &todo.Description,
        &todo.Completed,
        &todo.CreatedAt,
        &todo.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, errors.New("todo not found")
    }

    return todo, err
}
```

**Обработка ошибок:**
- `sql.ErrNoRows` — строка не найдена (НЕ критическая ошибка)

---

### READ ALL — Получение всех задач пользователя

```go
func (r *TodoRepository) GetAllByUserID(ctx context.Context, userID int64) ([]*model.Todo, error) {
    query := `
        SELECT id, user_id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()  // ⚠️ ОБЯЗАТЕЛЬНО!

    var todos []*model.Todo
    for rows.Next() {
        todo := &model.Todo{}
        err := rows.Scan(
            &todo.ID,
            &todo.UserID,
            &todo.Title,
            &todo.Description,
            &todo.Completed,
            &todo.CreatedAt,
            &todo.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        todos = append(todos, todo)
    }

    return todos, rows.Err()
}
```

**Важно:**
- `defer rows.Close()` — закрываем rows, иначе утечка соединений!
- `rows.Err()` — проверяем ошибки после цикла

---

### UPDATE — Обновление задачи

```go
func (r *TodoRepository) Update(ctx context.Context, todo *model.Todo) error {
    query := `
        UPDATE todos
        SET title = $1, description = $2, completed = $3, updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
    `

    result, err := r.db.ExecContext(ctx, query,
        todo.Title,
        todo.Description,
        todo.Completed,
        todo.ID,
    )
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("todo not found")
    }

    return nil
}
```

**Ключевые моменты:**
- `ExecContext` — для запросов без возврата данных
- `RowsAffected()` — сколько строк изменено (0 = не найдена)

---

### DELETE — Удаление задачи

```go
func (r *TodoRepository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM todos WHERE id = $1`

    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("todo not found")
    }

    return nil
}
```

---

## Связи между таблицами

### One-to-Many (Один ко многим)

**Пример:** Один пользователь → много задач

```
users (1)  ←───→  (N) todos
  ↓                    ↓
 id                 user_id (Foreign Key)
```

**SQL:**
```sql
CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    ...
);
```

**Запрос с JOIN:**
```sql
SELECT
    u.id, u.email,
    t.id, t.title, t.completed
FROM users u
LEFT JOIN todos t ON u.id = t.user_id
WHERE u.id = 1;
```

---

### Many-to-Many (Многие ко многим)

**Пример:** Студенты ↔ Курсы

```
students (N) ←───→ student_courses ←───→ (N) courses
```

**SQL:**
```sql
CREATE TABLE student_courses (
    student_id INTEGER REFERENCES students(id) ON DELETE CASCADE,
    course_id INTEGER REFERENCES courses(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (student_id, course_id)
);
```

**Запрос: все курсы студента**
```sql
SELECT c.id, c.title
FROM courses c
JOIN student_courses sc ON c.id = sc.course_id
WHERE sc.student_id = 1;
```

---

## Best Practices

✅ **Всегда используйте `context.Context`** для тайм-аутов
✅ **Закрывайте `rows.Close()`** — иначе утечка соединений
✅ **Параметризованные запросы** ($1, $2) — защита от SQL-инъекций
✅ **Обрабатывайте `sql.ErrNoRows`** отдельно
✅ **Один `*sql.DB`** на всё приложение
✅ **Используйте транзакции** для связанных операций

❌ **НЕ конкатенируйте строки для SQL** (SQL-инъекции!)
❌ **НЕ игнорируйте `rows.Err()`**
❌ **НЕ забывайте `defer rows.Close()`**

---

## Типичные ошибки

### 1. SQL-инъекция
```go
// ❌ ОПАСНО
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)

// ✅ БЕЗОПАСНО
query := "SELECT * FROM users WHERE email = $1"
db.QueryRowContext(ctx, query, email)
```

### 2. Забыли закрыть rows
```go
// ❌ Утечка соединений
rows, _ := db.QueryContext(ctx, query)
for rows.Next() { ... }

// ✅ Правильно
rows, _ := db.QueryContext(ctx, query)
defer rows.Close()
for rows.Next() { ... }
```

### 3. Не проверили RowsAffected
```go
// ❌ Не знаем, обновилась ли строка
db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", name, id)

// ✅ Проверяем
result, _ := db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", name, id)
if rows, _ := result.RowsAffected(); rows == 0 {
    return errors.New("user not found")
}
```

---

## Полный пример: Handler → Repository

### Handler:
```go
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
    var req CreateTodoRequest
    json.NewDecoder(r.Body).Decode(&req)

    if req.Title == "" {
        http.Error(w, "Title is required", 400)
        return
    }

    userID := getUserIDFromJWT(r)
    todo, err := h.service.CreateTodo(r.Context(), userID, req.Title)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    json.NewEncoder(w).Encode(todo)
}
```

### Service:
```go
func (s *TodoService) CreateTodo(ctx context.Context, userID int64, title string) (*model.Todo, error) {
    if len(title) > 255 {
        return nil, errors.New("title too long")
    }

    todo := &model.Todo{
        UserID: userID,
        Title:  title,
    }

    id, err := s.repo.Create(ctx, todo)
    if err != nil {
        return nil, err
    }

    todo.ID = id
    return todo, nil
}
```

### Repository:
```go
func (r *TodoRepository) Create(ctx context.Context, todo *model.Todo) (int64, error) {
    query := `INSERT INTO todos (user_id, title) VALUES ($1, $2) RETURNING id`
    var id int64
    err := r.db.QueryRowContext(ctx, query, todo.UserID, todo.Title).Scan(&id)
    return id, err
}
```

---

## Домашнее задание

**ДЗ #5:** Спроектировать модель данных для вашего проекта

**Требования:**
- Минимум 2 таблицы + `users`
- SQL-миграции (CREATE TABLE)
- docker-compose.yml
- Описание связей между таблицами

**Примеры:**

1. **Интернет-магазин:**
   - `users`, `products`, `orders`, `order_items`

2. **Блог:**
   - `users`, `posts`, `comments`, `tags`, `post_tags`

3. **Система бронирования:**
   - `users`, `rooms`, `bookings`

---

## Полезные ссылки

- [PostgreSQL Tutorial](https://www.postgresql.org/docs/current/tutorial.html)
- [database/sql Best Practices](https://go.dev/doc/database/querying)
- [pgx GitHub](https://github.com/jackc/pgx)
- [Three Dots Labs - Repository Pattern](https://threedots.tech/post/repository-pattern-in-go/)
