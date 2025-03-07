package server_test

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	svrPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server"
	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/joho/godotenv"
)

// StressTestReport holds the results of the stress test
type StressTestReport struct {
	TotalRequests     int             // Total number of requests made
	Successful        int             // Number of successful requests
	Failed            int             // Number of failed requests
	MinResponseTime   time.Duration   // Minimum response time
	MaxResponseTime   time.Duration   // Maximum response time
	TotalResponseTime time.Duration   // Total response time (for calculating average)
	Errors            []string        // List of errors encountered
	StatusCodeCounts  map[int]int     // Count of HTTP status codes
	SuccessRate       float64         // Percentage of successful requests
	Throughput        float64         // Requests processed per second
	ResponseTimes     []time.Duration // All response times, for percentiles.
	RequestTimes      []time.Time     // All request start times, for request rate over time.
	ErrorTimes        []time.Time     // All error times, for error rate over time.
}

func TestStressTest(t *testing.T) {

	environment := os.Getenv("ENVIRONMENT")

	if environment == "" {
		godotenv.Load(".test.env")
		environment = os.Getenv("ENVIRONMENT")
	}

	if environment == "cicd" {
		t.Skip("Skipping stress test due to environment")
		return
	}

	bed := testPkg.SetupServerTestBed()
	apiBaseUrl := "http://localhost:8080"

	// Start the server in a goroutine
	go func() {
		svrPkg.RunServer(bed.Server, bed.Mgr.GetRoutes, nil)
	}()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Define the number of concurrent users and requests
	numUsers := 150    // Number of concurrent users
	numRequests := 100 // Number of requests per user

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numUsers)

	// Track the start time
	startTime := time.Now()

	// Create a report to track metrics
	report := StressTestReport{
		Errors:           make([]string, 0),
		StatusCodeCounts: make(map[int]int),
		ResponseTimes:    make([]time.Duration, 0),
		RequestTimes:     make([]time.Time, 0),
		ErrorTimes:       make([]time.Time, 0),
	}

	// Use a mutex to safely update the report from multiple goroutines
	var mu sync.Mutex

	// Simulate concurrent users
	for i := 0; i < numUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			// Simulate multiple requests per user
			for j := 0; j < numRequests; j++ {
				// Generate a unique IP address for each request
				ip := fmt.Sprintf("192.168.1.%d", userID)

				// Track the start time of the request
				requestStartTime := time.Now()

				// Hit the endpoint
				rr := testPkg.HitEndpoint(
					t,
					bed.Server.Root.ServeHTTP, // Handler function
					"POST",                    // HTTP method
					apiBaseUrl+"/private/api/mock-structs/new", // Full URL
					`{"field1": "value1", "field2": "value2"}`, // Request body
					true,          // Authentication requirement
					bed.AdminUser, // User for authentication
					bed.Logger,    // Logger
					ip,            // IP address
				)

				// Track the response time
				responseTime := time.Since(requestStartTime)

				// Update the report
				mu.Lock()
				report.TotalRequests++
				report.RequestTimes = append(report.RequestTimes, requestStartTime)
				report.ResponseTimes = append(report.ResponseTimes, responseTime)

				if rr.Code >= 200 && rr.Code < 400 {
					report.Successful++
				} else {
					report.Failed++
					report.Errors = append(report.Errors, fmt.Sprintf("Code %d. Response: %s", rr.Code, rr.Body))
					report.ErrorTimes = append(report.ErrorTimes, time.Now())
				}

				// Update status code counts
				report.StatusCodeCounts[rr.Code]++

				// Update min and max response times
				if report.MinResponseTime == 0 || responseTime < report.MinResponseTime {
					report.MinResponseTime = responseTime
				}
				if responseTime > report.MaxResponseTime {
					report.MaxResponseTime = responseTime
				}
				report.TotalResponseTime += responseTime
				mu.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Calculate the total time taken
	totalTime := time.Since(startTime)

	// Calculate the average response time
	averageResponseTime := report.TotalResponseTime / time.Duration(report.TotalRequests)

	// Calculate success rate
	report.SuccessRate = float64(report.Successful) / float64(report.TotalRequests) * 100

	// Calculate throughput
	report.Throughput = float64(report.TotalRequests) / totalTime.Seconds()

	// Write the report to a file with a timestamp
	writeCompleteReportToFile(t, report, totalTime, averageResponseTime, numUsers, numRequests, environment)
}

// writeCompleteReportToFile writes the stress test report to a file with a timestamp
func writeCompleteReportToFile(t *testing.T, report StressTestReport, totalTime time.Duration, averageResponseTime time.Duration, numUsers int, numRequests int, environment string) {
	err := os.MkdirAll("test-reports", os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create test-reports directory: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("test-reports/%s-stress.txt", timestamp)

	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create report file: %v", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "Stress Test Report\n")
	fmt.Fprintf(file, "=================\n")
	fmt.Fprintf(file, "Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Environment: %s\n", environment)
	fmt.Fprintf(file, "Total time taken: %v\n", totalTime)
	fmt.Fprintf(file, "Total requests: %d\n", report.TotalRequests)
	fmt.Fprintf(file, "Requests per user: %d\n", numRequests)
	fmt.Fprintf(file, "Total users: %d\n", numUsers)
	fmt.Fprintf(file, "Successful requests: %d\n", report.Successful)
	fmt.Fprintf(file, "Failed requests: %d\n", report.Failed)
	fmt.Fprintf(file, "Success rate: %.2f%%\n", report.SuccessRate)
	fmt.Fprintf(file, "Throughput: %.2f requests/second\n", report.Throughput)
	fmt.Fprintf(file, "Min response time: %v\n", report.MinResponseTime)
	fmt.Fprintf(file, "Max response time: %v\n", report.MaxResponseTime)
	fmt.Fprintf(file, "Average response time: %v\n", averageResponseTime)

	fmt.Fprintf(file, "\nStatus Code Counts:\n")
	for code, count := range report.StatusCodeCounts {
		fmt.Fprintf(file, "- %d: %d\n", code, count)
	}

	sort.Slice(report.ResponseTimes, func(i, j int) bool { return report.ResponseTimes[i] < report.ResponseTimes[j] })

	fmt.Fprintf(file, "\nPercentile Response Times:\n")

	percentiles := []float64{50.0, 90.0, 95.0, 99.0}
	for _, p := range percentiles {
		index := int(float64(len(report.ResponseTimes)) * p / 100.0)
		if index >= len(report.ResponseTimes) {
			index = len(report.ResponseTimes) - 1
		}
		if index < 0 {
			index = 0
		}
		responseTime := report.ResponseTimes[index]
		fmt.Fprintf(file, "%vth percentile response time: %v\n", p, responseTime)
	}

	interval := time.Second
	fmt.Fprintf(file, "\nRequest and Error Rates Over Time (per second):\n")

	for i := time.Duration(0); i < totalTime; i += interval {
		startTime := time.Now().Add(-totalTime + i)
		endTime := startTime.Add(interval)
		requestCount := 0
		errorCount := 0

		for _, reqTime := range report.RequestTimes {
			if reqTime.After(startTime) && reqTime.Before(endTime) {
				requestCount++
			}
		}

		for _, errTime := range report.ErrorTimes {
			if errTime.After(startTime) && errTime.Before(endTime) {
				errorCount++
			}
		}

		fmt.Fprintf(file, "Time: %v - %v, Requests: %d, Errors: %d\n", startTime.Format("15:04:05"), endTime.Format("15:04:05"), requestCount, errorCount)
	}

	if len(report.Errors) > 0 {
		fmt.Fprintf(file, "\nErrors encountered:\n")
		for _, err := range report.Errors {
			fmt.Fprintf(file, "- %s\n", err)
		}
	}

	t.Logf("Stress test report written to %s", filename)
}
