
.PHONY: swag run build build-swag test zip clean

run:
	go run -tags swagger ./cmd/server

swag:
	swag init -g cmd/server/main.go

build:
	go build -o learn-server cmd/main.go

build-swag:
	go build -tags=swagger -o learn-server ./cmd/server

test:
	go test ./...

zip-server:
	zip -r ../learn-server.zip . -x "*.db" -x "*.vscode/*"

zip-web:
	zip -r ../learn-web.zip ../quiz-management -x "*.db" -x "*node_modules/*" -x "*.vscode/*"

clean:
	rm -f learn-server

