# Basic Auth Example

Пример базовой HTTP аутентификации.

## Запуск

```bash
go run main.go
```

## Тестирование

```bash
# Публичный endpoint
curl http://localhost:8080/public

# Защищенный endpoint с правильными credentials
curl -u admin:secret http://localhost:8080/protected

# С неправильными credentials
curl -u admin:wrong http://localhost:8080/protected

# Или с заголовком Authorization
curl -H "Authorization: Basic YWRtaW46c2VjcmV0" http://localhost:8080/protected
```

## Как работает

1. Клиент отправляет заголовок `Authorization: Basic <base64(username:password)>`
2. Сервер декодирует base64 и проверяет username и password
3. Если credentials верны - пропускает запрос дальше
4. Если нет - возвращает 401 Unauthorized

## Плюсы и минусы

**Плюсы:**
- Очень простая реализация
- Не требует состояния на сервере
- Стандартный HTTP метод

**Минусы:**
- Credentials передаются в каждом запросе
- Обязательно нужен HTTPS
- Нет возможности "выйти" из системы
- Браузер кеширует credentials
