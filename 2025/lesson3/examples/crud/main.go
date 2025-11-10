package main

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"crud-example/internal/handler"
	"crud-example/internal/repository"
	"crud-example/internal/service"
)

func main() {
	// 1. Подключаемся к БД с использованием sqlx
	dsn := "postgres://postgres:postgres@localhost:5432/myapp_db?sslmode=disable"

	// sqlx.Connect автоматически проверяет подключение (делает Ping)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 2. Настраиваем Connection Pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	log.Println("✅ Connected to PostgreSQL with sqlx!")

	// 3. Проверяем подключение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ Cannot ping database: %v", err)
	}

	// 4. Создаем слои приложения
	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoService)

	// 5. Регистрируем маршруты
	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			todoHandler.CreateTodo(w, r)
		case http.MethodGet:
			todoHandler.GetTodos(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/todos/get", todoHandler.GetTodo)
	http.HandleFunc("/todos/complete", todoHandler.CompleteTodo)
	http.HandleFunc("/todos/delete", todoHandler.DeleteTodo)

	// 6. Запускаем сервер
	port := ":8080"
	log.Printf("🚀 Server is running on http://localhost%s\n", port)
	log.Println("\n📝 Доступные эндпоинты:")
	log.Println("  POST   /todos              - Создать задачу")
	log.Println("  GET    /todos              - Список задач")
	log.Println("  GET    /todos/get?id=1     - Получить задачу")
	log.Println("  POST   /todos/complete?id=1 - Отметить выполненной")
	log.Println("  DELETE /todos/delete?id=1  - Удалить задачу")
	log.Println("\n💡 Преимущества sqlx:")
	log.Println("  ✅ Автоматический маппинг с помощью тегов `db`")
	log.Println("  ✅ db.Get() / db.Select() вместо ручного Scan()")
	log.Println("  ✅ Named queries (:name вместо $1, $2...)")
	log.Println("  ✅ sqlx.In() для работы с IN (...)")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
