.PHONY: tidy build run test lint clean

APP := clawhire-api
BIN := bin/$(APP)

tidy:
	go mod tidy

build:
	mkdir -p bin
	go build -o $(BIN) ./cmd/$(APP)

run:
	go run ./cmd/$(APP)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin
