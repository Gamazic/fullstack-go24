package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crud-example/internal/service"
)

// CreateTodoRequest - DTO для создания задачи
type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// TodoResponse - DTO для ответа с задачей
type TodoResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

// TodoHandler - HTTP handler для задач
type TodoHandler struct {
	service *service.TodoService
}

// NewTodoHandler - создает новый handler
func NewTodoHandler(service *service.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// CreateTodo - POST /todos - создание задачи
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим JSON
	var req CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 2. Валидация (базовая, остальная в сервисе)
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// 3. В реальном приложении userID берется из JWT
	// Для примера используем userID = 1
	userID := int64(1)

	// 4. Вызываем сервис
	todo, err := h.service.CreateTodo(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Конвертируем Entity → DTO
	response := TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 6. Возвращаем JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTodos - GET /todos - список задач пользователя
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	// В реальном приложении userID из JWT
	userID := int64(1)

	todos, err := h.service.GetUserTodos(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Конвертируем в DTO
	var response []TodoResponse
	for _, todo := range todos {
		response = append(response, TodoResponse{
			ID:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTodo - GET /todos/{id} - получить задачу по ID
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL (в реальном проекте используйте router типа gorilla/mux)
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todo, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CompleteTodo - POST /todos/{id}/complete - отметить как выполненную
func (h *TodoHandler) CompleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if err := h.service.CompleteTodo(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Todo completed"}`))
}

// DeleteTodo - DELETE /todos/{id} - удалить задачу
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTodo(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Todo deleted"}`))
}
