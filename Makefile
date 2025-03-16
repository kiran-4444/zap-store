BINARY_NAME=kv-store

build:
	go build

run: build
	./${BINARY_NAME}

clean:
	go clean

test:
	go test ./... -v