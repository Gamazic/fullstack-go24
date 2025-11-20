# Session Auth Example

Пример аутентификации через сессии с использованием cookies.

## Запуск

```bash
go run main.go
```

## Тестирование

```bash
# 1. Регистрация нового пользователя
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test User"}'

# 2. Логин (сохраняем cookie в файл)
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  -c cookies.txt

# 3. Доступ к защищенному endpoint с cookie
curl http://localhost:8080/profile -b cookies.txt

# 4. Выход
curl -X POST http://localhost:8080/logout -b cookies.txt

# 5. Попытка доступа после выхода (должна вернуть 401)
curl http://localhost:8080/profile -b cookies.txt
```

## Как работает

1. Пользователь отправляет логин и пароль на `/login`
2. Сервер проверяет credentials
3. Если верны - создает случайный session_id и сохраняет в памяти (в продакшене - Redis)
4. Отправляет session_id клиенту в cookie
5. Клиент автоматически отправляет cookie в каждом запросе
6. Сервер проверяет session_id и извлекает user_id

## Предустановленные пользователи

- Email: `user@example.com`
- Password: `password123`

## Плюсы и минусы

**Плюсы:**
- Credentials передаются только один раз
- Можно легко отозвать сессию
- Сервер полностью контролирует сессии

**Минусы:**
- Требует хранилище для сессий (память, Redis, БД)
- Сложнее масштабировать (нужно shared storage)
- Нужна защита от CSRF атак
