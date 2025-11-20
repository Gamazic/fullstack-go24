# JWT Auth Example

Пример аутентификации через JWT (JSON Web Tokens).

## Установка зависимостей

```bash
go mod download
```

## Запуск

```bash
go run main.go
```

## Тестирование

```bash
# 1. Логин (получить токен)
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Сохраните токен из ответа в переменную
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 2. Доступ к профилю с токеном
curl http://localhost:8080/profile \
  -H "Authorization: Bearer $TOKEN"

# 3. Проверка токена
curl http://localhost:8080/verify \
  -H "Authorization: Bearer $TOKEN"

# Или одной командой (требует jq):
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' | jq -r '.token')

curl http://localhost:8080/profile -H "Authorization: Bearer $TOKEN"
```

## Как работает

1. Пользователь отправляет логин и пароль на `/login`
2. Сервер проверяет credentials
3. Если верны - создает JWT токен с данными пользователя
4. Токен подписывается секретным ключом
5. Клиент сохраняет токен (localStorage, cookie)
6. Клиент отправляет токен в заголовке `Authorization: Bearer <token>`
7. Сервер проверяет подпись токена и извлекает данные

## Структура JWT токена

JWT состоит из трех частей, разделенных точкой:
```
header.payload.signature
```

**Header:**
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload (claims):**
```json
{
  "user_id": 1,
  "email": "user@example.com",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Signature:**
```
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret
)
```

## Предустановленные пользователи

- Email: `user@example.com`
- Password: `password123`

## Плюсы и минусы

**Плюсы:**
- Stateless - сервер не хранит сессии
- Легко масштабируется
- Токен содержит данные пользователя
- Подходит для микросервисов

**Минусы:**
- Нельзя отозвать токен до истечения срока
- Больший размер по сравнению с session_id
- Нужно хранить секретный ключ
- Токен может быть украден (нужен HTTPS)

## Безопасность

- Всегда используйте HTTPS
- Храните секретный ключ в переменных окружения
- Устанавливайте разумный срок жизни токена (не больше 24 часов)
- Используйте Refresh Tokens для длительных сессий
- Не храните чувствительные данные в токене
