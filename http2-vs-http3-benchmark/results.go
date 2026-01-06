package main

import (
	"fmt"
	"strings"
	"time"
)

type ResultsReport struct {
	results []BenchmarkResult
}

func NewResultsReport(results []BenchmarkResult) *ResultsReport {
	return &ResultsReport{results: results}
}

func (r *ResultsReport) Print() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  BENCHMARK RESULTS SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	currentScenario := ""
	for _, result := range r.results {
		if result.Scenario != currentScenario {
			currentScenario = result.Scenario
			fmt.Printf("\n[%s]\n", currentScenario)
			fmt.Println(strings.Repeat("-", 80))
		}

		r.printResult(result)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  COMPARATIVE ANALYSIS")
	fmt.Println(strings.Repeat("=", 80))
	r.printComparison()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  WINNER BY SCENARIO")
	fmt.Println(strings.Repeat("=", 80))
	r.printWinners()
}

func (r *ResultsReport) printResult(result BenchmarkResult) {
	fmt.Printf("\n%s:\n", result.Protocol)
	fmt.Printf("  Total Requests:     %d\n", result.TotalRequests)
	fmt.Printf("  Success:            %d\n", result.SuccessRequests)
	fmt.Printf("  Failed:             %d\n", result.FailedRequests)
	fmt.Printf("  Total Duration:     %s\n", result.TotalDuration.Round(time.Millisecond))
	fmt.Printf("  Requests/sec:       %.2f\n", result.RequestsPerSec)

	if result.SuccessRequests > 0 {
		fmt.Printf("\n  Latency:\n")
		fmt.Printf("    Min:              %s\n", result.MinLatency.Round(time.Millisecond))
		fmt.Printf("    Avg:              %s\n", result.AvgLatency.Round(time.Millisecond))
		fmt.Printf("    P50:              %s\n", result.P50Latency.Round(time.Millisecond))
		fmt.Printf("    P95:              %s\n", result.P95Latency.Round(time.Millisecond))
		fmt.Printf("    P99:              %s\n", result.P99Latency.Round(time.Millisecond))
		fmt.Printf("    Max:              %s\n", result.MaxLatency.Round(time.Millisecond))

		if result.TotalBytesRecv > 0 {
			fmt.Printf("\n  Throughput:         %.2f MB/s\n", result.Throughput)
			fmt.Printf("  Total Data:         %.2f MB\n", float64(result.TotalBytesRecv)/1024/1024)
		}
	}
}

func (r *ResultsReport) printComparison() {
	scenarios := make(map[string][]*BenchmarkResult)

	for i := range r.results {
		scenario := r.results[i].Scenario
		scenarios[scenario] = append(scenarios[scenario], &r.results[i])
	}

	for scenario, results := range scenarios {
		if len(results) != 2 {
			continue
		}

		fmt.Printf("\n[%s]\n", scenario)
		fmt.Println(strings.Repeat("-", 80))

		http2 := results[0]
		http3 := results[1]
		if http2.Protocol == "HTTP/3" {
			http2, http3 = http3, http2
		}

		if http2.SuccessRequests > 0 && http3.SuccessRequests > 0 {
			latencyImprovement := ((http2.AvgLatency.Seconds() - http3.AvgLatency.Seconds()) / http2.AvgLatency.Seconds()) * 100
			throughputImprovement := ((http3.Throughput - http2.Throughput) / http2.Throughput) * 100
			rpsImprovement := ((http3.RequestsPerSec - http2.RequestsPerSec) / http2.RequestsPerSec) * 100

			fmt.Printf("\nHTTP/3 vs HTTP/2:\n")
			r.printMetricComparison("Average Latency", latencyImprovement, http2.AvgLatency, http3.AvgLatency, true)
			r.printMetricComparison("P95 Latency",
				((http2.P95Latency.Seconds() - http3.P95Latency.Seconds()) / http2.P95Latency.Seconds()) * 100,
				http2.P95Latency, http3.P95Latency, true)
			r.printMetricComparison("Requests/sec", rpsImprovement, http2.RequestsPerSec, http3.RequestsPerSec, false)

			if http2.TotalBytesRecv > 0 && http3.TotalBytesRecv > 0 {
				r.printMetricComparison("Throughput (MB/s)", throughputImprovement, http2.Throughput, http3.Throughput, false)
			}
		}
	}
}

func (r *ResultsReport) printMetricComparison(metric string, improvement float64, http2Val, http3Val interface{}, lowerIsBetter bool) {
	var symbol string
	var status string

	if lowerIsBetter {
		if improvement > 0 {
			symbol = "↓"
			status = "better"
		} else {
			symbol = "↑"
			status = "worse"
			improvement = -improvement
		}
	} else {
		if improvement > 0 {
			symbol = "↑"
			status = "better"
		} else {
			symbol = "↓"
			status = "worse"
			improvement = -improvement
		}
	}

	fmt.Printf("  %-20s %s %.1f%% %s", metric+":", symbol, improvement, status)

	switch v := http2Val.(type) {
	case time.Duration:
		fmt.Printf(" (HTTP/2: %v, HTTP/3: %v)\n",
			v.Round(time.Millisecond),
			http3Val.(time.Duration).Round(time.Millisecond))
	case float64:
		fmt.Printf(" (HTTP/2: %.2f, HTTP/3: %.2f)\n", v, http3Val.(float64))
	}
}

func (r *ResultsReport) printWinners() {
	scenarios := make(map[string][]*BenchmarkResult)

	for i := range r.results {
		scenario := r.results[i].Scenario
		scenarios[scenario] = append(scenarios[scenario], &r.results[i])
	}

	http2Wins := 0
	http3Wins := 0
	ties := 0

	for scenario, results := range scenarios {
		if len(results) != 2 {
			continue
		}

		http2 := results[0]
		http3 := results[1]
		if http2.Protocol == "HTTP/3" {
			http2, http3 = http3, http2
		}

		var winner string
		if http3.AvgLatency < http2.AvgLatency && http3.RequestsPerSec > http2.RequestsPerSec {
			winner = "HTTP/3"
			http3Wins++
		} else if http2.AvgLatency < http3.AvgLatency && http2.RequestsPerSec > http3.RequestsPerSec {
			winner = "HTTP/2"
			http2Wins++
		} else {
			winner = "TIE"
			ties++
		}

		fmt.Printf("\n%-40s Winner: %s\n", scenario, winner)
	}

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Printf("\nOverall Score:\n")
	fmt.Printf("  HTTP/2: %d wins\n", http2Wins)
	fmt.Printf("  HTTP/3: %d wins\n", http3Wins)
	fmt.Printf("  Ties:   %d\n", ties)

	if http3Wins > http2Wins {
		fmt.Printf("\n🏆 HTTP/3 is the overall winner!\n")
	} else if http2Wins > http3Wins {
		fmt.Printf("\n🏆 HTTP/2 is the overall winner!\n")
	} else {
		fmt.Printf("\n🤝 It's a tie! Both protocols performed equally.\n")
	}
}

func (r *ResultsReport) SaveToFile(filename string) error {
	return nil
}
