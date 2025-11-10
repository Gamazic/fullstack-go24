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
	dsn := "postgres://postgres:postgres@localhost:5432/myapp_db?sslmode=disable"

	// Открываем соединение
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}
	defer db.Close()

	// Настраиваем Connection Pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ Cannot connect to database: %v", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL!")

	// Простой запрос: подсчет пользователей
	var userCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}

	fmt.Printf("📊 Total users in database: %d\n", userCount)

	// Получаем список всех пользователей
	fmt.Println("\n👥 Users:")
	rows, err := db.QueryContext(ctx, "SELECT id, email, created_at FROM users")
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var email string
		var createdAt time.Time

		if err := rows.Scan(&id, &email, &createdAt); err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}

		fmt.Printf("  [%d] %s (created: %s)\n", id, email, createdAt.Format("2006-01-02"))
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("❌ Rows error: %v", err)
	}

	// Получаем задачи
	fmt.Println("\n📝 Todos:")
	todoRows, err := db.QueryContext(ctx, `
		SELECT t.id, t.title, t.completed, u.email
		FROM todos t
		JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		log.Fatalf("❌ Query failed: %v", err)
	}
	defer todoRows.Close()

	for todoRows.Next() {
		var id int64
		var title string
		var completed bool
		var userEmail string

		if err := todoRows.Scan(&id, &title, &completed, &userEmail); err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}

		status := "⬜"
		if completed {
			status = "✅"
		}

		fmt.Printf("  %s [%d] %s (by %s)\n", status, id, title, userEmail)
	}

	if err := todoRows.Err(); err != nil {
		log.Fatalf("❌ Rows error: %v", err)
	}

	fmt.Println("\n✨ Done!")
}
