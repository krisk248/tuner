package benchmark

import (
	"fmt"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

// NetworkResult holds network benchmark results.
type NetworkResult struct {
	Server    string
	Latency   time.Duration
	Download  float64 // Mbps
	Upload    float64 // Mbps
}

// RunNetworkBenchmark performs a speedtest.
func RunNetworkBenchmark() (*NetworkResult, error) {
	fmt.Println("Finding best server...")

	client := speedtest.New()
	servers, err := client.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no speedtest servers found")
	}

	// Pick the closest server
	server := servers[0]

	result := &NetworkResult{
		Server: fmt.Sprintf("%s (%s)", server.Name, server.Country),
	}

	fmt.Printf("Server: %s\n", result.Server)

	// Latency
	fmt.Println("Testing latency...")
	if err := server.PingTest(nil); err != nil {
		return nil, fmt.Errorf("ping test failed: %w", err)
	}
	result.Latency = server.Latency
	fmt.Printf("  Latency: %v\n\n", result.Latency)

	// Download
	fmt.Println("Testing download speed...")
	if err := server.DownloadTest(); err != nil {
		return nil, fmt.Errorf("download test failed: %w", err)
	}
	result.Download = float64(server.DLSpeed)
	fmt.Printf("  Download: %.1f Mbps\n\n", result.Download)

	// Upload
	fmt.Println("Testing upload speed...")
	if err := server.UploadTest(); err != nil {
		return nil, fmt.Errorf("upload test failed: %w", err)
	}
	result.Upload = float64(server.ULSpeed)
	fmt.Printf("  Upload: %.1f Mbps\n\n", result.Upload)

	return result, nil
}

// PrintNetworkResult displays the network benchmark results.
func PrintNetworkResult(r *NetworkResult) {
	fmt.Println("── Network Benchmark Results ──")
	fmt.Printf("  Server      %s\n", r.Server)
	fmt.Printf("  Latency     %v\n", r.Latency)
	fmt.Printf("  Download    %.1f Mbps\n", r.Download)
	fmt.Printf("  Upload      %.1f Mbps\n", r.Upload)
}
