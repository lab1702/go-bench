package main

import (
	"flag"
	"fmt"
	"math"
	"runtime"
	"strconv"
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
}

func formatNumber(n float64) string {
	intPart := int64(n)
	str := strconv.FormatInt(intPart, 10)

	if len(str) <= 3 {
		return fmt.Sprintf("%.2f", n)
	}

	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}

	decimalPart := n - float64(intPart)
	if decimalPart > 0 {
		return fmt.Sprintf("%s.%02d", result, int(decimalPart*100))
	}
	return result + ".00"
}

func main() {
	duration := flag.Int("d", 10, "Duration of each test in seconds")
	testType := flag.String("t", "both", "Type of test: cpu, memory, or both")
	flag.Parse()

	maxRoutines := runtime.NumCPU() * 2

	bench := &Benchmark{
		duration:    time.Duration(*duration) * time.Second,
		maxRoutines: maxRoutines,
		testType:    *testType,
	}

	fmt.Printf("🚀 Go Goroutine Benchmark Tool\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Logical CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("Test Duration: %v per configuration\n", bench.duration)
	fmt.Printf("Max Goroutines: %d\n", bench.maxRoutines)
	fmt.Printf("Test Type: %s\n\n", bench.testType)

	var cpuResults []BenchmarkResult
	var memoryResults []BenchmarkResult

	if bench.testType == "cpu" || bench.testType == "both" {
		fmt.Printf("📊 CPU Benchmark\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		cpuResults = bench.runCPUBenchmark()
	}

	if bench.testType == "memory" || bench.testType == "both" {
		fmt.Printf("\n📊 Memory Benchmark\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		memoryResults = bench.runMemoryBenchmark()
	}

	// Print all results at the end
	if len(cpuResults) > 0 {
		bench.printResults(cpuResults, "CPU")
	}

	if len(memoryResults) > 0 {
		bench.printResults(memoryResults, "Memory")
	}

	fmt.Println()
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
				// Progress ticker removed with verbose option
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
				// Progress ticker removed with verbose option
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

	upperLimit := cpuCount + (cpuCount / 2)
	if upperLimit > cpuCount {
		testCases = append(testCases, upperLimit)
	}

	// Add test for 2x CPU count (which is our max)
	if b.maxRoutines >= cpuCount*2 {
		testCases = append(testCases, cpuCount*2)
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
	var singleGoroutineResult BenchmarkResult
	bestPerformance := 0.0

	for _, result := range results {
		fmt.Printf("%-12d | %-15.2f | %-12d\n",
			result.Goroutines, result.OpsPerSecond, result.TotalOps)

		if result.Goroutines == 1 {
			singleGoroutineResult = result
		}

		if result.OpsPerSecond > bestPerformance {
			bestPerformance = result.OpsPerSecond
			bestResult = result
		}
	}

	fmt.Printf("\n🏆 Optimal Configuration for %s:\n", testType)
	fmt.Printf("   Goroutines: %d\n", bestResult.Goroutines)
	cpuRatio := float64(bestResult.Goroutines) / float64(runtime.NumCPU())
	fmt.Printf("   CPU Ratio: %.2fx\n", cpuRatio)
	fmt.Printf("   Performance: %s ops/sec\n", formatNumber(bestResult.OpsPerSecond))

	if singleGoroutineResult.Goroutines == 1 {
		fmt.Printf("   Single Goroutine: %s ops/sec\n", formatNumber(singleGoroutineResult.OpsPerSecond))
		speedup := bestResult.OpsPerSecond / singleGoroutineResult.OpsPerSecond
		fmt.Printf("   Speed Increase: %.2fx vs single goroutine\n", speedup)
	}
}
