.PHONY: dev dev-backend dev-frontend

dev: buf-gen
	$(MAKE) -j 2 dev-backend dev-frontend

dev-backend: buf-gen-go
	air

dev-frontend: buf-gen-ts
	npm run dev

.PHONY: buf-gen buf-gen-go buf-gen-ts buf-clean-go buf-clean-ts install-tools install-go-tools install-ts-tools

buf-gen: install-tools buf-gen-go buf-gen-ts

buf-gen-go: buf-clean-go install-go-tools
	buf generate --template proto/buf.gen.go.yaml

buf-gen-ts: buf-clean-ts install-ts-tools
	buf generate --template proto/buf.gen.ts.yaml

buf-clean-go:
	rm -rf gen/go

buf-clean-ts:
	rm -rf gen/ts

install-tools: install-go-tools install-ts-tools

install-go-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

install-ts-tools:
	npm install -g @bufbuild/protoc-gen-es

.PHONY: run-server-container build-server-img 

SERVER_IMG_NAME = pb-server-img
SERVER_CONTAINER_NAME = pb-server-container

run-server-container: build-server-img
	docker run \
		--env PB_SERVER_PORT=8000 \
		--env PB_SERVER_HOST=0.0.0.0 \
		--interactive \
		--tty \
		--rm \
		--publish 8000:8000 \
		--name $(SERVER_CONTAINER_NAME)
		$(SERVER_IMG_NAME)

build-server-img:
	docker build \
		--file etc/docker/Dockerfile \
		--tag $(SERVER_IMG_NAME) .

.PHONY: up up-dev

up-dev:
	docker compose \
		--file etc/docker/docker-compose.dev.yaml \
		up \
		--build

up:
	docker compose \
		--file etc/docker/docker-compose.dev.yaml \
		--file etc/docker/docker-compose.prod.yaml \
		up \
		--build \
		--detach
