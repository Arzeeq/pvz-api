# PVZ api
[![Main workflow](https://github.com/Arzeeq/pvz-api/actions/workflows/main.yml/badge.svg)](https://github.com/Arzeeq/pvz-api/actions/workflows/main.yml)
## Описание

[Тестовое задание](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-spring-2025/Backend-trainee-assignment-spring-2025.md) для стажёра Backend-направления (весенняя волна 2025)

## Запуск

### Запуск сервиса api 
```bash
docker compose up
```
или
```bash
docker-compose up
```

после чего вам станут доступны следующие endpoint-ы:
- `POST`    <http://localhost:8080/dummyLogin>
- `POST`    <http://localhost:8080/register>
- `POST`    <http://localhost:8080/login>
- `POST`    <http://localhost:8080/pvz>
- `GET`     <http://localhost:8080/pvz>
- `POST`    <http://localhost:8080/pvz/{pvzId}/close_last_reception>
- `POST`    <http://localhost:8080/pvz/{pvzId}/delete_last_product>
- `POST`    <http://localhost:8080/receptions>
- `POST`    <http://localhost:8080/products>

более подробно про формат использования endpoint-ов можно прочитать в [swagger.yaml](api/swagger.yaml), или загрузить содержимое этого файла в [данный](https://editor.swagger.io/) ресурс.

### gRPC сервер
На порту `3000` будет запущен gRPC сервер c одним доступным методом
- `GetPVZList`

### Prometheus
Prometheus запускается на порту `9000`.   
Метрики prometheus можно получить по адресу <http://localhost:9000/metrics>.   
Доступны следующие метрики:
- `http_requests_total` - Общее количество HTTP запросов к серверу
- `http_request_duration_in_seconds` - Длительность запросов в секундах
- `pvz_created_total` - Количество созданных ПВЗ
- `receipts_created_total` - Количество открытых приемок
- `products_added_total` - Количество добавленных товаров

### Запуск тестов

Для запуска unit тестов выполните команду из корня проекта
```bash
go test ./internal/service -cover
```
Для запуска интеграционного теста выполните команду из корня проекта
```bash
go test -v ./internal/test
```

Таким образом:
- `Storage` слой покрыт **интеграционным** тестом
- `Service` слой покрыт **интеграционным** и **unit** тестами
- `Handler` слой покрыт **интеграционным** тестом

## Структура проекта

- `api/` - OpenAPI спецификация, .proto файл.
- `cmd/` - основные приложения
    - `pvz-api/` - запуск api сервера
        - `app/` - приложение
        - `main.go` - точка входа в приложение 
- `configs/` - различные конфигурации приложения
- `internal/` - внутренняя логика приложения
    - `config/` - работа с конфигурацией
    - `dto/` - DTO сгенерированные из спецификации swagger.yaml
    - `grpc/` - gRPC сгенерированный из pvz.proto
    - `handler/` - слой хэндлеров для эндпоинтов
        - `grpc/` - хэндлеры gRPC
        - `http/` - хэндлеры HTTP
    - `logger/` - настройка логирования в проекте
    - `metrics/` - регистрация метрик для Prometheus
    - `middleware/` - аутентификация и подсчет метрик
    - `server/` - модуль HTTP и gRPC сервера
    - `service/` - слой сервисов
    - `storage/` - слой базы данных
        - `pg/` - storages для PostgreSQL
            - `migrations/` - миграции базы данных PostgreSQL
    - `test/` - интеграционный тест
- `pkg/` - библиотеки, которые можно использовать в сторонних проектах
    - `auth/` - библиотека генерации jwt и шифрования пароля


## Docker файлы

- `Dockerfile` - конфигурация образа Docker контейнера для pvz-api
- `docker-compose.yaml` - файл для запуска pvz-api

## Кодогенерация

В проекте используется кодогенерация DTO и gRPC из [swagger.yaml](api/swagger.yaml) и [pvz.proto](api/pvz.proto) файлов соответственно.


Для генерации DTO используется пакет [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).    
В директории `internal/dto` лежит файл [generate.go](internal/dto/generate.go), который отвечает за генерацию, и [cfg.yaml](internal/dto/cfg.yaml) для настройки конфигурации.

## Github Actions
Настроен `workflow`, который при пуше в ветку автоматически запускает golangci-lint и тесты.
Для генерации gRPC кода можно воспользоваться утилитой `protoc`.
Из корня проекта нужно запустить

```bash
protoc --go-grpc_out=. --go_out=. --go-grpc_opt=module=github.com/Arzeeq/pvz-api --go_opt=module=github.com/Arzeeq/pvz-api ./api/pvz.proto
```

> [!NOTE]
> Если бы на выполнение задания было выделено больше времени, то я бы обязательно сделал следующие вещи:
> - Добавил больше интеграционных тестов, чтобы проверить различные сценарии работы приложения, в том числе случаи, когда должны срабатывать ограничения (например запрет на создание приемки, если в данном ПВЗ уже открыта другая приемка);
> - Нагрузочное тестирование с использованием k6, для того чтобы убедиться что сервер api удовлетворяет нефункциональным требованиям
