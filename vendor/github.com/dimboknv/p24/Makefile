export GO111MODULE := on

GO_TEST=go test -v -race -timeout=60s -count 1

lint:
	golangci-lint run --config .golangci.yml ./...

lint-fix:
	golangci-lint run --config .golangci.yml --fix ./...

test:
	$(GO_TEST) ./...

coverage:
	$(GO_TEST) -covermode=atomic -coverprofile=coverage.out ./...

coverage-html: coverage
	go tool cover -html=coverage.out

.DEFAULT_GOAL := test


