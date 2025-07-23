# Makefile для chain-xrpl

.PHONY: all deps wire build help run proto submodule-update

help:
	@echo "Доступные команды:"
	@echo "  make deps   - обновить зависимости и vendor (go mod tidy && go mod vendor)"
	@echo "  make wire   - сгенерировать wire (cd internal/di && wire)"
	@echo "  make proto  - сгенерировать Go-код из protobuf через buf (cd ../protobuf && buf generate)"
	@echo "  make build  - собрать бинарник (go build -o bin/chain-xrpl ./cmd/chain-xrpl)"
	@echo "  make run    - запустить приложение (go run ./cmd/chain-xrpl)"
	@echo "  make all    - proto + deps + wire + build"
	@echo "  make submodule-update - обновить git submodule (git submodule update --init --recursive && git submodule foreach git pull origin main)"

all: proto deps wire build

# Обновление зависимостей и vendor

deps:
	go mod tidy
	go mod vendor

# Генерация wire (DI)

wire:
	cd internal/di && wire

# Генерация Go-кода из protobuf через buf

proto:
	cd ./proto && buf generate

# Сборка бинарника

build: wire proto
	go build -o bin/chain-xrpl ./cmd/chain-xrpl 

run: wire proto
	go run ./cmd/chain-xrpl 

submodule-update:
	git submodule update --init --recursive
	git submodule foreach git pull origin master 