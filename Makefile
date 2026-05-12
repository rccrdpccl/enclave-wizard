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
