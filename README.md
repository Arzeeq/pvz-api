# PVZ api


## Структура проекта

### Основные директории

- `api/` - OpenAPI спецификация, .proto файл.

- `cmd/` - основные приложения
    - `pvz-api/` - запуск api сервера

- `configs/` - различные конфигурации приложения

- `internal/` - внутренняя логика приложения
    - `config/` - работа с конфигурацией
    - `dto/` - Data Transfer Objects
    - `handler/` - слой хэндлеров для эндпоинтов
        - `grpc/` - хэндлеры gRPC
        - `http/` - хэндлеры HTTP
    - `logger/` - настройка логирования в проекте
    - `middleware/` - middleware аутентификации
    - `server/` - модуль сервера
    - `service/` - слой сервисов
    - `storage/` - слой базы данных
        - `pg/` - storages для PostgreSQL
            - `migrations/` - миграции базы данных
    - `test/` - интеграционный тест

- `pkg/` - библиотеки, которые можно использовать в сторонних проектах

### Docker файлы

- `Dockerfile` - конфигурация образа Docker контейнера для pvz-api
- `docker-compose.yaml` - файл для запуска pvz-api


protoc --go-grpc_out=. --go_out=. --go-grpc_opt=module=github.com/Arzeeq/pvz-api --go_opt=module=github.com/Arzeeq/pvz-api ./api/pvz.proto