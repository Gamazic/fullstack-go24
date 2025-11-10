# CRUD пример с использованием sqlx

## Что такое sqlx?

**sqlx** — это расширение стандартной библиотеки `database/sql`, которое добавляет удобные методы для работы с БД:

✅ Автоматический маппинг результатов в структуры (через теги `db`)
✅ `Get()` и `Select()` вместо ручного `Scan()`
✅ Named queries (`:name` вместо `$1, $2, ...`)
✅ Поддержка `IN (...)` через `sqlx.In()`
✅ Все возможности `database/sql` + удобство

**Важно:** sqlx НЕ является ORM! Это просто удобная обертка над `database/sql`.

---

## Архитектура проекта

```
Handler (HTTP) → Service (Бизнес-логика) → Repository (sqlx → PostgreSQL)
```

### Структура файлов:

```
crud/
├── main.go                           # Точка входа (sqlx.Connect)
├── internal/
│   ├── model/
│   │   └── todo.go                   # Entity с тегами `db`
│   ├── repository/
│   │   └── todo_repository.go        # sqlx методы (Get, Select, Named)
│   ├── service/
│   │   └── todo_service.go           # Бизнес-логика
│   └── handler/
│       └── todo_handler.go           # HTTP handlers + DTO
└── go.mod
```

---

## Ключевые отличия от database/sql

### 1. Model с тегами `db`

```go
type Todo struct {
    ID          int64     `db:"id"`           // ← Теги для sqlx
    UserID      int64     `db:"user_id"`
    Title       string    `db:"title"`
    Description string    `db:"description"`
    Completed   bool      `db:"completed"`
    CreatedAt   time.Time `db:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"`
}
```

### 2. Get() вместо QueryRow + Scan

**database/sql:**
```go
var todo Todo
err := db.QueryRowContext(ctx, query, id).Scan(
    &todo.ID,
    &todo.UserID,
    &todo.Title,
    &todo.Description,
    &todo.Completed,
    &todo.CreatedAt,
    &todo.UpdatedAt,
)
```

**sqlx:**
```go
var todo Todo
err := db.GetContext(ctx, &todo, query, id)  // ← Автоматический Scan!
```

### 3. Select() вместо Query + rows.Scan() в цикле

**database/sql:**
```go
rows, err := db.QueryContext(ctx, query, userID)
defer rows.Close()

var todos []*Todo
for rows.Next() {
    todo := &Todo{}
    err := rows.Scan(&todo.ID, &todo.UserID, ...) // ← Много boilerplate
    todos = append(todos, todo)
}
```

**sqlx:**
```go
var todos []*Todo
err := db.SelectContext(ctx, &todos, query, userID)  // ← Всё автоматически!
```

### 4. Named queries

**database/sql:**
```go
query := "UPDATE todos SET title = $1, description = $2, completed = $3 WHERE id = $4"
db.ExecContext(ctx, query, todo.Title, todo.Description, todo.Completed, todo.ID)
```

**sqlx:**
```go
query := `
    UPDATE todos
    SET title = :title, description = :description, completed = :completed
    WHERE id = :id
`
db.NamedExecContext(ctx, query, todo)  // ← Использует теги `db`
```

### 5. IN (...) запросы

**database/sql:**
```go
// Сложно: нужно вручную строить $1, $2, $3...
```

**sqlx:**
```go
ids := []int64{1, 2, 3}
query := "SELECT * FROM todos WHERE id IN (?)"

query, args, _ := sqlx.In(query, ids)  // → SELECT * FROM todos WHERE id IN ($1, $2, $3)
query = db.Rebind(query)               // Для PostgreSQL

var todos []*Todo
db.SelectContext(ctx, &todos, query, args...)
```

---

## Как запустить

### 1. Запустите PostgreSQL

В корневой папке lesson3:

```bash
cd ../..
docker-compose up -d
```

### 2. Установите зависимости

```bash
cd examples/crud
go mod download
```

### 3. Запустите сервер

```bash
go run main.go
```

Ожидаемый вывод:

```
✅ Connected to PostgreSQL with sqlx!
🚀 Server is running on http://localhost:8080

📝 Доступные эндпоинты:
  POST   /todos              - Создать задачу
  GET    /todos              - Список задач
  GET    /todos/get?id=1     - Получить задачу
  POST   /todos/complete?id=1 - Отметить выполненной
  DELETE /todos/delete?id=1  - Удалить задачу

💡 Преимущества sqlx:
  ✅ Автоматический маппинг с помощью тегов `db`
  ✅ db.Get() / db.Select() вместо ручного Scan()
  ✅ Named queries (:name вместо $1, $2...)
  ✅ sqlx.In() для работы с IN (...)
```

---

## Примеры запросов

### 1. Создать задачу

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Изучить sqlx", "description": "Понять преимущества над database/sql"}'
```

### 2. Получить список задач

```bash
curl http://localhost:8080/todos
```

### 3. Отметить задачу как выполненную

```bash
curl -X POST http://localhost:8080/todos/complete?id=1
```

---

## Разбор кода Repository

### GetByID - автоматический маппинг в структуру

```go
func (r *PostgresTodoRepository) GetByID(ctx context.Context, id int64) (*model.Todo, error) {
    query := `
        SELECT id, user_id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE id = $1
    `

    todo := &model.Todo{}

    // ✨ sqlx.Get автоматически делает Scan благодаря тегам `db`
    err := r.db.GetContext(ctx, todo, query, id)
    if err != nil {
        if err.Error() == "sql: no rows in result set" {
            return nil, errors.New("todo not found")
        }
        return nil, err
    }

    return todo, nil
}
```

**Как работает:**
1. sqlx смотрит на теги `db` в структуре `Todo`
2. Автоматически находит соответствующие колонки в результате
3. Заполняет поля структуры

---

### GetAllByUserID - автоматический маппинг slice

```go
func (r *PostgresTodoRepository) GetAllByUserID(ctx context.Context, userID int64) ([]*model.Todo, error) {
    query := `
        SELECT id, user_id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    var todos []*model.Todo

    // ✨ sqlx.Select автоматически создает slice и заполняет его
    // НЕ НУЖНО:
    //   - defer rows.Close()
    //   - for rows.Next() { ... }
    //   - rows.Scan(...)
    err := r.db.SelectContext(ctx, &todos, query, userID)
    if err != nil {
        return nil, err
    }

    return todos, nil
}
```

**Экономия кода:** ~10 строк на каждый запрос!

---

### UpdateNamed - Named queries

```go
func (r *PostgresTodoRepository) UpdateNamed(ctx context.Context, todo *model.Todo) error {
    query := `
        UPDATE todos
        SET title = :title,
            description = :description,
            completed = :completed,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = :id
    `

    // ✨ NamedExecContext использует теги `db` из структуры
    result, err := r.db.NamedExecContext(ctx, query, todo)
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

**Преимущества:**
- Читаемость: `:title` вместо `$1, $2, $3, $4...`
- Меньше ошибок: не нужно считать порядок параметров
- Легко добавлять/удалять поля

---

### GetByIDs - работа с IN (...)

```go
func (r *PostgresTodoRepository) GetByIDs(ctx context.Context, ids []int64) ([]*model.Todo, error) {
    query := `
        SELECT id, user_id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE id IN (?)
        ORDER BY created_at DESC
    `

    // ✨ sqlx.In преобразует ? в $1, $2, $3 для PostgreSQL
    query, args, err := sqlx.In(query, ids)
    if err != nil {
        return nil, err
    }

    // Rebind для правильных placeholder'ов PostgreSQL
    query = r.db.Rebind(query)

    var todos []*model.Todo
    err = r.db.SelectContext(ctx, &todos, query, args...)
    if err != nil {
        return nil, err
    }

    return todos, nil
}
```

**Пример использования:**
```go
todos, _ := repo.GetByIDs(ctx, []int64{1, 2, 3, 4, 5})
// → SELECT ... WHERE id IN ($1, $2, $3, $4, $5)
```

---

### BatchInsert - массовая вставка

```go
func (r *PostgresTodoRepository) BatchInsert(ctx context.Context, todos []*model.Todo) error {
    query := `
        INSERT INTO todos (user_id, title, description, completed)
        VALUES (:user_id, :title, :description, :completed)
    `

    // ✨ NamedExec может принимать slice структур
    _, err := r.db.NamedExecContext(ctx, query, todos)
    return err
}
```

---

## Сравнение: database/sql vs sqlx

| Задача | database/sql | sqlx | Экономия строк |
|--------|--------------|------|----------------|
| Одна строка | QueryRow + Scan (7 полей) | Get() | ~6 строк |
| Много строк | Query + defer + loop + Scan | Select() | ~10 строк |
| UPDATE/INSERT | $1, $2, $3... | :name, :email... | Читаемость ⬆ |
| IN (...) | Вручную строить | sqlx.In() | ~15 строк |

---

## Когда использовать sqlx?

### ✅ Используйте sqlx:

- Много CRUD операций
- Хотите меньше boilerplate
- Нужны Named queries
- Работаете с IN (...)
- Хотите читаемый код

### ❌ НЕ используйте sqlx:

- Очень сложные SQL-запросы с динамическими условиями
- Нужен полный контроль над каждым байтом
- Микрооптимизации производительности критичны

---

## Производительность

**sqlx почти идентичен database/sql по скорости**, потому что:
- Под капотом те же `database/sql` методы
- Единственный overhead — рефлексия для маппинга структур
- Overhead минимален (несколько микросекунд на запрос)

**Вывод:** sqlx не замедляет ваше приложение!

---

## Задания для практики

1. Добавьте метод `GetCompletedTodos(ctx, userID)` в Repository
2. Реализуйте `BatchUpdate` для обновления нескольких задач
3. Добавьте метод `SearchByTitle(ctx, pattern)` с LIKE
4. Создайте метод `GetStatistics(ctx, userID)` с агрегацией (COUNT, SUM)
5. Добавьте фильтрацию с динамическими условиями

---

## Полезные ссылки

- [sqlx GitHub](https://github.com/jmoiron/sqlx)
- [sqlx Illustrated Guide](http://jmoiron.github.io/sqlx/)
- [database/sql Documentation](https://pkg.go.dev/database/sql)

---

## Ключевые моменты

✅ **sqlx = database/sql + удобство**
✅ **Теги `db`** для автоматического маппинга
✅ **Get() / Select()** вместо ручного Scan()
✅ **Named queries** (`:name`) для читаемости
✅ **sqlx.In()** для работы с IN (...)
✅ **Производительность ≈ database/sql**
✅ **НЕ ORM!** Вы все еще пишете SQL

---

## Что дальше?

- Добавить транзакции (`db.Beginx()`)
- Добавить middleware для логирования
- Использовать `sqlx.NamedQuery()` для сложных запросов
- Добавить тесты с testify/sqlmock
