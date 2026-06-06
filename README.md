# Bank API System

Банковская система с поддержкой счетов, карт, кредитов и аналитики.

## Как пользоваться сервисом

### 1. Запуск сервера

```bash
# Запустить PostgreSQL (Docker)
docker run -d --name bank-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=13d3791semenek \
  -e POSTGRES_DB=bankdb \
  -p 5432:5432 \
  postgres:15

# Запустить приложение
go run cmd/main.go
```

Сервер запустится на `http://localhost:8080`

### 2. Рабочий процесс

1. **Регистрация** → создаете пользователя
2. **Логин** → получаете JWT-токен
3. **Создаете счет** → в валюте RUB
4. **Пополняете счет** → вносите деньги
5. **Выпускаете карту** → получаете пластиковую карту
6. **Оформляете кредит** → деньги поступают на счет
7. **Смотрите аналитику** → доходы/расходы, прогнозы

## Команды API

### Публичные (без токена)

| Команда | Запрос | Описание |
|---------|--------|----------|
| Регистрация | `POST /register` | Создать пользователя |
| Логин | `POST /login` | Получить JWT-токен |

### Защищенные (с токеном)

**Счета**
| Команда | Запрос | Описание |
|---------|--------|----------|
| Создать счет | `POST /accounts` | Открыть новый счет |
| Мои счета | `GET /accounts` | Список всех счетов |
| Информация | `GET /accounts/{id}` | Данные счета |
| Пополнить | `POST /accounts/{id}/deposit` | Внести деньги |
| Снять | `POST /accounts/{id}/withdraw` | Вывести деньги |
| Перевести | `POST /transfer` | Перевод между счетами |

**Карты**
| Команда | Запрос | Описание |
|---------|--------|----------|
| Выпустить карту | `POST /cards` | Выпустить новую карту |
| Карты счета | `GET /accounts/{id}/cards` | Список карт счета |

**Кредиты**
| Команда | Запрос | Описание |
|---------|--------|----------|
| Оформить кредит | `POST /credits` | Получить кредит |
| График платежей | `GET /credits/{id}/schedule` | План выплат |
| Мои кредиты | `GET /credits` | Список кредитов |

**Аналитика**
| Команда | Запрос | Описание |
|---------|--------|----------|
| Финансы | `GET /analytics` | Доходы/расходы/долги |
| Прогноз | `GET /accounts/{id}/forecast?days=30` | Будущий баланс |

## Как тестировать

### Быстрый тест (PowerShell)

```powershell
# 1. Регистрация
Invoke-RestMethod -Uri "http://localhost:8080/register" -Method POST `
  -ContentType "application/json" `
  -Body '{"username":"test","email":"test@test.com","password":"test12345"}'

# 2. Логин (получить токен)
$login = Invoke-RestMethod -Uri "http://localhost:8080/login" -Method POST `
  -ContentType "application/json" `
  -Body '{"email":"test@test.com","password":"test12345"}'
$token = $login.token
$headers = @{ Authorization = "Bearer $token" }

# 3. Создать счет
$account = Invoke-RestMethod -Uri "http://localhost:8080/accounts" -Method POST `
  -Headers $headers -ContentType "application/json" -Body '{"currency":"RUB"}'

# 4. Пополнить на 50000
Invoke-RestMethod -Uri "http://localhost:8080/accounts/$($account.id)/deposit" `
  -Method POST -Headers $headers -ContentType "application/json" -Body '{"amount":50000}'

# 5. Выпустить карту
Invoke-RestMethod -Uri "http://localhost:8080/cards" -Method POST `
  -Headers $headers -ContentType "application/json" -Body "{`"account_id`": $($account.id)}"

# 6. Оформить кредит 100000 на 12 месяцев
Invoke-RestMethod -Uri "http://localhost:8080/credits" -Method POST `
  -Headers $headers -ContentType "application/json" `
  -Body "{`"account_id`": $($account.id), `"principal`": 100000, `"term_months`": 12}"

# 7. Посмотреть аналитику
Invoke-RestMethod -Uri "http://localhost:8080/analytics" -Method GET -Headers $headers

Write-Host "✅ Все тесты пройдены!" -ForegroundColor Green
```

### Тест перевода (между двумя пользователями)

```powershell
# Создать второго пользователя
Invoke-RestMethod -Uri "http://localhost:8080/register" -Method POST `
  -ContentType "application/json" `
  -Body '{"username":"user2","email":"user2@test.com","password":"test12345"}'

# Логин второго
$login2 = Invoke-RestMethod -Uri "http://localhost:8080/login" -Method POST `
  -ContentType "application/json" -Body '{"email":"user2@test.com","password":"test12345"}'
$headers2 = @{ Authorization = "Bearer $($login2.token)" }

# Создать счет второму
$account2 = Invoke-RestMethod -Uri "http://localhost:8080/accounts" -Method POST `
  -Headers $headers2 -ContentType "application/json" -Body '{"currency":"RUB"}'

# Перевод 10000 от первого ко второму (используем токен первого)
Invoke-RestMethod -Uri "http://localhost:8080/transfer" -Method POST `
  -Headers $headers -ContentType "application/json" `
  -Body "{`"from_account_id`": $($account.id), `"to_account_id`": $($account2.id), `"amount`": 10000}"

Write-Host "✅ Перевод выполнен!" -ForegroundColor Green
```

### Негативные тесты (проверка ошибок)

```powershell
# Неверный пароль
Invoke-RestMethod -Uri "http://localhost:8080/login" -Method POST `
  -ContentType "application/json" -Body '{"email":"test@test.com","password":"wrong"}'
# Ожидаем: 401 Unauthorized

# Снятие больше чем есть
Invoke-RestMethod -Uri "http://localhost:8080/accounts/$($account.id)/withdraw" `
  -Method POST -Headers $headers -ContentType "application/json" -Body '{"amount":999999}'
# Ожидаем: 400 Bad Request (insufficient funds)

# Без токена
Invoke-RestMethod -Uri "http://localhost:8080/accounts" -Method GET
# Ожидаем: 401 Unauthorized
```

### Проверка БД

```bash
# Подключиться к PostgreSQL
docker exec -it bank-postgres psql -U postgres -d bankdb

# Посмотреть данные
\dt
SELECT * FROM users;
SELECT * FROM accounts;
SELECT * FROM cards;
SELECT * FROM transactions;
SELECT * FROM credits;
```

## Ожидаемые ответы

**Успех (201 Created)**
```json
{
  "id": 1,
  "username": "test",
  "email": "test@test.com",
  "created_at": "2026-06-06T18:52:12Z"
}
```

**Ошибка (400 Bad Request)**
```json
{
  "error": "password must be at least 8 characters"
}
```

## Готовые сценарии

| Что тестировать | Как |
|----------------|-----|
| Полный цикл | Скопировать и выполнить "Быстрый тест" |
| Переводы | Запустить "Тест перевода" |
| Ошибки | Выполнить "Негативные тесты" |
| Кредиты | Оформить кредит, посмотреть график |
| Прогноз | Вызвать `/forecast?days=30` |

## Примечания

- После каждого запроса проверяйте HTTP-статус (201/200 = успех)
- Токен живет 24 часа
- Все защищенные запросы требуют заголовок: `Authorization: Bearer <token>`
- Поддерживается только RUB
- Данные карт шифруются (PGP + HMAC)
- Пароли и CVV хранятся в хешированном виде (bcrypt)
