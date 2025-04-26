[![codecov](https://codecov.io/gh/kiran-4444/zap-store/graph/badge.svg?token=YTobWfz3N1)](https://codecov.io/gh/kiran-4444/zap-store)

# **ZapStore 🚀**

A high-performance, in-memory key-value store built in Go, inspired by Designing Data-Intensive Applications (DDIA). This project is my playground for mastering Go, exploring storage engine design, and diving into the world of distributed systems. It currently supports basic `Get`, `Set`, and `Del` operations with a thread-safe in-memory engine, and I’m actively working on adding a Bitcask storage engine and distributed capabilities.

## **🌟 Why This Project?**

I built ZapStore to deepen my understanding of key-value storage systems while honing my Go skills. Drawing inspiration from DDIA, I’m exploring the trade-offs of in-memory vs. disk-based storage, concurrency, and scalability. This project is a hands-on journey through the fundamentals of data-intensive applications, with a focus on clean code, performance optimization, and extensibility.

### **Key Features**

- **In-Memory Storage Engine**: A thread-safe, lightning-fast engine using Go’s sync.Mutex for concurrent access.
- **Basic Operations**: Supports Get, Set, and Del with robust error handling.
- **Performance Optimized**: Achieves sub-50 ns/op for all operations on an Apple M1 (see ).
- **Extensible Design**: Built with a pluggable StorageEngine interface to support future engines like Bitcask.
- **Automated Benchmarking**: Includes a Makefile to run and save benchmark results with timestamps for performance tracking.

## **🚀 Getting Started**

### **Prerequisites**

- Go 1.18 or higher
- A Unix-like shell (e.g., Bash on macOS/Linux) for the Makefile. On Windows, use Git Bash or WSL.

### **Installation**

1. Clone the repository:

```bash
git clone https://github.com/kiran-4444/zap-store.git
cd zap-store
```

1. Build the project:

```bash
make build
```

This creates a binary named `zap-store`.

### Usage

Run the ZapStore CLI:

```bash
make run
```

This starts an interactive CLI where you can issue commands like:

- `set key value` – Store a key-value pair.
- `get key` – Retrieve the value for a key.
- `del key` – Delete a key-value pair.
- `exit` – Quit the CLI.

Example session:

```bash
➜  zap-store git:(main) ✗ make run
go build
./zap-store
> set foo bar
> get foo
bar
> del foo
Deleted foo
> get foo
key not found
> exit
Exiting...
```

## 📊 Benchmarks

I’ve optimized the in-memory engine for performance, achieving impressive results on an Apple M1 (darwin/arm64):

| Operation | Time | Allocations |
| --- | --- | --- |
| `Set` | 22.82 ns/op | 0 B/op, 0 allocs/op |
| `Get` | 21.10 ns/op | 0 B/op, 0 allocs/op |
| `Del` | 42.88 ns/op | 0 B/op, 0 allocs/op |
| Mixed (50% Get, 40% Set, 10% Del) | 43.59 ns/op | 0 B/op, 0 allocs/op |

Run benchmarks yourself:

```bash
make bench
```

## 🛠️ Roadmap

- [x]  **In-Memory Storage Engine**: A thread-safe engine with sub-50 ns/op performance.
- [x]  **Bitcask Storage Engine**: Implement a disk-based engine inspired by Bitcask for persistence and larger datasets.
- [ ]  **Server-Client Architecture**: Transform ZapStore into a server that multiple clients can connect to, using a custom query language for interaction.
- [ ]  **Distributed System**: Add replication and sharding to make ZapStore distributed, exploring consistency and fault tolerance.

## 🧠 What I’ve Learned

**Go Concurrency**: Mastered `sync.Mutex` and `sync.RWMutex` to make the in-memory engine thread-safe, handling 8 goroutines on an Apple M1 with minimal contention.

**Performance Optimization**: Reduced allocations from 147 B/op to 0 B/op, improving Set from 423.8 ns/op to 22.82 ns/op through pre-allocation and benchmark tuning.

**Storage Design**: Designed a pluggable `StorageEngine` interface, preparing for Bitcask and future engines.

**Benchmarking**: Automated performance tracking with a `Makefile`, saving timestamped results for trend analysis.

## 🤝 Contributing

Contributions are welcome! Whether it’s adding features, optimizing performance, or fixing bugs, I’d love to collaborate. Here’s how to get started:

1. Fork the repository.
2. Create a branch: `git checkout -b feature/your-feature`.
3. Commit your changes: `git commit -m "feat: You feature"`.
4. Push to your fork: `git push origin feature/your-feature`.
5. Open a pull request.

Please include tests and update benchmarks if applicable. Run make test and `make bench` before submitting.

## 📚 Resources

[Designing Data-Intensive Applications by Martin Kleppmann](https://dataintensive.net/) – The inspiration for this project.

[Go Documentation](https://go.dev/doc/) – For learning Go best practices.

[Bitcask Paper](https://riak.com/assets/bitcask-intro.pdf) – The basis for the upcoming Bitcask engine.

## 📬 Contact

Feel free to reach out with questions, feedback, or collaboration ideas:

GitHub: [kiran-4444](https://github.com/kiran-4444)

Email: <chandrakiran.g19@gmail.com>
