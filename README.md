# GETTING STARTED

Для подключения к бд и запуска миграций нужно запустить
```
docker compose up -d
```

в корне проекта набрать команды:
```
go run cmd/server/main.go
go run cmd/client/main.go
```

# Tests

Дя проверки покрытия кода тестами в корне проекта набрать команду
```
make test-cover
```
