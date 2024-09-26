package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func checkMatchers(body []byte) []string {
	var detectedTechs []string
	bodyStr := string(body)

	for _, matcher := range matchers {
		allMatched := true

		if matcher.FullMatch {
			for _, pattern := range matcher.Matchers {
				if !strings.Contains(bodyStr, pattern) {
					allMatched = false
					break
				}
			}
		} else {
			for _, pattern := range matcher.Matchers {
				matched, err := regexp.Match(pattern, body)
				if err != nil || !matched {
					allMatched = false
					break
				}
			}
		}

		if allMatched {
			detectedTechs = append(detectedTechs, matcher.Tech)
		}
	}
	return detectedTechs
}

func fetchURL(target string, headers map[string]string, timeout time.Duration, results chan<- Result, wg *sync.WaitGroup, errors *[]string) {
	defer wg.Done()

	if strings.HasSuffix(target, "/") {
		target += "somerandompathwhichmightnotexist"
	} else {
		target += "/somerandompathwhichmightnotexist"
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		*errors = append(*errors, fmt.Sprintf("Error creating request for %s: %v", target, err))
		return
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		*errors = append(*errors, fmt.Sprintf("Error making request to %s: %v", target, err))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		*errors = append(*errors, fmt.Sprintf("Error reading response body from %s: %v", target, err))
		return
	}

	detectedTechs := checkMatchers(body)

	if len(detectedTechs) > 0 {
		results <- Result{Target: target, Stacks: detectedTechs}
	}
}

func handleTarget(target string) []string {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		return []string{"http://" + target, "https://" + target}
	}
	return []string{target}
}

func readTargets(targets string) ([]string, error) {
	if _, err := os.Stat(targets); err == nil {
		data, err := ioutil.ReadFile(targets)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(data), "\n")
		var trimmedLines []string
		for _, line := range lines {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				trimmedLines = append(trimmedLines, trimmed)
			}
		}
		return trimmedLines, nil
	}

	return strings.Split(targets, ","), nil
}

func main() {
	targetsFlag := flag.String("scan", "", "Comma-separated list of targets or file containing targets")
	timeoutFlag := flag.Duration("timeout", DefaultTimeout, "Timeout for HTTP requests")
	headersFlag := flag.String("headers", "", "Comma-separated custom headers (e.g., 'User-Agent:Go,Accept:*/*')")
	workersFlag := flag.Int("workers", 5, "Number of concurrent workers")

	flag.Parse()

	fmt.Print(banner)

	if *targetsFlag == "" {
		fmt.Println("\nError: Provide at least one target using the -scan flag.")
		os.Exit(1)
	}

	headers := make(map[string]string)
	if *headersFlag != "" {
		for _, h := range strings.Split(*headersFlag, ",") {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	targets, err := readTargets(*targetsFlag)
	if err != nil {
		fmt.Println("Error reading targets:", err)
		os.Exit(1)
	}

	results := make(chan Result, len(targets))
	var wg sync.WaitGroup
	var errors []string

	semaphore := make(chan struct{}, *workersFlag)

	for _, target := range targets {
		for _, fullTarget := range handleTarget(strings.TrimSpace(target)) {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(t string) {
				defer func() { <-semaphore }()
				fetchURL(t, headers, *timeoutFlag, results, &wg, &errors)
			}(fullTarget)
		}
	}

	wg.Wait()
	close(results)

	var finalResults []Result
	for res := range results {
		finalResults = append(finalResults, res)
	}

	output := Output{
		Results: finalResults,
		Errors:  errors,
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling final results:", err)
		os.Exit(1)
	}

	fmt.Println("\n" + string(jsonOutput))
}
