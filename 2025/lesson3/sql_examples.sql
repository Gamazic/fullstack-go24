-- ========================================
-- ПРИМЕРЫ SQL-ЗАПРОСОВ ДЛЯ ПРАКТИКИ
-- ========================================

-- Как запустить этот файл:
-- 1. Запустите PostgreSQL: docker-compose up -d
-- 2. Подключитесь к БД: docker exec -it lesson3_db psql -U postgres -d myapp_db
-- 3. Выполняйте команды по одной, копируя из этого файла

-- Или выполните весь файл разом:
-- docker exec -i lesson3_db psql -U postgres -d myapp_db < sql_examples.sql

-- ========================================
-- 1. СОЗДАНИЕ ТАБЛИЦЫ (CREATE TABLE)
-- ========================================

-- Создадим таблицу для продуктов
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Проверяем, что таблица создана:
-- \dt (список всех таблиц)
-- \d products (описание таблицы products)

-- ========================================
-- 2. ВСТАВКА ДАННЫХ (INSERT)
-- ========================================

-- Вставляем один товар
INSERT INTO products (name, price, stock)
VALUES ('Ноутбук', 75000.00, 10);

-- Вставляем несколько товаров разом
INSERT INTO products (name, price, stock) VALUES
    ('Мышь', 1500.00, 50),
    ('Клавиатура', 3500.00, 30),
    ('Монитор', 25000.00, 15),
    ('Наушники', 5000.00, 20);

-- Вставка с возвратом ID (RETURNING)
INSERT INTO products (name, price, stock)
VALUES ('Веб-камера', 4000.00, 8)
RETURNING id, name, created_at;

-- ========================================
-- 3. ЧТЕНИЕ ДАННЫХ (SELECT)
-- ========================================

-- Выбрать все товары
SELECT * FROM products;

-- Выбрать конкретные поля
SELECT id, name, price FROM products;

-- Выбрать с условием (WHERE)
SELECT * FROM products WHERE price > 5000;

-- Выбрать с сортировкой (ORDER BY)
SELECT * FROM products ORDER BY price DESC;

-- Выбрать с лимитом
SELECT * FROM products ORDER BY price DESC LIMIT 3;

-- Подсчет количества
SELECT COUNT(*) FROM products;

-- Подсчет с условием
SELECT COUNT(*) FROM products WHERE stock > 10;

-- Агрегация: средняя цена
SELECT AVG(price) as average_price FROM products;

-- Агрегация: сумма всех товаров на складе
SELECT SUM(stock) as total_stock FROM products;

-- Группировка (например, товары по ценовым диапазонам)
SELECT
    CASE
        WHEN price < 5000 THEN 'Дешевые'
        WHEN price < 20000 THEN 'Средние'
        ELSE 'Дорогие'
    END as price_range,
    COUNT(*) as count
FROM products
GROUP BY price_range;

-- ========================================
-- 4. ОБНОВЛЕНИЕ ДАННЫХ (UPDATE)
-- ========================================

-- Обновить цену одного товара
UPDATE products
SET price = 80000.00
WHERE name = 'Ноутбук';

-- Обновить несколько полей
UPDATE products
SET price = 1200.00, stock = 60
WHERE name = 'Мышь';

-- Увеличить цену на 10% для всех товаров
UPDATE products
SET price = price * 1.1;

-- Уменьшить остаток (симуляция продажи)
UPDATE products
SET stock = stock - 1
WHERE name = 'Клавиатура'
RETURNING id, name, stock;

-- ========================================
-- 5. ПРОВЕРКА ПОСЛЕ UPDATE (SELECT)
-- ========================================

-- Смотрим, что изменилось
SELECT name, price, stock FROM products ORDER BY price DESC;

-- ========================================
-- 6. УДАЛЕНИЕ ДАННЫХ (DELETE)
-- ========================================

-- Удалить товар с нулевым остатком
UPDATE products SET stock = 0 WHERE name = 'Наушники';
DELETE FROM products WHERE stock = 0;

-- Удалить товары дороже определенной цены
DELETE FROM products WHERE price > 90000;

-- ========================================
-- 7. ПРОВЕРКА ПОСЛЕ DELETE (SELECT)
-- ========================================

-- Смотрим, что осталось
SELECT * FROM products;

-- ========================================
-- 8. СВЯЗИ МЕЖДУ ТАБЛИЦАМИ (JOIN)
-- ========================================

-- Выберем задачи пользователей (таблицы из миграций)
SELECT
    u.email,
    t.title,
    t.completed
FROM users u
JOIN todos t ON u.id = t.user_id
ORDER BY u.email, t.created_at;

-- Подсчет задач по пользователям
SELECT
    u.email,
    COUNT(t.id) as total_todos,
    SUM(CASE WHEN t.completed THEN 1 ELSE 0 END) as completed_todos
FROM users u
LEFT JOIN todos t ON u.id = t.user_id
GROUP BY u.email;

-- ========================================
-- 9. ТРАНЗАКЦИИ
-- ========================================

-- Начало транзакции
BEGIN;

-- Создаем нового пользователя
INSERT INTO users (email, password_hash)
VALUES ('charlie@example.com', '$2a$10$hash3');

-- Создаем для него задачу (получаем ID из предыдущего INSERT)
INSERT INTO todos (user_id, title, completed)
VALUES (currval('users_id_seq'), 'Изучить транзакции SQL', false);

-- Коммит (сохранение изменений)
COMMIT;

-- Проверка
SELECT u.email, t.title
FROM users u
JOIN todos t ON u.id = t.user_id
WHERE u.email = 'charlie@example.com';

-- ========================================
-- 10. ОТКАТ ТРАНЗАКЦИИ (ROLLBACK)
-- ========================================

BEGIN;

-- Удаляем все продукты
DELETE FROM products;

-- Проверяем (внутри транзакции)
SELECT COUNT(*) FROM products;  -- будет 0

-- Отменяем изменения!
ROLLBACK;

-- Проверяем снова (товары вернулись)
SELECT COUNT(*) FROM products;

-- ========================================
-- 11. ДОПОЛНИТЕЛЬНЫЕ ПРИМЕРЫ
-- ========================================

-- Поиск по подстроке (LIKE)
SELECT * FROM products WHERE name LIKE '%ноут%';

-- Поиск с несколькими условиями
SELECT * FROM products
WHERE price > 2000 AND stock > 10;

-- Обновление с условием
UPDATE products
SET stock = stock + 5
WHERE stock < 20;

-- Удаление всех записей (осторожно!)
-- DELETE FROM products;  -- раскомментируйте, если хотите удалить все

-- ========================================
-- 12. ПОЛЕЗНЫЕ КОМАНДЫ psql
-- ========================================

-- \dt                  - список всех таблиц
-- \d table_name        - описание таблицы
-- \du                  - список пользователей
-- \l                   - список баз данных
-- \q                   - выход из psql
-- \?                   - справка по командам
