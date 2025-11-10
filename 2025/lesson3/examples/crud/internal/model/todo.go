package model

import "time"

// Todo - модель задачи (Entity)
// Теги `db` используются библиотекой sqlx для автоматического маппинга
type Todo struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Completed   bool      `db:"completed"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
