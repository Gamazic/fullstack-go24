package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Ждем пока БД поднимется
	time.Sleep(3 * time.Second)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@db:5432/myapp?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем таблицу
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/items", itemsHandler)

	log.Println("API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, name FROM items")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var items []Item
		for rows.Next() {
			var item Item
			rows.Scan(&item.ID, &item.Name)
			items = append(items, item)
		}
		if items == nil {
			items = []Item{}
		}
		json.NewEncoder(w).Encode(items)

	case http.MethodPost:
		var item Item
		json.NewDecoder(r.Body).Decode(&item)
		err := db.QueryRow("INSERT INTO items (name) VALUES ($1) RETURNING id", item.Name).Scan(&item.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

