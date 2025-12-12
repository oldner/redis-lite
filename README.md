# Redis-Lite

![Go Version](https://img.shields.io/github/go-mod/go-version/oldner/redis-lite)
![Build Status](https://github.com/oldner/redis-lite/actions/workflows/go.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Redis-Lite** is a lightweight, multi-threaded, in-memory key-value store built from scratch in Go. 

It is designed to showcase core backend engineering concepts: **TCP networking**, **Concurrency (Mutexes & Goroutines)**, **Sharding**, and **Protocol Parsing (RESP)**.

## 🚀 Features

- **In-Memory Storage**: High-performance reads/writes using native Go maps.
- **Concurrent & Thread-Safe**: Uses `sync.RWMutex` with **Sharding** (256 shards) to minimize lock contention.
- **RESP Compatible**: Speaks the Redis Serialization Protocol (can connect via `redis-cli`).
- **TTL Support**: Keys automatically expire after a set duration.
- **Supported Commands**:
  - `SET key value [ttl]`
  - `GET key`
  - `DEL key`
  - `HSET key field value`
  - `HGET key field`
  - `PING`

## 🛠️ Installation & Usage

### Prerequisites
- Go 1.22+

### Running the Server
```bash
# Clone the repo
git clone [https://github.com/YOUR_USERNAME/redis-lite.git](https://github.com/YOUR_USERNAME/redis-lite.git)
cd redis-lite

# Run the server (Defaults to port 6379)
go run cmd/server/main.go
```

### Connecting with the Built-in Client

We include a simple CLI tool to interact with the server:

```bash
# Open a new terminal
go run cmd/client/main.go --host localhost --port 6379

> SET mykey "Hello World"
+OK
> GET mykey
$11
Hello World
```

## 🏗️ Architecture

The project follows a modular structure to separate the Network Layer from the Data Layer.
- `cmd/server`: Entry point, handles configuration and wiring.
- `internal/server`: TCP listener and connection handling (Networking).
- `pkg/database`: The core storage engine (Sharding, Locking, Janitor).

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
