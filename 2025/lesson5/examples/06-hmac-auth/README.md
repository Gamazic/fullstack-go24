# HMAC Authentication Example

Пример аутентификации между сервисами с использованием HMAC подписи.

## Что демонстрирует

- Генерация HMAC подписи для запроса
- Проверка подписи на сервере
- Защита от replay-атак с помощью timestamp
- Сравнение с использованием `hmac.Equal` (защита от timing-атак)

## Как запустить

```bash
go run main.go
```

## Как это работает

1. **Клиент** создает подпись:
   - Берет тело запроса + текущий timestamp
   - Вычисляет HMAC(body + timestamp, secret_key)
   - Отправляет запрос с заголовками `X-Signature` и `X-Timestamp`

2. **Сервер** проверяет подпись:
   - Читает body и timestamp из запроса
   - Проверяет, что запрос не старше 5 минут
   - Вычисляет HMAC(body + timestamp, secret_key)
   - Сравнивает с полученной подписью

## Что вы увидите

Программа запустит сервер и отправит два запроса:
1. ✅ Валидный запрос с правильной подписью (успех)
2. ❌ Невалидный запрос с неправильной подписью (ошибка 401)

## Тестирование с curl

```bash
# Вручную создать подпись (пример с Python)
python3 -c "import hmac, hashlib, time; body='test'; ts=str(int(time.time())); print(f'Timestamp: {ts}'); print(f'Signature: {hmac.new(b\"super-secret-key-123\", (body+ts).encode(), hashlib.sha256).hexdigest()}')"

# Отправить запрос
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "X-Timestamp: <timestamp>" \
  -H "X-Signature: <signature>" \
  -d "test"
```
