SQLC_DB_FILES        := $(wildcard assets/*.sql)
SQLC_GENERATED_FILES := $(wildcard internal/db/*.sql.go) internal/db/db.go internal/db/models.go
SQLC_CONFIG          := assets/sqlc.yaml

OPENAPI_FILES           := $(wildcard assets/api/*.yaml)
OPENAPI_GENERATED_FILES := $(wildcard internal/api/*.gen.go)

CODEGEN_SENTINEL := .codegen.sentinel

$(CODEGEN_SENTINEL): $(OPENAPI_FILES) $(SQLC_DB_FILES) $(SQLC_CONFIG)
	go generate ./...
	@touch $@

.PHONY: codegen
codegen: $(CODEGEN_SENTINEL)

.PHONY: clean
clean:
	rm $(SQLC_GENERATED_FILES) $(CODEGEN_SENTINEL) $(OPENAPI_GENERATED_FILES)

.PHONY: test
test:
	go test -race -shuffle=on ./...

.PHONY: run
run: fmt
	go run ./...

.PHONY: fmt
fmt:
	gofmt -w . && \
		golines -w .

.PHONY: vulncheck
vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: lint
lint:
	@test -z "$$(gofmt -l .)" || (echo "gofmt: unformatted files:" && gofmt -l . && exit 1)
	golangci-lint run


#####################################
# // test coverage
#
# go test ./... -coverprofile=coverage.out -coverpkg=./...
# go tool cover -html=coverage.out -o coverage.html
