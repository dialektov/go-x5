# Swagger

Сгенерированные файлы `docs.go`, `swagger.json` и `swagger.yaml` описывают обязательные эндпоинты.

Пересобрать документацию:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -o docs
```
