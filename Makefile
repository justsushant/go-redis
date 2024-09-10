build:
	go build -o ./bin/go-redis .

run: build
	./bin/go-redis

test:
	go test ./...