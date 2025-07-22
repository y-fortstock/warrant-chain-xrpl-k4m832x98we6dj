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

## Configuration

The service uses [Viper](https://github.com/spf13/viper) for configuration management. You can configure the log level via environment variable, YAML file, or CLI flag.

### Environment Variable

Set the log level using the environment variable:

```
LOG_LEVEL=debug ./chain-xrpl
```

### YAML Config

Create a `config.yaml` file in the project root (or specify with `--config`). Example:

```yaml
log:
  level: info # or debug, warn, error
```

### CLI Flag

You can specify a custom config file with:

```
./chain-xrpl --config /path/to/config.yaml
```

The precedence is: CLI flag > environment variable > YAML file > default (info).

---
