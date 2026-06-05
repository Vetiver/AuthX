# AuthX — микросервис аутентификации и авторизации с Kafka-аудитом

## Описание

AuthX — Go-микросервис для регистрации, аутентификации и ролевой авторизации.
Все события безопасности публикуются в Kafka через transactional outbox.
Отдельный consumer сохраняет аудит в PostgreSQL. Невалидные события попадают в DLQ.

## Запуск

```bash
# Поднять все сервисы
docker compose up -d

# Создать топики (один раз)
docker exec authx-kafka-1 kafka-topics --bootstrap-server localhost:9092 --create --topic authx.auth.events --partitions 3 --replication-factor 1
docker exec authx-kafka-1 kafka-topics --bootstrap-server localhost:9092 --create --topic authx.auth.events.dlq --partitions 3 --replication-factor 1

# Проверить топики
docker exec authx-kafka-1 kafka-topics --bootstrap-server localhost:9092 --list

## Компоненты

| Сервис | Порт | Описание |
|---|---|---|
| authx | 8888 | Основной API |
| postgresql-authx | 5449 | PostgreSQL |
| redis-authx | 6379 | Redis |
| kafka | 9092 | Kafka broker |
| audit-consumer | - | Consumer аудита |

## API

### Регистрация

Создаёт пользователя с ролью `user`. Событие `user_registered` пишется в outbox и отправляется в Kafka.

```
POST /auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

**Ответ (201):**
```json
{
    "code": "SUCCESS",
    "message": "user registered successfully"
}
```

**Ошибка (409):**
```json
{
    "code": "REGISTRATION_FAILED",
    "message": "email user@example.com already registered"
}
```

**Ошибка (400):**
```json
{
    "code": "INVALID_REQUEST",
    "message": "Key: 'RegisterReq.Email' Error:Field validation for 'Email' failed on the 'required' tag"
}
```

---

### Вход

Проверяет email/пароль, генерирует JWT access token. Старый токен пользователя инвалидируется через Redis blacklist. События `auth_login_succeeded` или `auth_login_failed` пишутся в outbox.

```
POST /auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

**Ответ (200):**
```json
{
    "code": "SUCCESS",
    "data": {
        "id": 1,
        "email": "user@example.com",
        "role": "user",
        "created_at": "2026-06-05T10:00:00Z",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
}
```

**Ошибка (401):**
```json
{
    "code": "LOGIN_FAILED",
    "message": "invalid email or password"
}
```

---

### Список ролей

Требует валидный JWT токен в заголовке Authorization.

```
GET /auth/roles
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Ответ (200):**
```json
{
    "code": "SUCCESS",
    "data": ["admin", "user"]
}
```

**Ошибка (401):**
```json
{
    "code": "TOKEN_INVALID",
    "message": "token invalid or expired"
}
```

**Ошибка (401) — токен отозван:**
```json
{
    "code": "TOKEN_REVOKED",
    "message": "token has been revoked"
}
```

---

### Ping

```
GET /ping
```

**Ответ (200):**
```
OK
```

## Kafka

### Топики

| Топик | Назначение | Retention |
|---|---|---|
| authx.auth.events | События безопасности | 7 дней |
| authx.auth.events.dlq | Dead Letter Queue | 14 дней |

### События

| Тип | Когда |
|---|---|
| user_registered | После регистрации |
| auth_login_succeeded | Успешный вход |
| auth_login_failed | Неуспешный вход |
| auth_token_validated | Токен валиден |
| auth_token_rejected | Токен невалиден |

### Формат сообщения

```json
{
    "event_id": "a1b2c3d4-...",
    "event_type": "auth_login_succeeded",
    "occurred_at": "2026-06-05T10:00:00Z",
    "user_id": 1,
    "email": "user@example.com",
    "role": "user",
    "ip": "192.168.1.1",
    "user_agent": "curl/8.0.1",
    "metadata": {}
}
```

### Просмотр сообщений

```bash
# Основной топик
docker exec authx-kafka-1 kafka-console-consumer --bootstrap-server localhost:9092 --topic authx.auth.events --from-beginning

# DLQ
docker exec authx-kafka-1 kafka-console-consumer --bootstrap-server localhost:9092 --topic authx.auth.events.dlq --from-beginning
```

## Аудит

Consumer сохраняет события в таблицу `audit_events`.

### Просмотр аудита

```sql
SELECT * FROM audit_events ORDER BY created_at DESC;
```
