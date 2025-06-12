# Go CRUD-Bench

This is the Go implementation of the crud-bench benchmarking tool, an open-source benchmarking tool for testing and comparing the performance of various databases across different workloads.

## Features

- Benchmarks CRUD operations (Create, Read, Update, Delete)
- Supports scan operations with various projections
- Configurable concurrency with multiple clients and threads
- Automatic Docker container management for database instances
- Customizable data generation with templating
- Support for various key types (integer, string, UUID)

## Supported Databases

Currently implemented:

- MySQL

Planned implementations:

- SQLite
- PostgreSQL
- MongoDB
- Redis
- RocksDB
- SurrealDB
- And more...

## Requirements

- [Go](https://golang.org/) 1.22 or higher
- [Docker](https://www.docker.com/) (optional, for containerized database testing)

## Installation

### From Source

```bash
git clone https://github.com/surrealdb/go-crud-bench.git
cd go-crud-bench
make build
```

The binary will be available in the `bin` directory.

## Usage

```
Usage: crud-bench [OPTIONS] --database <DATABASE> --samples <SAMPLES>

Options:
  -n, --name string        An optional name for the test, used as a suffix for the JSON result file name
  -d, --database string    The database to benchmark (required)
  -i, --image string       Specify a custom Docker image
  -p, --privileged         Whether to run Docker in privileged mode
  -e, --endpoint string    Specify a custom endpoint to connect to
  -b, --blocking int       Maximum number of blocking threads (default 12)
  -w, --workers int        Number of async runtime workers (default 12)
  -c, --clients int        Number of concurrent clients (default 1)
  -t, --threads int        Number of concurrent threads per client (default 1)
  -s, --samples int        Number of samples to be created, read, updated, and deleted (required)
  -r, --random             Generate the keys in a pseudo-randomized order
  -k, --key string         The type of the key (default "integer")
  -v, --value string       Size of the text value (default "{\n\t\"text\": \"string:50\",\n\t\"integer\": \"int\"\n}")
      --show-sample        Print-out an example of a generated value
      --pid int            Collect system information for a given pid
  -a, --scans string       An array of scan specifications
```

### Examples

#### MySQL Benchmark

```bash
# Using Docker container (automatically started)
make mysql-test

# Using existing MySQL instance
make mysql-custom
```

Or directly:

```bash
./bin/crud-bench -d mysql -s 10000 -c 4 -t 8
```

## Value Templates

You can customize the data being inserted using value templates. For example:

```json
{
  "text": "text:30",
  "text_range": "text:10..50",
  "bool": "bool",
  "string_enum": "enum:foo,bar",
  "datetime": "datetime",
  "float": "float",
  "float_range": "float:1..10",
  "float_enum": "float:1.1,2.2,3.3",
  "integer": "int",
  "integer_range": "int:1..5",
  "integer_enum": "int:1,2,3",
  "uuid": "uuid",
  "nested": {
    "text": "text:100",
    "array": [
      "string:10",
      "string:2..5"
    ]
  }
}
```

## Scan Configuration

You can customize scan operations using the `--scans` parameter:

```json
[
  {
    "name": "limit100",
    "projection": "FULL",
    "start": 0,
    "limit": 100,
    "expect": 100
  },
  {
    "name": "start100",
    "projection": "ID",
    "start": 100,
    "limit": 100,
    "expect": 100
  }
]
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the same license as the original crud-bench project.
