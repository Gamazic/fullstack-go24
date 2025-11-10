# Лекция 3: Базы данных и работа с PostgreSQL в Go

## Содержание лекции

### Теоретические материалы (для студентов):

1. **[01_architecture.md](01_architecture.md)** — Архитектура приложения
   - Зачем нужна БД
   - MVC архитектура
   - Слоистая архитектура (Handler → Service → Repository)
   - Структура проекта

2. **[02_database_basics.md](02_database_basics.md)** — Основы работы с БД в Go
   - Стандартная библиотека database/sql
   - Драйверы (pq, pgx)
   - Connection Pool и его настройка
   - Repository Pattern
   - DTO vs Entity
   - Raw SQL vs Query Builder vs ORM
   - Защита от SQL-инъекций

3. **[03_practice_postgresql.md](03_practice_postgresql.md)** — Практика с PostgreSQL
   - Запуск PostgreSQL через Docker
   - CRUD операции (Create, Read, Update, Delete)
   - Связи между таблицами (One-to-Many, Many-to-Many)
   - Best practices
   - Типичные ошибки

4. **[SQLX_GUIDE.md](SQLX_GUIDE.md)** — Руководство по sqlx
   - Что такое sqlx и зачем он нужен
   - Get() / Select() вместо ручного Scan()
   - Named queries (`:name` вместо `$1, $2`)
   - Работа с IN (...)
   - Сравнение с database/sql

---

## Практические примеры

### 1. Docker и миграции

- **[docker-compose.yml](docker-compose.yml)** — конфигурация PostgreSQL
- **[migrations/001_init.sql](migrations/001_init.sql)** — создание таблиц и тестовые данные
- **[sql_examples.sql](sql_examples.sql)** — примеры SQL-запросов для практики

### 2. Простой пример подключения к БД

📁 **[examples/basic/](examples/basic/)** — базовое подключение к PostgreSQL

- Настройка Connection Pool
- Простые SELECT запросы
- JOIN между таблицами
- Работа с rows.Close()

```bash
cd examples/basic
go run main.go
```

### 3. Полный пример с архитектурой (sqlx)

📁 **[examples/crud/](examples/crud/)** — полноценное приложение с HTTP API и **sqlx**

Архитектура: **Handler → Service → Repository (sqlx)**

- **sqlx вместо database/sql** для удобства
- CRUD операции через HTTP
- Автоматический маппинг с тегами `db`
- `Get()` и `Select()` вместо ручного Scan()
- Named queries (`:name` вместо `$1, $2`)
- Интерфейсы для репозиториев
- DTO для API

```bash
cd examples/crud
go run main.go
# Сервер запустится на http://localhost:8080
```

**Примеры запросов:**

```bash
# Создать задачу
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Изучить Go", "description": "Лекция 3"}'

# Получить список задач
curl http://localhost:8080/todos

# Отметить как выполненную
curl -X POST http://localhost:8080/todos/complete?id=1
```

---

## Быстрый старт

### 1. Запустите PostgreSQL

```bash
docker-compose up -d
```

Проверка:

```bash
docker-compose ps
docker exec -it lesson3_db psql -U postgres -d myapp_db
```

Внутри psql:

```sql
\dt                    -- Список таблиц
SELECT * FROM users;   -- Проверка данных
\q                     -- Выход
```

### 2. Попробуйте SQL-запросы вручную

```bash
docker exec -i lesson3_db psql -U postgres -d myapp_db < sql_examples.sql
```

Или выполняйте запросы по одному из [sql_examples.sql](sql_examples.sql).

### 3. Запустите примеры на Go

**Простой пример:**

```bash
cd examples/basic
go mod download
go run main.go
```

**CRUD пример:**

```bash
cd examples/crud
go mod download
go run main.go
```

---

## Структура проекта

```
lesson3/
├── README.md                      # Этот файл
├── plan.md                        # Исходный план лекции
│
├── 01_architecture.md             # Теория: архитектура
├── 02_database_basics.md          # Теория: database/sql
├── 03_practice_postgresql.md      # Теория: практика
│
├── docker-compose.yml             # PostgreSQL
├── migrations/
│   └── 001_init.sql               # Создание таблиц
├── sql_examples.sql               # Примеры SQL-запросов
│
└── examples/
    ├── basic/                     # Простой пример
    │   ├── main.go
    │   ├── go.mod
    │   └── README.md
    │
    └── crud/                      # CRUD с архитектурой
        ├── main.go
        ├── go.mod
        ├── README.md
        └── internal/
            ├── model/             # Entity
            ├── repository/        # БД
            ├── service/           # Бизнес-логика
            └── handler/           # HTTP + DTO
```

---

## Домашнее задание #5

**Задача:** Спроектировать модель данных для вашего проекта

**Требования:**

1. Минимум 2 таблицы + `users`
2. SQL-миграции (CREATE TABLE)
3. docker-compose.yml для PostgreSQL
4. Описание связей между таблицами (диаграмма или текст)

**Примеры проектов:**

### Интернет-магазин:

```sql
users (id, email, password_hash)
products (id, name, price, stock)
orders (id, user_id, total, status, created_at)
order_items (id, order_id, product_id, quantity, price)
```

Связи:
- users (1) → (N) orders
- orders (1) → (N) order_items
- products (1) → (N) order_items

### Блог:

```sql
users (id, username, email)
posts (id, user_id, title, content, created_at)
comments (id, post_id, user_id, content, created_at)
tags (id, name)
post_tags (post_id, tag_id)
```

Связи:
- users (1) → (N) posts
- posts (1) → (N) comments
- posts (N) ↔ (N) tags (через post_tags)

### Система бронирования:

```sql
users (id, email, phone)
rooms (id, name, capacity, price_per_night)
bookings (id, user_id, room_id, check_in, check_out, total)
```

Связи:
- users (1) → (N) bookings
- rooms (1) → (N) bookings

---

## Полезные команды

### Docker:

```bash
docker-compose up -d              # Запустить PostgreSQL
docker-compose ps                 # Проверить статус
docker-compose logs postgres      # Логи
docker-compose stop               # Остановить
docker-compose start              # Запустить снова
docker-compose down -v            # Удалить контейнер И данные
```

### PostgreSQL (psql):

```bash
# Подключение
docker exec -it lesson3_db psql -U postgres -d myapp_db

# Внутри psql:
\dt                               # Список таблиц
\d table_name                     # Описание таблицы
\du                               # Список пользователей
\l                                # Список баз данных
\q                                # Выход
```

### Go:

```bash
go mod init myproject             # Инициализация модуля
go mod download                   # Скачать зависимости
go get github.com/jackc/pgx/v5    # Установить пакет
go run main.go                    # Запустить программу
```

---

## Ключевые моменты лекции

✅ **Архитектура:** Handler → Service → Repository
✅ **database/sql** — стандартная библиотека для SQL
✅ **pgx** — современный драйвер для PostgreSQL
✅ **Connection Pool** — обязательно настраивать!
✅ **Параметризованные запросы** ($1, $2) — защита от SQL-инъекций
✅ **defer rows.Close()** — всегда закрывать rows!
✅ **Интерфейсы** для репозиториев — гибкость и тестируемость
✅ **DTO ≠ Entity** — разные структуры для БД и API

---

## Полезные ссылки

- [PostgreSQL Tutorial](https://www.postgresql.org/docs/current/tutorial.html)
- [database/sql Documentation](https://pkg.go.dev/database/sql)
- [pgx GitHub](https://github.com/jackc/pgx)
- [Go by Example: Database](https://gobyexample.com/)
- [Three Dots Labs - Repository Pattern](https://threedots.tech/post/repository-pattern-in-go/)
- [Alex Edwards - Practical Persistence](https://www.alexedwards.net/blog/practical-persistence-sql)

---

## Вопросы для самопроверки

1. Зачем разделять код на Handler, Service и Repository?
2. Что такое Connection Pool и зачем он нужен?
3. В чем разница между Entity и DTO?
4. Как защититься от SQL-инъекций в Go?
5. Почему нужно закрывать `rows.Close()`?
6. Когда использовать транзакции?
7. Что возвращает `sql.ErrNoRows` и как его обрабатывать?
8. Зачем нужен интерфейс для репозитория?

---

## Следующая лекция

**Лекция 4:** Аутентификация и авторизация

- JWT токены
- Middleware
- Хеширование паролей (bcrypt)
- Сессии vs Токены
- CORS

---

**Удачи с домашним заданием! 🚀**
