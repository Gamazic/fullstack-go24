# RBAC (Role-Based Access Control) Example

Пример управления доступом на основе ролей.

## Установка зависимостей

```bash
go mod download
```

## Запуск

```bash
go run main.go
```

## Роли и доступы

### Роли
- **admin** - полный доступ ко всем ресурсам
- **moderator** - доступ к модерации контента
- **user** - базовый доступ

### Предустановленные пользователи

| Email | Password | Role |
|-------|----------|------|
| admin@example.com | admin123 | admin |
| moderator@example.com | mod123 | moderator |
| user@example.com | user123 | user |

## Endpoints и требуемые роли

| Endpoint | Роль | Описание |
|----------|------|----------|
| `/public` | - | Публичный endpoint |
| `/login` | - | Логин |
| `/profile` | любая | Профиль пользователя |
| `/admin/users` | admin | Список всех пользователей |
| `/admin/delete-user` | admin | Удаление пользователя |
| `/moderation` | admin, moderator | Панель модерации |

## Тестирование

### 1. Логин как администратор

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'

# Сохраните токен
TOKEN="полученный_токен"
```

### 2. Доступ к админ панели

```bash
curl http://localhost:8080/admin/users \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Логин как обычный пользователь

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"user123"}'

USER_TOKEN="полученный_токен"
```

### 4. Попытка доступа к админ панели (403 Forbidden)

```bash
curl http://localhost:8080/admin/users \
  -H "Authorization: Bearer $USER_TOKEN"
```

### 5. Доступ к профилю (разрешен для всех ролей)

```bash
curl http://localhost:8080/profile \
  -H "Authorization: Bearer $USER_TOKEN"
```

### 6. Доступ к модерации (admin и moderator)

```bash
# Логин как модератор
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@example.com","password":"mod123"}'

MOD_TOKEN="полученный_токен"

# Доступ к модерации
curl http://localhost:8080/moderation \
  -H "Authorization: Bearer $MOD_TOKEN"
```

### 7. Удаление пользователя (только admin)

```bash
curl -X DELETE http://localhost:8080/admin/delete-user \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

## Автоматизированные тесты

```bash
# Получение токена админа
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' | jq -r '.token')

# Получение токена пользователя
USER_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"user123"}' | jq -r '.token')

# Тест 1: Админ может получить список пользователей
curl -s http://localhost:8080/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Тест 2: Обычный пользователь не может (403)
curl -s http://localhost:8080/admin/users \
  -H "Authorization: Bearer $USER_TOKEN"

# Тест 3: Все могут получить свой профиль
curl -s http://localhost:8080/profile \
  -H "Authorization: Bearer $USER_TOKEN" | jq
```

## Как работает RBAC

1. **Аутентификация**: Пользователь логинится и получает JWT токен
2. **Токен содержит роль**: В claims токена сохранена роль пользователя
3. **Middleware проверяет роль**: При запросе middleware извлекает роль из токена
4. **Сравнение с требуемой ролью**: Проверяется соответствие роли пользователя требованиям endpoint'а
5. **Доступ или отказ**: 200 OK или 403 Forbidden

## Структура middleware

```go
// Проверка конкретной роли
requireRole(RoleAdmin, handler)

// Проверка одной из ролей
requireAnyRole([]Role{RoleAdmin, RoleModerator}, handler)
```

## Расширение системы ролей

Можно добавить:
- **Permissions** - детальные разрешения (read, write, delete)
- **Hierarchy** - иерархия ролей (admin включает moderator)
- **Dynamic roles** - роли из БД
- **Resource-based** - права на конкретные ресурсы
- **Time-based** - временные роли

## Использование в продакшене

- Храните роли в БД
- Кешируйте роли пользователей (Redis)
- Используйте более сложную систему разрешений (RBAC + ABAC)
- Логируйте попытки несанкционированного доступа
- Регулярно проверяйте актуальность ролей
