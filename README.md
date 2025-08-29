# go-bench

A simple Go benchmarking tool to measure optimal goroutine counts for CPU and memory-intensive operations.

## Installation

### Install directly from GitHub
```bash
go install github.com/lab1702/go-bench@latest
```

### Build from source
```bash
git clone https://github.com/lab1702/go-bench.git
cd go-bench
go build
```

## Usage

```bash
# Run both CPU and memory benchmarks (default)
go-bench

# Run only CPU benchmark
go-bench -t cpu

# Run only memory benchmark  
go-bench -t memory

# Adjust test duration (in seconds)
go-bench -d 5
```

## Command-line Options

- `-t` : Type of test to run (`cpu`, `memory`, or `both`). Default: `both`
- `-d` : Duration of each test configuration in seconds. Default: `10`

## What it does

The tool benchmarks your system with different goroutine counts to find the optimal configuration:
- 1 goroutine (baseline single-threaded performance)
- CPU cores / 2
- CPU cores
- CPU cores * 1.5
- CPU cores * 2

For each configuration, it measures:
- Operations per second
- Total operations completed
- Speed increase compared to single goroutine
- CPU ratio (goroutines / CPU cores)

## Output

The tool provides:
- Real-time progress during benchmarking
- Detailed results table for each configuration
- Summary of optimal configuration
- Performance comparison against single-threaded baseline

## Requirements

- Go 1.25 or higher

## License

MIT License - see LICENSE file for details
