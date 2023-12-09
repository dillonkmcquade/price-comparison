
build:
	go build -o bin/price-comparison

test:
	go test -v ./...

lint:
	staticcheck ./...
	gosec ./...

dev:
	concurrently --names "CLIENT,SERVER" -c "bgBlue.bold,bgMagenta.bold" "cd client && pnpm dev" "go run main.go"
