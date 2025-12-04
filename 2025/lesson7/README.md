# Лекция 7: Развертывание Backend + Frontend

## 1. Подключение к серверу

```bash
# Генерация SSH ключа (если нет)
ssh-keygen -t ed25519 -C "your_email@example.com"

# Копирование ключа на сервер
ssh-copy-id user@your-server-ip

# Подключение
ssh user@your-server-ip
```

### Копирование файлов на сервер
```bash
# Копируем папку example на сервер
rsync -avz --progress ./example/ user@your-server-ip:~/project/

# Если нужно исключить файлы (например, node_modules)
rsync -avz --progress --exclude 'node_modules' ./example/ user@your-server-ip:~/project/
```

---

## 2. Установка Docker на Ubuntu

```bash
# Обновляем пакеты
sudo apt update && sudo apt upgrade -y

# Устанавливаем зависимости
sudo apt install -y ca-certificates curl gnupg

# Добавляем GPG ключ Docker
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Добавляем репозиторий
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Устанавливаем Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Добавляем пользователя в группу docker (чтобы не писать sudo)
sudo usermod -aG docker $USER
newgrp docker

# Проверяем
docker --version
docker compose version
```

---

## 3. Структура проекта

```
example/
├── api/
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   └── index.html
├── nginx/
│   └── nginx.conf
└── docker-compose.yml
```

---

## 4. Запуск API + PostgreSQL

```bash
cd example

# Запускаем
docker compose up -d --build

# Смотрим логи
docker compose logs -f api

# Тестируем API
curl http://localhost:8080/api/health

# Создаем item
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Item"}'

# Получаем список
curl http://localhost:8080/api/items
```

---

## 5. Nginx: что это и зачем

**Nginx** — веб-сервер и reverse proxy.

Основные функции:
- **Раздача статики** (HTML, CSS, JS, изображения)
- **Reverse proxy** — проксирование запросов на backend
- **Load balancing** — балансировка нагрузки
- **SSL termination** — обработка HTTPS
- **Кеширование**

### Почему не раздавать статику напрямую из Go?
- Nginx оптимизирован для статики (sendfile, кеширование)
- Разделение ответственности
- Легко масштабировать

---

## 6. Базовая конфигурация Nginx

### Структура конфига
```nginx
worker_processes auto;

events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;

        location / {
            # Обработка путей
        }
    }
}
```

### Основные директивы

| Директива | Описание |
|-----------|----------|
| `listen` | Порт |
| `server_name` | Домен |
| `root` | Корневая папка статики |
| `index` | Файл по умолчанию |
| `location` | Правило для пути |
| `proxy_pass` | Проксирование на backend |

---

## 7. Раздаем статику (без backend)

Смотри `nginx/nginx.conf` — сначала только `location /` для статики.

```bash
docker compose up -d --build

# Открываем в браузере http://your-server-ip
# Страница загружается, но кнопка не работает!
# В консоли браузера: Failed to fetch /api/items
```

---

## 8. Добавляем proxy на backend

Добавляем в nginx.conf:
```nginx
location /api/ {
    proxy_pass http://api:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

```bash
docker compose up -d --build

# Теперь страничка работает!
curl http://localhost/api/health
```

---

## 9. Полезные команды

```bash
docker compose up -d --build    # Запуск
docker compose logs -f          # Логи
docker compose logs -f api      # Логи одного сервиса
docker compose down             # Остановка
docker compose down -v          # Остановка + удаление volumes
docker compose restart nginx    # Перезапуск сервиса
docker compose ps               # Статус
```

---

## 10. Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                         Internet                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Nginx (:80)                            │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │   /                 │    │   /api/*                    │ │
│  │   Static files      │    │   proxy_pass → api:8080     │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      API (:8080)                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   PostgreSQL (:5432)                        │
└─────────────────────────────────────────────────────────────┘
```

---

## 11. Как сделать полноценный сайт?

- **HTTPS** — Let's Encrypt + certbot
- **Домен** — привязать доменное имя
- **CI/CD** — автоматическое развертывание
- **Мониторинг** — логи, метрики
- **SEO** - показ сайта в поисковиках
