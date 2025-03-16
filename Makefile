BINARY_NAME=kv-store

build:
	go build

run: build
	./${BINARY_NAME}

clean:
	go clean

test:
	go test ./... -v

bench:
	go test -bench=. -benchmem ./...

save_benchmark_ts:
	mkdir -p benchmarks
	go test ./... -bench=. -benchmem > benchmarks/bench_results_$$(date +%Y%m%d_%H%M%S).txt

save_benchmark:
	mkdir -p benchmarks
	go test ./... -bench=. -benchmem > benchmarks/bench_results_$$(date +%Y%m%d).txt