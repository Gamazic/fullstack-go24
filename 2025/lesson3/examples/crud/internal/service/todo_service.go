package service

import (
	"context"
	"errors"
	"time"

	"crud-example/internal/model"
)

type todoRepository interface {
	Create(ctx context.Context, todo *model.Todo) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Todo, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]*model.Todo, error)
	Update(ctx context.Context, todo *model.Todo) error
	Delete(ctx context.Context, id int64) error
}

// TodoService - бизнес-логика для задач
type TodoService struct {
	repo todoRepository
}

// NewTodoService - создает новый сервис
func NewTodoService(repo todoRepository) *TodoService {
	return &TodoService{repo: repo}
}

// CreateTodo - создает новую задачу с валидацией
func (s *TodoService) CreateTodo(ctx context.Context, userID int64, title, description string) (*model.Todo, error) {
	// Бизнес-логика: валидация
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	if len(title) > 255 {
		return nil, errors.New("title too long (max 255 characters)")
	}

	// Создаем модель
	todo := &model.Todo{
		UserID:      userID,
		Title:       title,
		Description: description,
		Completed:   false,
	}

	// Сохраняем через репозиторий
	id, err := s.repo.Create(ctx, todo)
	if err != nil {
		return nil, err
	}

	todo.ID = id
	return todo, nil
}

// GetTodoByID - получает задачу по ID
// Важно: данная функция скорее антипрактик (и другие функции где просто вызов другой функции). В ней ничего не происходит кроме вызова
// другой такой же функции. Лучше всего удалить эту функцию и в хэндлере вызывать напрямую репозиторий
// смысл будет такой же.
func (s *TodoService) GetTodoByID(ctx context.Context, id int64) (*model.Todo, error) {
	return s.repo.GetByID(ctx, id)
}

// GetUserTodos - получает все задачи пользователя
func (s *TodoService) GetUserTodos(ctx context.Context, userID int64) ([]*model.Todo, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

// CompleteTodo - отмечает задачу как выполненную
func (s *TodoService) CompleteTodo(ctx context.Context, todoID int64) error {
	// Получаем задачу
	todo, err := s.repo.GetByID(ctx, todoID)
	if err != nil {
		return err
	}

	// Меняем статус
	todo.Completed = true
	todo.UpdatedAt = time.Now()

	// Сохраняем
	return s.repo.Update(ctx, todo)
}

// UpdateTodo - обновляет задачу
func (s *TodoService) UpdateTodo(ctx context.Context, id int64, title, description string) error {
	// Валидация
	if title == "" {
		return errors.New("title cannot be empty")
	}

	if len(title) > 255 {
		return errors.New("title too long")
	}

	// Получаем задачу
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Обновляем поля
	todo.Title = title
	todo.Description = description

	// Сохраняем
	return s.repo.Update(ctx, todo)
}

// DeleteTodo - удаляет задачу
func (s *TodoService) DeleteTodo(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
