package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"crud-example/internal/model"
)

// TodoRepository - интерфейс для работы с задачами
type TodoRepository interface {
	Create(ctx context.Context, todo *model.Todo) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Todo, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id int64) error
}

// PostgresTodoRepository - реализация для PostgreSQL с использованием sqlx
type PostgresTodoRepository struct {
	db *sqlx.DB
}

// NewTodoRepository - создает новый репозиторий
func NewTodoRepository(db *sqlx.DB) TodoRepository {
	return &PostgresTodoRepository{db: db}
}

// Create - добавляет новую задачу в БД
func (r *PostgresTodoRepository) Create(ctx context.Context, todo *model.Todo) (int64, error) {
	query := `
		INSERT INTO todos (user_id, title, description, completed)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	// sqlx.QueryRowxContext возвращает *sqlx.Row с методом StructScan
	// но для RETURNING проще использовать обычный Scan
	err := r.db.QueryRowContext(
		ctx,
		query,
		todo.UserID,
		todo.Title,
		todo.Description,
		todo.Completed,
	).Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		return 0, err
	}

	return todo.ID, nil
}

// GetByID - получает задачу по ID
// Используем sqlx.Get для автоматического маппинга в структуру
func (r *PostgresTodoRepository) GetByID(ctx context.Context, id int64) (*model.Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	todo := &model.Todo{}

	// sqlx.Get автоматически делает Scan в структуру благодаря тегам `db`
	err := r.db.GetContext(ctx, todo, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("todo not found")
		}
		return nil, err
	}

	return todo, nil
}

// GetAllByUserID - получает все задачи пользователя
// Используем sqlx.Select для автоматического маппинга slice
func (r *PostgresTodoRepository) GetAllByUserID(ctx context.Context, userID int64) ([]*model.Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var todos []*model.Todo

	// sqlx.Select автоматически создает slice и заполняет его
	// НЕ НУЖНО вручную делать rows.Scan() в цикле!
	err := r.db.SelectContext(ctx, &todos, query, userID)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

// Update - обновляет задачу
func (r *PostgresTodoRepository) Update(ctx context.Context, todo *model.Todo) error {
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

// UpdateNamed - альтернативный способ обновления с использованием Named queries
// Это удобнее для больших структур с множеством полей
func (r *PostgresTodoRepository) UpdateNamed(ctx context.Context, todo *model.Todo) error {
	query := `
		UPDATE todos
		SET title = :title,
		    description = :description,
		    completed = :completed,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`

	// NamedExecContext использует теги `db` из структуры
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

// Delete - удаляет задачу
func (r *PostgresTodoRepository) Delete(ctx context.Context, id int64) error {
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

// BatchInsert - пример массовой вставки с sqlx
func (r *PostgresTodoRepository) BatchInsert(ctx context.Context, todos []*model.Todo) error {
	query := `
		INSERT INTO todos (user_id, title, description, completed)
		VALUES (:user_id, :title, :description, :completed)
	`

	// NamedExec может принимать slice структур
	_, err := r.db.NamedExecContext(ctx, query, todos)
	return err
}

// GetWithRawSQL - пример использования sqlx.In для запросов с IN (...)
func (r *PostgresTodoRepository) GetByIDs(ctx context.Context, ids []int64) ([]*model.Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id IN (?)
		ORDER BY created_at DESC
	`

	// sqlx.In преобразует ? в $1, $2, $3 для PostgreSQL
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}

	// Rebind для правильных placeholder'ов PostgreSQL ($1, $2...)

	// select ... where id IN ($1, $2, $3, ...) ..
	query = r.db.Rebind(query)

	var todos []*model.Todo
	err = r.db.SelectContext(ctx, &todos, query, args...)
	if err != nil {
		return nil, err
	}

	return todos, nil
}
