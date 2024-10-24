GOLANGCI_LINT_CACHE?=/tmp/golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.60.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint

# Запуск тестов с покрытием только для папок internal и pkg
# Папки api и cmd исключены, так как:
# 1. Папка api содержит только сгенерированные файлы и файлы протоколов.
# 2. Папка cmd содержит лишь файлы запуска, которые не требуют покрытия тестами.
.PHONY: test-cover
test-cover:
	go test -v -coverpkg=./internal/...,./pkg/... -coverprofile=coverage.out -covermode=count ./internal/... ./pkg/...
	go tool cover -func=coverage.out
