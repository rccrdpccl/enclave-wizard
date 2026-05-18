BINARY := enclave-wizard
GO := go
CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null || echo docker)
UI_IMAGE := enclave-wizard-ui:dev

.PHONY: build build-linux build-ui run test lint clean tidy deploy teardown generate

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

generate:
	$(GO) generate ./...

rpm: build-linux build-ui
	hack/rpm/build-rpm.sh

deploy: build-linux build-ui
	@test -n "$(TARGET)" || (echo "Usage: make deploy TARGET=root@host" && exit 1)
	hack/deploy-wizard $(TARGET)

teardown:
	@test -n "$(TARGET)" || (echo "Usage: make teardown TARGET=root@host" && exit 1)
	hack/teardown-wizard $(TARGET)

e2e: rpm
	@test -n "$(TARGET)" || (echo "Usage: make e2e TARGET=root@host" && exit 1)
	hack/e2e/run-e2e.sh --host $(TARGET)

e2e-rerun:
	@test -n "$(TARGET)" || (echo "Usage: make e2e-rerun TARGET=root@host" && exit 1)
	hack/e2e/run-e2e.sh --host $(TARGET) --skip-deploy --skip-teardown

e2e-browser:
	@test -n "$(WIZARD_URL)" || (echo "Usage: make e2e-browser WIZARD_URL=https://localhost:3443" && exit 1)
	cd ui/apps/wizard && WIZARD_URL=$(WIZARD_URL) yarn e2e

e2e-full: rpm
	@test -n "$(TARGET)" || (echo "Usage: make e2e-full TARGET=root@host" && exit 1)
	hack/e2e/run-e2e.sh --host $(TARGET)
	$(MAKE) e2e-browser WIZARD_URL=https://$(shell echo $(TARGET) | cut -d@ -f2):3443
