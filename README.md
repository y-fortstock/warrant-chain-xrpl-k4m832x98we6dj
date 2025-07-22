# chain-xrpl

## Quickstart

### 1. Установка зависимостей и vendor

```
go mod tidy
go mod vendor
```

### 2. Генерация wire

```
cd internal/di
wire
```

### 3. Запуск CLI

```
go run ./cmd/chain-xrpl
```

### 4. Старт gRPC сервера (пример)

Добавьте команду или вызовите из main:

```
server := di.InitializeServer()
server.Run(":50051")
```

### 5. Подключение proto через git submodule

Рекомендуется подключить proto-файлы как сабмодуль:

```
git submodule add <repo_with_proto> proto
```

---
