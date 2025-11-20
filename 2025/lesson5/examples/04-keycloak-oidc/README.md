# Google OIDC Example

Пример интеграции с Google через OpenID Connect (OIDC).

## Предварительные требования

1. Google аккаунт
2. Go 1.22+

## Установка зависимостей

```bash
go mod download
```

## Настройка Google OAuth 2.0

### 1. Создание проекта в Google Cloud Console

1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект или выберите существующий
3. Включите "Google+ API":
   - Перейдите в "APIs & Services" → "Library"
   - Найдите "Google+ API" и нажмите "Enable"

### 2. Создание OAuth 2.0 Client ID

1. Перейдите в "APIs & Services" → "Credentials"
2. Нажмите "Create Credentials" → "OAuth client ID"
3. Если запросит, настройте OAuth consent screen:
   - User Type: `External` (для тестирования)
   - App name: `My Test App`
   - User support email: ваш email
   - Developer contact: ваш email
   - Нажмите "Save and Continue"
   - Scopes: оставьте по умолчанию, нажмите "Save and Continue"
   - Test users: добавьте свой email (для тестирования)
   - Нажмите "Save and Continue"
4. Создание OAuth client:
   - Application type: `Web application`
   - Name: `My Test App`
   - Authorized redirect URIs: `http://localhost:8080/callback`
   - Нажмите "Create"
5. Скопируйте **Client ID** и **Client Secret**

### 3. Обновление конфигурации

В файле `main.go` обновите:

```go
var (
    clientID     = "ваш-google-client-id.apps.googleusercontent.com"
    clientSecret = "ваш-google-client-secret"
    redirectURL  = "http://localhost:8080/callback"
    googleIssuer = "https://accounts.google.com"
)
```

## Запуск приложения

```bash
go run main.go
```

Откройте браузер: http://localhost:8080

## Тестирование

1. Откройте http://localhost:8080
2. Нажмите "Войти через Google"
3. Выберите Google аккаунт (должен быть в списке test users)
4. Разрешите доступ приложению
5. Вы будете перенаправлены обратно с профилем пользователя

## Как работает OAuth 2.0 / OIDC

1. **Пользователь** нажимает "Войти через Google"
2. **Приложение** перенаправляет на Google с параметрами:
   - client_id
   - redirect_uri
   - scope (openid, profile, email)
   - state (для защиты от CSRF)
3. **Google** показывает форму выбора аккаунта и запрос разрешений
4. **Пользователь** выбирает аккаунт и разрешает доступ
5. **Google** перенаправляет обратно на redirect_uri с `code`
6. **Приложение** обменивает `code` на токены:
   - access_token (для доступа к Google API)
   - id_token (JWT с данными пользователя)
   - refresh_token (для обновления токенов)
7. **Приложение** верифицирует id_token и извлекает данные пользователя
8. **Приложение** создает свою сессию

## API Endpoints

- `GET /` - главная страница
- `GET /login` - начало OAuth flow
- `GET /callback` - callback от Google
- `GET /profile` - профиль пользователя (требует аутентификации)
- `GET /logout` - выход
- `GET /api/profile` - JSON API с данными пользователя

## Преимущества использования Google OAuth

1. **Не нужно хранить пароли** - Google управляет аутентификацией
2. **Доверие пользователей** - Google - известный провайдер
3. **Двухфакторная аутентификация** - если включена у пользователя
4. **Безопасность** - Google управляет безопасностью аккаунтов
5. **Простота интеграции** - стандартный OAuth 2.0 / OIDC

## Когда использовать

- Публичные веб-приложения
- Нужна быстрая интеграция без собственной системы пользователей
- Хотите использовать существующие Google аккаунты пользователей
- Не нужна полная кастомизация процесса входа

## Безопасность

- Параметр `state` защищает от CSRF атак
- ID token подписан и верифицируется (JWT)
- Client secret хранится на сервере (никогда не передается клиенту)
- Используйте HTTPS в продакшене
- Настройте правильные redirect URIs в Google Console
- Проверяйте `aud` claim в id_token (должен совпадать с client_id)

## Отладка

Если возникают проблемы:

1. Проверьте, что redirect URI точно совпадает: `http://localhost:8080/callback`
2. Убедитесь, что ваш email добавлен в test users (для тестирования)
3. Проверьте, что Google+ API включен
4. Проверьте логи приложения на ошибки
