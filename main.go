package main

import (
	"flag"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type BenchmarkResult struct {
	Goroutines   int
	OpsPerSecond float64
	MemoryAllocs uint64
	TotalOps     uint64
}

type Benchmark struct {
	duration    time.Duration
	maxRoutines int
	testType    string
	verbose     bool
}

func main() {
	duration := flag.Int("duration", 10, "Duration of each test in seconds")
	maxRoutines := flag.Int("max", 0, "Maximum number of goroutines to test (0 = 2x CPU cores)")
	testType := flag.String("type", "both", "Type of test: cpu, memory, or both")
	verbose := flag.Bool("v", false, "Verbose output showing progress")
	flag.Parse()

	if *maxRoutines == 0 {
		*maxRoutines = runtime.NumCPU() * 2
	}

	bench := &Benchmark{
		duration:    time.Duration(*duration) * time.Second,
		maxRoutines: *maxRoutines,
		testType:    *testType,
		verbose:     *verbose,
	}

	fmt.Printf("🚀 Go Goroutine Benchmark Tool\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("Test Duration: %v per configuration\n", bench.duration)
	fmt.Printf("Max Goroutines: %d\n", bench.maxRoutines)
	fmt.Printf("Test Type: %s\n\n", bench.testType)

	if bench.testType == "cpu" || bench.testType == "both" {
		fmt.Printf("📊 CPU Benchmark\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		results := bench.runCPUBenchmark()
		bench.printResults(results, "CPU")
	}

	if bench.testType == "memory" || bench.testType == "both" {
		fmt.Printf("\n📊 Memory Benchmark\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		results := bench.runMemoryBenchmark()
		bench.printResults(results, "Memory")
	}
}

func (b *Benchmark) runCPUBenchmark() []BenchmarkResult {
	var results []BenchmarkResult
	testCases := b.getTestCases()

	for _, numGoroutines := range testCases {
		fmt.Printf("\n▶ Testing with %d goroutines...\n", numGoroutines)
		result := b.benchmarkCPU(numGoroutines)
		results = append(results, result)
		
		fmt.Printf("  ✓ Operations: %d | Rate: %.2f ops/sec\n", 
			result.TotalOps, result.OpsPerSecond)
	}

	return results
}

func (b *Benchmark) benchmarkCPU(numGoroutines int) BenchmarkResult {
	var totalOps uint64
	var wg sync.WaitGroup
	stop := make(chan bool)

	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()

	startTime := time.Now()
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var localOps uint64
			for {
				select {
				case <-stop:
					atomic.AddUint64(&totalOps, localOps)
					return
				default:
					result := 0.0
					for j := 0; j < 100; j++ {
						result += math.Sqrt(float64(j))
						result *= 1.0001
						result = math.Sin(result) + math.Cos(result)
					}
					localOps++
					
					if localOps%10000 == 0 {
						atomic.AddUint64(&totalOps, 10000)
						localOps -= 10000
					}
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case <-progressTicker.C:
				if b.verbose {
					elapsed := time.Since(startTime)
					progress := float64(elapsed) / float64(b.duration) * 100
					ops := atomic.LoadUint64(&totalOps)
					fmt.Printf("  Progress: %.0f%% | Ops: %d\r", progress, ops)
				}
			case <-stop:
				return
			}
		}
	}()

	time.Sleep(b.duration)
	close(stop)
	wg.Wait()

	finalOps := atomic.LoadUint64(&totalOps)
	opsPerSecond := float64(finalOps) / b.duration.Seconds()

	return BenchmarkResult{
		Goroutines:   numGoroutines,
		OpsPerSecond: opsPerSecond,
		TotalOps:     finalOps,
	}
}

func (b *Benchmark) runMemoryBenchmark() []BenchmarkResult {
	var results []BenchmarkResult
	testCases := b.getTestCases()

	for _, numGoroutines := range testCases {
		fmt.Printf("\n▶ Testing with %d goroutines...\n", numGoroutines)
		result := b.benchmarkMemory(numGoroutines)
		results = append(results, result)
		
		fmt.Printf("  ✓ Allocations: %d | Rate: %.2f allocs/sec\n", 
			result.MemoryAllocs, result.OpsPerSecond)
	}

	return results
}

func (b *Benchmark) benchmarkMemory(numGoroutines int) BenchmarkResult {
	var totalAllocs uint64
	var wg sync.WaitGroup
	stop := make(chan bool)

	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var localAllocs uint64
			for {
				select {
				case <-stop:
					atomic.AddUint64(&totalAllocs, localAllocs)
					return
				default:
					sizes := []int{64, 256, 1024, 4096}
					for _, size := range sizes {
						buffer := make([]byte, size)
						for j := range buffer {
							buffer[j] = byte(j % 256)
						}
						_ = buffer
						localAllocs++
					}
					
					if localAllocs%1000 == 0 {
						atomic.AddUint64(&totalAllocs, 1000)
						localAllocs -= 1000
					}
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case <-progressTicker.C:
				if b.verbose {
					elapsed := time.Since(startTime)
					progress := float64(elapsed) / float64(b.duration) * 100
					allocs := atomic.LoadUint64(&totalAllocs)
					fmt.Printf("  Progress: %.0f%% | Allocs: %d\r", progress, allocs)
				}
			case <-stop:
				return
			}
		}
	}()

	time.Sleep(b.duration)
	close(stop)
	wg.Wait()

	finalAllocs := atomic.LoadUint64(&totalAllocs)
	allocsPerSecond := float64(finalAllocs) / b.duration.Seconds()

	return BenchmarkResult{
		Goroutines:   numGoroutines,
		OpsPerSecond: allocsPerSecond,
		MemoryAllocs: finalAllocs,
		TotalOps:     finalAllocs,
	}
}

func (b *Benchmark) getTestCases() []int {
	cpuCount := runtime.NumCPU()
	testCases := []int{1}
	
	if cpuCount > 1 {
		testCases = append(testCases, cpuCount/2)
	}
	testCases = append(testCases, cpuCount)
	testCases = append(testCases, cpuCount*2)
	
	if b.maxRoutines > cpuCount*2 {
		step := (b.maxRoutines - cpuCount*2) / 3
		if step > 0 {
			for i := cpuCount*2 + step; i <= b.maxRoutines; i += step {
				testCases = append(testCases, i)
			}
		}
		if testCases[len(testCases)-1] != b.maxRoutines {
			testCases = append(testCases, b.maxRoutines)
		}
	}
	
	return testCases
}

func (b *Benchmark) printResults(results []BenchmarkResult, testType string) {
	if len(results) == 0 {
		return
	}

	fmt.Printf("\n📈 %s Benchmark Results\n", testType)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%-12s | %-15s | %-12s\n", "Goroutines", "Ops/Second", "Total Ops")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	var bestResult BenchmarkResult
	bestPerformance := 0.0

	for _, result := range results {
		fmt.Printf("%-12d | %-15.2f | %-12d\n", 
			result.Goroutines, result.OpsPerSecond, result.TotalOps)
		
		if result.OpsPerSecond > bestPerformance {
			bestPerformance = result.OpsPerSecond
			bestResult = result
		}
	}

	fmt.Printf("\n🏆 Optimal Configuration for %s:\n", testType)
	fmt.Printf("   Goroutines: %d\n", bestResult.Goroutines)
	fmt.Printf("   Performance: %.2f ops/sec\n", bestResult.OpsPerSecond)
	
	cpuRatio := float64(bestResult.Goroutines) / float64(runtime.NumCPU())
	fmt.Printf("   CPU Ratio: %.2fx\n", cpuRatio)
}