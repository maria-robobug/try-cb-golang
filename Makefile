build:
	go build -o try-cb-golang

test:
	go test -race -cover ./...

run:
	go run -race main.go
