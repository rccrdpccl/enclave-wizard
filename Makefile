BINARY := enclave-wizard
GO := go
CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null || echo docker)
UI_IMAGE := enclave-wizard-ui:dev

.PHONY: build build-linux build-ui run test lint clean tidy deploy teardown

build:
	$(GO) build -o $(BINARY) .

build-linux:
	$(CONTAINER_RUNTIME) run --rm -v $(PWD):/app:z -w /app golang:latest \
		sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY) ."

build-ui:
	$(CONTAINER_RUNTIME) build -q -f ui/Containerfile -t $(UI_IMAGE) ui/

run: build
	./$(BINARY) --port 8080 --enclave-dir ../enclave

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

clean:
	rm -f $(BINARY)

tidy:
	$(GO) mod tidy

deploy: build-linux build-ui
	@test -n "$(TARGET)" || (echo "Usage: make deploy TARGET=root@host" && exit 1)
	hack/deploy-wizard $(TARGET)

teardown:
	@test -n "$(TARGET)" || (echo "Usage: make teardown TARGET=root@host" && exit 1)
	hack/teardown-wizard $(TARGET)
