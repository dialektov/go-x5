# Сервис заказов

Backend-сервис заказов на Go. Данные хранятся в памяти процесса.

## Запуск

Нужен Go 1.22 или новее.

```bash
go mod tidy
go run ./cmd/api
```

Адрес по умолчанию — `:8080`. Его можно заменить переменной `HTTP_ADDR`.

Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Эндпоинты

| Метод | Путь | Успех |
|---|---|---:|
| GET | `/health` | 200 |
| POST | `/api/v1/orders` | 201 |
| GET | `/api/v1/orders` | 200 |
| GET | `/api/v1/orders/{id}` | 200 |
| PATCH | `/api/v1/orders/{id}/status` | 200 |
| GET | `/api/v1/orders/{id}/history` | 200 |

Создание заказа:

```json
{
  "items": [
    {
      "name": "Keyboard",
      "quantity": 1,
      "price": 500000
    }
  ]
}
```

`price` — целое число в копейках. Итоговая стоимость считается как сумма `price × quantity`. Новый заказ получает статус `new`.

Смена статуса:

```json
{
  "status": "processing"
}
```

Допустимые переходы: `new → processing`, `new → cancelled`, `processing → completed`, `processing → cancelled`. Статусы `completed` и `cancelled` конечные.

Ошибка возвращается в одном формате:

```json
{
  "error": "status transition is not allowed",
  "code": "INVALID_STATUS_TRANSITION"
}
```

Коды: `400` для некорректного JSON и данных, `404` если заказ не найден, `409` для недопустимого перехода, `500` для внутренней ошибки.

## Структура

```text
cmd/api                 сборка зависимостей и HTTP-сервер
internal/handler        HTTP, JSON и middleware
internal/service        валидация, стоимость, статусы и история
internal/order          модель, ошибки и интерфейс репозитория
internal/repository/memory
docs                    Swagger
```

Запрос проходит по цепочке handler → service → repository. Репозиторий передаётся в сервис через конструктор.

Каждый HTTP-запрос пишется в `slog` с полями `method`, `path`, `status`, `duration` и `request_id`. Panic перехватывается и превращается в JSON-ответ `500`.

## Проверка

```bash
go test ./...
go test ./internal/service -cover
go vet ./...
go build ./...
```

Swagger заново собирается так:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -o docs
```
