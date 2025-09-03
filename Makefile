.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: build
build: tidy
	go build -o=/tmp/bin/server ./cmd/server

.PHONY: build/wasm
build/wasm: tidy
	GOOS=js GOARCH=wasm go build -o /tmp/bin/game.wasm ./cmd/wasm

.PHONY: run
run: build
	/tmp/bin/server

.PHONY: run/local
run/local: build build/wasm
	cp web/*.html /tmp/bin/
	cp ${shell go env GOROOT}/misc/wasm/wasm_exec.js /tmp/bin/
	STATIC_DIR=/tmp/bin SERVER_PORT=9001 /tmp/bin/server

.PHONY: run/docker
run/docker:
	docker build -t 15-puzzle:dev .
	docker run -it -p 9001:9001 --rm --env SERVER_PORT=9001 --name 15-puzzle 15-puzzle:dev

.PHONY: run/ui
run/ui:
	go build -v -o=/tmp/bin/ui ./cmd/ui/
	/tmp/bin/ui
