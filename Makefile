# Makefile для chain-xrpl

.PHONY: docker-make deps gen submodule-update regen build run test-api stop help image k3d

setup:
	docker network create athens-net || true
	docker run --rm -d \
		-e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
		-e ATHENS_STORAGE_TYPE=disk \
		--network athens-net \
		-v /Users/undead/SrcCode/data/gomod:/var/lib/athens \
		-p 3333:3000 \
		--name athens \
		gomods/athens:latest

docker-make:
	docker build -f Dockerfile.make -t chain-xrpl-make .

deps: docker-make
	docker run --rm --network athens-net -v "$(shell pwd)":/app -w /app -e GOPROXY=http://athens:3000 -e GOSUMDB=off chain-xrpl-make go mod tidy && go mod vendor
  # docker run --rm -v "$(shell pwd)":/app -w /app -e GOPROXY=https://goproxy.cn,direct -e GOSUMDB=sum.golang.google.cn chain-xrpl-make go mod tidy && go mod vendor
  # docker run --rm -v "$(shell pwd)":/app -w /app -e GOPROXY=direct -e GOSUMDB=off -e GODEBUG=netdns=go,http2client=0 chain-xrpl-make go mod tidy && go mod vendor

gen:
	docker run --rm -v "$(shell pwd)":/app -w /app chain-xrpl-make sh -c "cd internal/di && wire"
	docker run --rm -v "$(shell pwd)":/app -w /app chain-xrpl-make sh -c "cd ./proto && buf generate"

submodule-update:
	git submodule update --init --recursive
	git submodule foreach git pull origin master
	go mod tidy
	go mod vendor

regen: submodule-update deps gen

build:
	docker build -t chain-xrpl .

rebuild: regen build

run: build
	docker run -d --rm --name chain-xrpl -p 8099:8099 chain-xrpl

test-unit:
	go test ./... -v
	go test -bench=. -benchmem ./...

test-api:
	bash .debug/api-tests/test_grpc_api.sh

stop:
	docker stop chain-xrpl

rerun: stop run

image:
	docker build -t localhost:5010/warrant1/warrant/chain-xrpl:latest .
	docker push localhost:5010/warrant1/warrant/chain-xrpl:latest

k3d: image
	kubectl rollout restart deployment chain-xrpl -n warrant

help:
	@echo "\033[1;33mAvailable commands:\033[0m"
	@echo "  \033[1;33mdocker-make\033[0m       \033[0;37m- Build Docker image for make environment\033[0m"
	@echo "  \033[1;33mdeps\033[0m              \033[0;37m- Install Go dependencies and vendor them\033[0m"
	@echo "  \033[1;33mgen\033[0m               \033[0;37m- Run wire and buf generate\033[0m"
	@echo "  \033[1;33msubmodule-update\033[0m  \033[0;37m- Update git submodules\033[0m"
	@echo "  \033[1;33mregen\033[0m             \033[0;37m- Update submodules, deps, and generate code\033[0m"
	@echo "  \033[1;33mbuild\033[0m             \033[0;37m- Build Docker image for chain-xrpl\033[0m"
	@echo "  \033[1;33mrebuild\033[0m           \033[0;37m- Full rebuild (regen + build)\033[0m"
	@echo "  \033[1;33mrun\033[0m               \033[0;37m- Run chain-xrpl container on port 8099\033[0m"
	@echo "  \033[1;33mstop\033[0m              \033[0;37m- Stop chain-xrpl container\033[0m"
	@echo "  \033[1;33mtest-api\033[0m          \033[0;37m- Run grpcurl tests\033[0m"
	@echo "  \033[1;33mimage\033[0m             \033[0;37m- Build Docker image for chain-xrpl\033[0m"
	@echo "  \033[1;33mk3d\033[0m               \033[0;37m- Rollout restart deployment chain-xrpl in k3d\033[0m"
