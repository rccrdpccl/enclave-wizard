BINARY := enclave-wizard
GO := go

.PHONY: build run test lint clean tidy

build:
	$(GO) build -o $(BINARY) .

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

deploy:
	@test -n "$(TARGET)" || (echo "Usage: make deploy TARGET=root@host" && exit 1)
	hack/deploy-wizard $(TARGET)

teardown:
	@test -n "$(TARGET)" || (echo "Usage: make teardown TARGET=root@host" && exit 1)
	hack/teardown-wizard $(TARGET)
