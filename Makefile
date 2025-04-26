SERVER_BINARY_NAME=zapstore-server
CLI_BINARY_NAME=zapstore-cli

ARGS ?= -engine bitcask -dataDir data

build-server:
	go build -o zapstore-server cmd/server/main.go

build-cli:
	go build -o zapstore-cli cmd/cli/main.go

run-server: build-server
	./${SERVER_BINARY_NAME} $(ARGS)

run-cli: build-cli
	./${CLI_BINARY_NAME}


clean:
	go clean

test:
	go test ./... -v

bench:
	go test -bench=. -benchmem ./...

# Saves the benchmarks with timestamp, used when tweaking for performance
save_benchmark_ts:
	mkdir -p benchmarks
	go test ./... -bench=. -benchmem > benchmarks/bench_results_$$(date +%Y%m%d_%H%M%S).txt

# Saves the benchmarks with date. This will probably be the optimal one I found for that day.
save_benchmark:
	mkdir -p benchmarks
	go test ./... -bench=. -benchmem > benchmarks/bench_results_$$(date +%Y%m%d).txt