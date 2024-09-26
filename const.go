package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type TechMatcher struct {
	Tech      string   `json:"Tech"`
	Matchers  []string `json:"Matchers"`
	FullMatch bool     `json:"FullMatch"`
}

type Result struct {
	Target string   `json:"target"`
	Stacks []string `json:"techs"`
}

type Output struct {
	Results []Result `json:"results"`
	Errors  []string `json:"errors,omitempty"`
}

var (
	matchers []TechMatcher

	DefaultTimeout = 10 * time.Second
)

func init() {
	tmpDir := "/tmp/"
	matchersFile := tmpDir + "matchers.json"

	// Check if /tmp/ directory exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		fmt.Println("Error: /tmp/ directory does not exist. Falling back to default matchers.json.")
		readDefaultMatchers()
		return
	}

	// Check if matchers.json exists in /tmp/
	if _, err := os.Stat(matchersFile); os.IsNotExist(err) {
		downloadMatchers(matchersFile)
	}

	// Read matchers.json from /tmp/
	file, err := ioutil.ReadFile(matchersFile)
	if err != nil {
		fmt.Println("Error reading matchers.json from /tmp/:", err)
		readDefaultMatchers()
		return
	}

	if err := json.Unmarshal(file, &matchers); err != nil {
		fmt.Println("Error unmarshaling matchers:", err)
		readDefaultMatchers()
	}
}

func readDefaultMatchers() {
	file, err := ioutil.ReadFile("matchers.json")
	if err != nil {
		fmt.Println("Error reading default matchers.json:", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(file, &matchers); err != nil {
		fmt.Println("Error unmarshaling default matchers:", err)
		os.Exit(1)
	}
}

func downloadMatchers(dest string) {
	url := "https://raw.githubusercontent.com/umair9747/4oFour/refs/heads/main/matchers.json"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading matchers.json:", err)
		readDefaultMatchers()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: received non-200 response code:", resp.StatusCode)
		readDefaultMatchers()
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		readDefaultMatchers()
		return
	}

	if err := ioutil.WriteFile(dest, data, 0644); err != nil {
		fmt.Println("Error writing matchers.json to /tmp/:", err)
		readDefaultMatchers()
		return
	}
}

const banner = `
	     4     |         |         
	-----------+---------+-----------
	           |    O    |         
	-----------+---------+-----------
	           |         |    four   

The tech enumeration toolkit for 404 Not found pages.
				- Developed by 0x9747 (inspired by 0xdf)
`
