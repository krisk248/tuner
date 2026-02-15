package benchmark

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	testFile     = "/tmp/tuner-bench-test"
	seqBlockSize = 1024 * 1024 // 1 MB
	seqBlocks    = 256         // 256 MB total
	rndBlockSize = 4096        // 4 KB
	rndOps       = 4096        // 4096 random ops
)

// DiskResult holds disk benchmark results.
type DiskResult struct {
	SeqWriteMBps float64
	SeqReadMBps  float64
	Rnd4KWriteIOPS float64
	Rnd4KReadIOPS  float64
}

// RunDiskBenchmark performs sequential and random I/O benchmarks.
func RunDiskBenchmark() (*DiskResult, error) {
	result := &DiskResult{}

	defer os.Remove(testFile)

	// Sequential write
	fmt.Println("Sequential write (256 MB)...")
	bar := progressbar.Default(int64(seqBlocks))
	data := make([]byte, seqBlockSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	f, err := os.Create(testFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create test file: %w", err)
	}

	start := time.Now()
	for i := 0; i < seqBlocks; i++ {
		if _, err := f.Write(data); err != nil {
			f.Close()
			return nil, fmt.Errorf("write error: %w", err)
		}
		bar.Add(1)
	}
	f.Sync()
	f.Close()
	elapsed := time.Since(start).Seconds()
	result.SeqWriteMBps = float64(seqBlocks) / elapsed
	fmt.Printf("  %.1f MB/s\n\n", result.SeqWriteMBps)

	// Drop caches before read test (needs root, skip if not available)
	dropCaches()

	// Sequential read
	fmt.Println("Sequential read (256 MB)...")
	bar = progressbar.Default(int64(seqBlocks))
	readBuf := make([]byte, seqBlockSize)

	f, err = os.Open(testFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open test file: %w", err)
	}

	start = time.Now()
	for i := 0; i < seqBlocks; i++ {
		if _, err := f.Read(readBuf); err != nil {
			f.Close()
			return nil, fmt.Errorf("read error: %w", err)
		}
		bar.Add(1)
	}
	f.Close()
	elapsed = time.Since(start).Seconds()
	result.SeqReadMBps = float64(seqBlocks) / elapsed
	fmt.Printf("  %.1f MB/s\n\n", result.SeqReadMBps)

	os.Remove(testFile)

	// Random 4K write
	fmt.Println("Random 4K write (4096 ops)...")
	bar = progressbar.Default(int64(rndOps))
	smallData := make([]byte, rndBlockSize)
	for i := range smallData {
		smallData[i] = 0xAA
	}

	f, err = os.Create(testFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create test file: %w", err)
	}
	// Pre-allocate
	f.Truncate(int64(rndOps * rndBlockSize))

	start = time.Now()
	for i := 0; i < rndOps; i++ {
		offset := int64((i * 7919) % rndOps) * int64(rndBlockSize) // pseudo-random spread
		f.WriteAt(smallData, offset)
		bar.Add(1)
	}
	f.Sync()
	f.Close()
	elapsed = time.Since(start).Seconds()
	result.Rnd4KWriteIOPS = float64(rndOps) / elapsed
	fmt.Printf("  %.0f IOPS\n\n", result.Rnd4KWriteIOPS)

	dropCaches()

	// Random 4K read
	fmt.Println("Random 4K read (4096 ops)...")
	bar = progressbar.Default(int64(rndOps))
	readSmall := make([]byte, rndBlockSize)

	f, err = os.Open(testFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open test file: %w", err)
	}

	start = time.Now()
	for i := 0; i < rndOps; i++ {
		offset := int64((i * 7919) % rndOps) * int64(rndBlockSize)
		f.ReadAt(readSmall, offset)
		bar.Add(1)
	}
	f.Close()
	elapsed = time.Since(start).Seconds()
	result.Rnd4KReadIOPS = float64(rndOps) / elapsed
	fmt.Printf("  %.0f IOPS\n\n", result.Rnd4KReadIOPS)

	os.Remove(testFile)

	return result, nil
}

func dropCaches() {
	// Best effort - requires root
	os.WriteFile("/proc/sys/vm/drop_caches", []byte("3"), 0644)
}

// PrintDiskResult displays the disk benchmark results.
func PrintDiskResult(r *DiskResult) {
	fmt.Println("── Disk Benchmark Results ──")
	fmt.Printf("  Sequential Write   %.1f MB/s\n", r.SeqWriteMBps)
	fmt.Printf("  Sequential Read    %.1f MB/s\n", r.SeqReadMBps)
	fmt.Printf("  Random 4K Write    %.0f IOPS\n", r.Rnd4KWriteIOPS)
	fmt.Printf("  Random 4K Read     %.0f IOPS\n", r.Rnd4KReadIOPS)
}
