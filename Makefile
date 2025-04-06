BINARY_NAME=zap-store

build:
	go build

run: build
	./${BINARY_NAME} $(ARGS)

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