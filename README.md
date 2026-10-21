# E-Shop Backend

REST API для интернет-магазина на Go с использованием Gin, GORM и PostgreSQL.

## Стек технологий

- **Go 1.27.1** — основной язык
- **Gin** — HTTP фреймворк
- **GORM** — ORM для PostgreSQL
- **PostgreSQL 16** — база данных
- **golang-migrate** — миграции БД
- **shopspring/decimal** — точная работа с деньгами
- **go-playground/validator** — валидация входных данных

## Быстрый старт

### Локальная разработка

1. Клонируйте репозиторий и перейдите в директорию:
   ```bash
   cd backend
   ```

2. Скопируйте файл окружения:
   ```bash
   cp .env.example .env
   ```

3. Запустите сервисы:
   ```bash
   make docker-up
   ```

Приложение будет доступно на `http://localhost:8080`.

### Без Docker

Требуется установленный Go 1.27+ и PostgreSQL:

```bash
make run
```

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `DB_HOST` | Хост базы данных | `localhost` |
| `DB_PORT` | Порт базы данных | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `postgres` |
| `DB_NAME` | Имя базы данных | `shop` |
| `SERVER_PORT` | Порт HTTP сервера | `8080` |

## API Endpoints

### Products (Товары)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/api/products` | Создать товар |
| `GET` | `/api/products` | Получить список товаров (пагинация) |
| `GET` | `/api/products/:id` | Получить товар по ID |
| `PUT` | `/api/products/:id` | Обновить товар |
| `DELETE` | `/api/products/:id` | Удалить товар (soft delete) |

**Query параметры для списка:**
- `page` — номер страницы (по умолчанию 1)
- `page_size` — размер страницы (по умолчанию 20, макс 100)

### Orders (Заказы)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/api/orders` | Создать заказ |
| `GET` | `/api/orders` | Получить список заказов (пагинация) |
| `GET` | `/api/orders/:id` | Получить заказ по ID |
| `PUT` | `/api/orders/:id` | Обновить заказ |
| `DELETE` | `/api/orders/:id` | Удалить заказ (soft delete) |

**Query параметры для списка:**
- `page` — номер страницы (по умолчанию 1)
- `page_size` — размер страницы (по умолчанию 20, макс 100)

## Структура проекта

```
backend/
├── cmd/api/              # Точка входа приложения
├── internal/
│   ├── config/           # Конфигурация
│   ├── database/         # Подключение к БД и миграции
│   │   └── migrations/   # SQL файлы миграций
│   ├── domain/           # Доменные модели и ошибки
│   ├── repository/       # Слой доступа к данным
│   ├── service/          # Бизнес-логика
│   └── transport/http/   # HTTP handlers, middleware, router
```

## Миграции

Миграции применяются **автоматически** при запуске приложения через `database.RunMigrations()`.

Файлы миграций находятся в `internal/database/migrations/`:
- `000001_init_schema.up.sql` — создание таблиц
- `000001_init_schema.down.sql` — откат миграции

## Тестирование

```bash
# Запустить все тесты с race detector
make test

# Проверить код линтером
make lint
```

Покрытие сервисного слоя: ~90%

## Разработка

### Доступные команды Make

```bash
make help          # Показать все доступные команды
make run           # Запустить приложение локально
make build         # Собрать бинарник
make test          # Запустить тесты
make lint          # Проверить код
make docker-up     # Запустить через Docker
make docker-down   # Остановить Docker контейнеры
make clean         # Удалить артефакты сборки
```

## Известные ограничения

1. **Soft-delete race condition**: Параллельное создание заказа и удаление товара может привести к заказу с удалённым товаром (FK указывает на строку с `deleted_at IS NOT NULL`). Требуется триггер БД или дополнительная блокировка.

2. **Пагинация без total**: Список товаров и заказов не возвращает общее количество записей. Требуется `COUNT(*)` запрос.

3. **Нет аутентификации**: API открыт для всех запросов. Требуется JWT/OAuth для production.

4. **Нет rate limiting**: Отсутствует защита от DDoS. Требуется middleware.

5. **Метрики не собираются**: Нет Prometheus метрик. Требуется отдельная ветка для observability.

## Лицензия

MIT

---

**Проверка:**

```bash
# Проверить применение FOR UPDATE fix
grep -n "clause.Locking" internal/repository/product_repository.go
# Должно вывести: db.Clauses(clause.Locking{Strength: "UPDATE"})

# Запустить Docker build (multi-arch)
docker build -t eshop-api .

# Запустить через docker-compose
docker compose up --build -d

# Проверить работу
curl http://localhost:8080/api/products

# Остановить
docker compose down