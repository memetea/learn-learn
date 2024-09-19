
.PHONY: swag-init run build build-swag test clean

run:
	go run -tags swagger ./cmd/server

swag-init:
	swag init -g cmd/server/main.go

build:
	go build -o learn-server cmd/main.go

build-swag:
	go build -tags=swagger -o learn-server ./cmd/server

test:
	go test ./...

clean:
	rm -f learn-server

