SQLC_DB_FILES        := $(wildcard assets/*.sql)
SQLC_GENERATED_FILES := $(wildcard internal/db/*.sql.go) internal/db/db.go internal/db/models.go
SQLC_CONFIG          := assets/sqlc.yaml
SQLC_SENTINEL        := assets/.sqlc.sentinel

$(SQLC_SENTINEL): $(SQLC_DB_FILES) $(SQLC_CONFIG)
	sqlc generate -f $(SQLC_CONFIG)
	@touch $@

.PHONY: clean
clean:
	rm $(SQLC_GENERATED_FILES) $(SQLC_SENTINEL)

.PHONY: test
test:
	go test -v -race -shuffle=on ./...

run:
	go run ./...
