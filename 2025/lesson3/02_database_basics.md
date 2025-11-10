# Часть 2: Основы работы с БД в Go

## Стандартная библиотека database/sql

### Что такое database/sql?

`database/sql` — это **интерфейс** для работы с SQL-базами данных в Go.

✅ Единый API для разных БД (PostgreSQL, MySQL, SQLite)
✅ Connection Pool из коробки
✅ Поддержка транзакций
✅ Защита от SQL-инъекций
✅ Поддержка контекста (тайм-ауты, отмена)

⚠️ **database/sql — это НЕ драйвер!** Это интерфейс.

---

## Драйверы для PostgreSQL

```
database/sql (интерфейс)
    ↓
Драйвер (реализация)
    ↓
PostgreSQL
```

### Два популярных драйвера:

#### 1. lib/pq (старый, стабильный)
```go
import _ "github.com/lib/pq"
db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname")
```

#### 2. pgx (современный, рекомендуется)
```go
import _ "github.com/jackc/pgx/v5/stdlib"
db, err := sql.Open("pgx", "postgres://user:pass@localhost/dbname")
```

**Для вашего проекта используйте pgx!**

---

## Базовое подключение

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    // DSN (Data Source Name) - строка подключения
    dsn := "postgres://postgres:postgres@localhost:5432/mydb?sslmode=disable"

    // Открываем соединение
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // ⚠️ ВАЖНО: sql.Open() не проверяет подключение!
    // Нужно явно вызвать Ping()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        log.Fatal("Cannot connect:", err)
    }

    fmt.Println("✅ Connected to PostgreSQL!")
}
```

---

## Connection Pool (Пул соединений)

### Что это такое?

**Connection Pool** — набор готовых соединений к БД, которые переиспользуются.

```
БЕЗ пула:
Каждый запрос → Открыть соединение → Закрыть соединение
❌ Медленно

С ПУЛОМ:
[Пул из 10 соединений]
Запрос 1 → Берет из пула → Возвращает
Запрос 2 → Берет то же → Возвращает
✅ Быстро
```

### Настройка пула

```go
db, _ := sql.Open("pgx", dsn)

// Максимум открытых соединений одновременно
db.SetMaxOpenConns(25)

// Максимум простаивающих соединений
db.SetMaxIdleConns(5)

// Время жизни соединения
db.SetConnMaxLifetime(5 * time.Minute)

// Время простоя для idle-соединения
db.SetConnMaxIdleTime(1 * time.Minute)
```

### Параметры:

| Параметр | Что делает | Рекомендация |
|----------|------------|--------------|
| `MaxOpenConns` | Максимум соединений | 10-25 (dev), 20-100 (prod) |
| `MaxIdleConns` | Сколько держать "про запас" | MaxOpenConns / 2 |
| `ConnMaxLifetime` | Пересоздавать соединение через | 5-15 минут |
| `ConnMaxIdleTime` | Закрывать неиспользуемые через | 1-5 минут |

---

## Best Practices для Connection Pool

✅ **Всегда настраивайте `MaxOpenConns`** (не unlimited!)
✅ **Один `*sql.DB` на всё приложение** (создаем в `main()`)
✅ **Используйте `context.Context`** для тайм-аутов
✅ **НЕ закрывайте `db.Close()` после каждого запроса!**

❌ **НЕ создавайте `sql.Open()` в handler'ах**

---

## Repository Pattern

**Идея:** Весь код работы с БД — в отдельном слое.

### Интерфейс репозитория

```go
// Определяем интерфейс
type TodoRepository interface {
    Create(ctx context.Context, todo *model.Todo) (int64, error)
    GetByID(ctx context.Context, id int64) (*model.Todo, error)
    GetAll(ctx context.Context) ([]*model.Todo, error)
    Update(ctx context.Context, todo *model.Todo) error
    Delete(ctx context.Context, id int64) error
}

// Реализация для PostgreSQL
type PostgresTodoRepository struct {
    db *sql.DB
}

func NewTodoRepository(db *sql.DB) TodoRepository {
    return &PostgresTodoRepository{db: db}
}
```

**Зачем интерфейс?**
- Легко заменить PostgreSQL на MongoDB
- Легко создать mock для тестов

---

## DTO vs Entity

**Entity** — структура для БД
**DTO** — структура для API (JSON)

```go
// Entity (internal/model/user.go)
type User struct {
    ID           int64
    Email        string
    PasswordHash string     // ❌ НЕ отдаем в JSON!
    CreatedAt    time.Time
    DeletedAt    *time.Time // Soft delete
}

// DTO (internal/handler/dto.go)
type UserResponse struct {
    ID    int64  `json:"id"`
    Email string `json:"email"`
    // Нет password_hash!
}
```

**Зачем разные структуры?**
- **Безопасность** — не отдаем пароли
- **Гибкость** — можем менять API без изменения БД
- **Версионирование** — API v1 и v2 разные

---

## Raw SQL vs Query Builder vs ORM

| Подход | Плюсы | Минусы | Пример |
|--------|-------|--------|--------|
| **Raw SQL** | Полный контроль, производительность | Больше кода | `database/sql` |
| **Query Builder** | Меньше ошибок | Нужно учить API | `squirrel` |
| **ORM** | Быстро писать | "Магия", медленнее | `GORM` |

### Raw SQL (рекомендуется для обучения)

```go
query := "SELECT * FROM users WHERE id = $1"
db.QueryRowContext(ctx, query, userID)
```

### Query Builder

```go
sql, args, _ := squirrel.
    Select("*").
    From("users").
    Where(squirrel.Eq{"id": userID}).
    ToSql()
db.QueryContext(ctx, sql, args...)
```

### ORM

```go
var user User
db.First(&user, userID)  // Автоматически: SELECT * FROM users WHERE id = ?
```

**Для курсового проекта:** используйте **Raw SQL** (database/sql)

---

## Защита от SQL-инъекций

### ❌ ОПАСНО (конкатенация строк):
```go
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userInput)
// Если userInput = "'; DROP TABLE users; --"
// → SELECT * FROM users WHERE email = ''; DROP TABLE users; --'
```

### ✅ БЕЗОПАСНО (параметризованные запросы):
```go
query := "SELECT * FROM users WHERE email = $1"
db.QueryRowContext(ctx, query, userInput)
```

**Всегда используйте $1, $2, $3 для параметров!**

---

## Основные методы database/sql

### 1. QueryRowContext (одна строка)
```go
var name string
err := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", userID).Scan(&name)
if err == sql.ErrNoRows {
    return errors.New("user not found")
}
```

### 2. QueryContext (много строк)
```go
rows, err := db.QueryContext(ctx, "SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close()  // ⚠️ ОБЯЗАТЕЛЬНО!

for rows.Next() {
    var user User
    rows.Scan(&user.ID, &user.Name)
    // ...
}
```

### 3. ExecContext (INSERT/UPDATE/DELETE)
```go
result, err := db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", name, id)
rowsAffected, _ := result.RowsAffected()
```

---

## Обработка ошибок

```go
// 1. sql.ErrNoRows — строка не найдена (НЕ критическая ошибка)
err := db.QueryRowContext(ctx, query, id).Scan(&user)
if err == sql.ErrNoRows {
    return ErrUserNotFound
}

// 2. Constraint violation (UNIQUE, Foreign Key)
if err != nil && strings.Contains(err.Error(), "duplicate key") {
    return ErrEmailAlreadyExists
}

// 3. Таймаут
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

err := db.QueryRowContext(ctx, query).Scan(&result)
if err == context.DeadlineExceeded {
    return ErrQueryTimeout
}
```

---

## Транзакции

**Транзакция** — набор операций, которые либо все выполняются, либо все откатываются.

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()  // Откат, если что-то пойдет не так

// Несколько операций
_, err = tx.ExecContext(ctx, "INSERT INTO users ...")
if err != nil {
    return err  // Автоматический Rollback
}

_, err = tx.ExecContext(ctx, "INSERT INTO orders ...")
if err != nil {
    return err
}

// Коммит транзакции
return tx.Commit()
```

**Когда использовать транзакции?**
- Создание пользователя + его профиля
- Перевод денег между счетами
- Создание заказа + списание товара

---

## Ключевые моменты

✅ `database/sql` — интерфейс для работы с БД
✅ Используйте драйвер **pgx** для PostgreSQL
✅ **Connection Pool** — обязательно настраивайте!
✅ **Один `*sql.DB`** на всё приложение (создаем в `main()`)
✅ **Repository Pattern** — изолируем SQL в отдельный слой
✅ **Параметризованные запросы** ($1, $2) — против SQL-инъекций
✅ **Всегда используйте `context.Context`**
✅ **Закрывайте `rows.Close()`** — иначе утечка соединений

ACID - определение транзакции
A - атоморность 
С - согласованность (consistency)
I - изолированность 
D - дурабилити - устойчивость