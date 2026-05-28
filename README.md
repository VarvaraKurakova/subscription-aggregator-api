# Subscription Aggregator API

REST-сервис на Go для управления онлайн-подписками пользователей и расчёта суммарной стоимости подписок за выбранный период.

Проект реализован как backend-сервис с PostgreSQL, миграциями, Docker Compose, Swagger-документацией, логированием и слоистой архитектурой.

## Возможности

- CRUD-операции над пользовательскими подписками.
- Хранение подписок с полями:
  - название сервиса;
  - стоимость месячной подписки в рублях;
  - ID пользователя в формате UUID;
  - дата начала подписки в формате `MM-YYYY`;
  - опциональная дата окончания подписки.
- Расчёт суммарной стоимости подписок за выбранный период.
- Фильтрация расчёта по `user_id` и `service_name`.
- Валидация входных данных.
- PostgreSQL-хранилище с миграциями.
- Swagger/OpenAPI-документация.
- Запуск приложения через Docker Compose.

## Стек

- Go
- chi
- PostgreSQL
- pgx / pgxpool
- Docker / Docker Compose
- SQL migrations
- Swagger / OpenAPI
- slog

## Архитектура проекта

Проект разделён на слои:

```text
cmd/app             — точка входа в приложение
internal/config     — загрузка конфигурации
internal/db         — подключение к PostgreSQL
internal/domain     — доменные модели и работа с датами
internal/repository — работа с базой данных
internal/service    — бизнес-логика и валидация
internal/handler    — HTTP handlers
internal/transport  — router и middleware
migrations          — SQL-миграции
docs                — Swagger-документация
```

Основной поток обработки запроса:

```text
HTTP request
    ↓
handler
    ↓
service
    ↓
repository
    ↓
PostgreSQL
```

## Структура подписки

Пример создания подписки:

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2f41-4721-ae6d-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

Поле `end_date` является опциональным. Если оно отсутствует, подписка считается активной.

## Логика расчёта суммы

Сервис рассчитывает стоимость подписок за выбранный период с учётом пересечения периода подписки и периода запроса.

Например:

```text
Подписка:
price = 400
start_date = 07-2025
end_date = 12-2025

Период запроса:
from = 09-2025
to = 11-2025
```

В расчёт попадут сентябрь, октябрь и ноябрь:

```text
400 * 3 = 1200
```

Если `end_date` не указана, подписка считается активной до конца запрошенного периода.

## Запуск проекта

### 1. Клонировать репозиторий

```bash
git clone https://github.com/VarvaraKurakova/subscription-aggregator-api.git
cd subscription-aggregator-api
```

### 2. Запустить сервис через Docker Compose

```bash
docker compose up --build
```

Будут подняты:

- Go-приложение;
- PostgreSQL;
- контейнер для применения миграций.

Приложение будет доступно по адресу:

```text
http://localhost:8080
```

## Переменные окружения

Пример конфигурации находится в `.env.example`:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=subscriptions
DB_SSLMODE=disable

LOG_LEVEL=info
```

При запуске через Docker Compose переменные окружения для приложения задаются в `docker-compose.yml`.

## Swagger

Swagger UI доступен после запуска приложения:

```text
http://localhost:8080/swagger/index.html
```

JSON-документация:

```text
http://localhost:8080/swagger/doc.json
```

## API endpoints

### Health-check

```http
GET /health
```

### Создать подписку

```http
POST /subscriptions/
```

Пример:

```bash
curl -X POST http://localhost:8080/subscriptions/ \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2f41-4721-ae6d-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Получить список подписок

```http
GET /subscriptions/
```

С фильтрами:

```bash
curl "http://localhost:8080/subscriptions/?user_id=60601fee-2f41-4721-ae6d-7636e79a0cba&service_name=Yandex%20Plus"
```

Поддерживаемые query-параметры:

```text
user_id
service_name
limit
offset
```

### Получить подписку по ID

```http
GET /subscriptions/{id}
```

Пример:

```bash
curl http://localhost:8080/subscriptions/1
```

### Обновить подписку

```http
PUT /subscriptions/{id}
```

Пример:

```bash
curl -X PUT http://localhost:8080/subscriptions/1 \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 500,
    "user_id": "60601fee-2f41-4721-ae6d-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

### Удалить подписку

```http
DELETE /subscriptions/{id}
```

Пример:

```bash
curl -i -X DELETE http://localhost:8080/subscriptions/1
```

При успешном удалении сервис возвращает:

```text
204 No Content
```

### Рассчитать суммарную стоимость подписок

```http
GET /subscriptions/total?from=07-2025&to=09-2025
```

Пример:

```bash
curl "http://localhost:8080/subscriptions/total?from=07-2025&to=09-2025"
```

Пример ответа:

```json
{
  "total": 1200,
  "currency": "RUB",
  "period_from": "07-2025",
  "period_to": "09-2025"
}
```

С фильтрами:

```bash
curl "http://localhost:8080/subscriptions/total?from=07-2025&to=09-2025&user_id=60601fee-2f41-4721-ae6d-7636e79a0cba&service_name=Yandex%20Plus"
```

Обязательные параметры:

```text
from
to
```

Опциональные параметры:

```text
user_id
service_name
```

## Миграции

Миграции находятся в папке:

```text
migrations/
```

При запуске через Docker Compose миграции применяются автоматически отдельным контейнером `migrate`.

Основная таблица:

```sql
subscriptions
```

Поля:

```text
id
service_name
price
user_id
start_date
end_date
created_at
updated_at
```

## Проверка проекта

Локальная проверка Go-кода:

```bash
go fmt ./...
go mod tidy
go test ./...
```

Проверка запуска через Docker Compose:

```bash
docker compose down
docker compose up --build
```

## Особенности реализации

- Даты в API принимаются в формате `MM-YYYY`.
- В PostgreSQL даты хранятся как `DATE`, где день всегда равен первому числу месяца.
- `end_date` включается в расчёт.
- Если `end_date = null`, подписка считается активной до конца выбранного периода.
- Расчёт суммы выполняется в service layer.
- Работа с PostgreSQL вынесена в repository layer.
- HTTP-слой отделён от бизнес-логики.
