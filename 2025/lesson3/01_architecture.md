# Часть 1: Архитектура приложения

## Зачем нужна база данных?

**Проблема:** Хранение данных в памяти (переменные, массивы, мапы) → данные теряются после перезапуска

**Решение:** База данных — постоянное хранилище

| Хранилище | Скорость | Персистентность |
|-----------|----------|-----------------|
| RAM (память) | Очень быстро | Нет (теряется при перезапуске) |
| База данных | Быстро | Да (сохраняется навсегда) |
| Файлы | Медленно | Да |

---

## MVC архитектура

```
MVC = Model-View-Controller

Model (Модель)       → Данные и бизнес-логика
View (Представление) → Отображение данных (фронтенд)
Controller           → Обработка запросов
```

**В Go мы используем слоистую архитектуру:**

```
HTTP Request
    ↓
┌─────────────┐
│  HANDLER    │  ← Обработка HTTP (парсинг JSON, валидация)
└──────┬──────┘
       ↓
┌─────────────┐
│  SERVICE    │  ← Бизнес-логика
└──────┬──────┘
       ↓
┌─────────────┐
│ REPOSITORY  │  ← Работа с БД (SQL-запросы)
└──────┬──────┘
       ↓
   Database
```

---

## Слои приложения

### 1. Handler (Контроллер)

**Ответственность:**
- Принимает HTTP-запрос
- Парсит JSON/параметры
- Валидирует входные данные
- Вызывает Service
- Возвращает HTTP-ответ

**Пример:**
```go
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
    // 1. Парсим JSON
    var req CreateTodoRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Валидация
    if req.Title == "" {
        http.Error(w, "Title is required", 400)
        return
    }

    // 3. Вызываем сервис
    todo, err := h.service.CreateTodo(r.Context(), req.Title)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // 4. Возвращаем ответ
    json.NewEncoder(w).Encode(todo)
}
```

---

### 2. Service (Бизнес-логика)

**Ответственность:**
- Реализует бизнес-правила
- Координирует работу нескольких репозиториев
- НЕ знает про HTTP

**Пример:**
```go
func (s *TodoService) CreateTodo(ctx context.Context, title string) (*model.Todo, error) {
    // Бизнес-логика
    if len(title) > 100 {
        return nil, errors.New("title too long")
    }

    todo := &model.Todo{
        Title:     title,
        CreatedAt: time.Now(),
    }

    // Вызываем репозиторий
    id, err := s.repo.Create(ctx, todo)
    if err != nil {
        return nil, err
    }

    todo.ID = id
    return todo, nil
}
```

---

### 3. Repository (Доступ к данным)

**Ответственность:**
- CRUD операции с БД
- SQL-запросы
- НЕ содержит бизнес-логику

**Пример:**
```go
func (r *TodoRepository) Create(ctx context.Context, todo *model.Todo) (int64, error) {
    query := `
        INSERT INTO todos (title, created_at)
        VALUES ($1, $2)
        RETURNING id
    `

    var id int64
    err := r.db.QueryRowContext(ctx, query, todo.Title, todo.CreatedAt).Scan(&id)
    return id, err
}
```

---

## Зачем разделять на слои?

✅ **Separation of Concerns** — каждый слой отвечает за свою задачу
✅ **Тестируемость** — можно тестировать бизнес-логику без БД
✅ **Переиспользование кода** — Service можно вызывать из разных Handler'ов
✅ **Легко менять** — поменять БД → правим только Repository

---

## Правила слоистой архитектуры

❌ **НЕ писать SQL в Handler**
❌ **НЕ писать бизнес-логику в Repository**
❌ **НЕ работать с `http.Request` в Service**

✅ **Handler** → "Что пришло от клиента?"
✅ **Service** → "Что делать с этими данными?"
✅ **Repository** → "Как сохранить/получить из БД?"

---

## Структура проекта

```
/your-project
├── cmd/
│   └── server/
│       └── main.go          # Точка входа
├── internal/
│   ├── handler/             # HTTP handlers
│   │   └── todo_handler.go
│   ├── service/             # Бизнес-логика
│   │   └── todo_service.go
│   ├── repository/          # Работа с БД
│   │   └── todo_repository.go
│   └── model/               # Структуры данных
│       └── todo.go
├── migrations/              # SQL миграции
│   └── 001_create_todos.sql
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## Примеры для ваших проектов

### Интернет-магазин: Добавление товара в корзину

```
POST /api/cart/items

Handler:
  - Парсит JSON: { "product_id": 123, "quantity": 2 }
  - Проверяет quantity > 0
  - Получает user_id из JWT

Service:
  - Проверяет, существует ли товар (ProductRepository)
  - Проверяет наличие на складе (бизнес-логика)
  - Добавляет в корзину (CartRepository)

Repository:
  - INSERT INTO cart_items (user_id, product_id, quantity) VALUES (...)
```

### Блог: Создание поста

```
POST /api/posts

Handler → парсит заголовок, контент
Service → проверяет права, создает slug
Repository → INSERT INTO posts
```

---

## Ключевые моменты

- Handler ← HTTP-запросы
- Service ← Бизнес-логика
- Repository ← SQL-запросы
- Один слой не должен знать про детали другого
- Используем интерфейсы для гибкости
