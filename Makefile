export GO111MODULE := on

build:
	go build -o bin/p24 .

test:
	go test -race -v -mod=vendor -timeout=60s -count 1 ./...

lint:
	golangci-lint run --config .golangci.yml ./...

lint-fix:
	golangci-lint run --config .golangci.yml --fix ./...

docker:
	docker build -t p24 .

.DEFAULT_GOAL := test
