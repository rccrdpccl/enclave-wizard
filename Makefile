BINARY := enclave-wizard
GO := go
CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null || echo docker)

.PHONY: build build-linux build-ui run test lint clean tidy deploy teardown generate enclave-mock clean-enclave-mock run-mock preview deploy-preview bm-emulation bm-emulation-config bm-emulation-cleanup

build-ui:
	$(CONTAINER_RUNTIME) run --rm -v $(PWD)/ui:/app:z -w /app node:22-alpine \
		sh -c "corepack enable && yarn install && \
		yarn workspace @enclave-wizard-ui/wizard run -T vite build"

build: build-ui
	$(GO) build -ldflags="-w -s" -tags "$(TAGS)" -o $(BINARY) .

build-linux: build-ui
	$(CONTAINER_RUNTIME) run --rm -v $(PWD):/app:z -w /app golang:latest \
		sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s' -tags '$(TAGS)' -o $(BINARY) ."

run: build
	./$(BINARY) --enclave-dir ../enclave --tls-cert hack/tls/server.crt --tls-key hack/tls/server.key

run-demo: build-ui
	$(GO) build -ldflags="-w -s" -tags dev -o $(BINARY) .
	./$(BINARY) --demo-deploy --enclave-dir ../enclave --tls-cert hack/tls/server.crt --tls-key hack/tls/server.key

preview: build-ui
	$(CONTAINER_RUNTIME) run --rm -v $(PWD):/app:z -w /app golang:latest \
		sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s' -tags dev -o $(BINARY) ."
	hack/run-preview.sh $(PORT)

test:
	$(GO) test -cover ./...

lint:
	$(GO) vet ./...

clean:
	rm -f $(BINARY)
	rm -rf ui/apps/wizard/dist

tidy:
	$(GO) mod tidy

generate:
	$(GO) generate ./...

rpm: build-linux
	hack/rpm/build-rpm.sh

deploy-preview:
	@test -n "$(TARGET)" || (echo "Usage: make deploy-preview TARGET=root@host [PORT=3443]" && exit 1)
	hack/deploy-preview.sh $(TARGET) $(PORT)

deploy: rpm
	@test -n '$(TARGET)' || (echo "Usage: make deploy TARGET=root@host [AUTH=none]" && exit 1)
	AUTH='$(AUTH)' hack/deploy-wizard '$(TARGET)'

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

bm-emulation:
	@test -n "$(TARGET)" || (echo "Usage: make bm-emulation TARGET=root@host" && exit 1)
	hack/infra/bm-emulation.sh --host $(TARGET)

bm-emulation-config:
	@test -n '$(TARGET)' || (echo "Usage: make bm-emulation-config TARGET=root@host PULL_SECRET=/path/to/pull-secret.json" && exit 1)
	@test -n '$(PULL_SECRET)' || (echo "Usage: make bm-emulation-config TARGET=root@host PULL_SECRET=/path/to/pull-secret.json" && exit 1)
	hack/infra/bm-emulation-config.sh --host '$(TARGET)' --pull-secret '$(PULL_SECRET)'

bm-emulation-cleanup:
	@test -n "$(TARGET)" || (echo "Usage: make bm-emulation-cleanup TARGET=root@host" && exit 1)
	hack/infra/bm-emulation-cleanup.sh --host $(TARGET)

ENCLAVE_MOCK_BRANCH ?= main
ENCLAVE_MOCK_REPO ?= git@github.com:rccrdpccl/enclave.git

enclave-mock:
	python3 hack/generate-enclave-mock.py \
		--branch $(ENCLAVE_MOCK_BRANCH) \
		--repo $(ENCLAVE_MOCK_REPO)

clean-enclave-mock:
	rm -rf enclave-mock

run-mock: build
	./$(BINARY) --enclave-dir enclave-mock \
		--tls-cert hack/tls/server.crt --tls-key hack/tls/server.key

dev: build-ui
	@mkdir -p hack/tls
	@test -f hack/tls/server.crt || openssl req -new -x509 -nodes -days 365 \
		-subj "/CN=localhost" -keyout hack/tls/server.key -out hack/tls/server.crt 2>/dev/null
	$(GO) build -ldflags="-w -s" -tags dev -o $(BINARY) .
	./$(BINARY) --no-auth --enclave-dir enclave-mock \
		--password-file /tmp/enclave-wizard-dev-pass \
		--tls-cert hack/tls/server.crt --tls-key hack/tls/server.key
