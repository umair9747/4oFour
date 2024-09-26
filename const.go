package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Stacks []string `json:"stacks"`
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
	file, err := ioutil.ReadFile("matchers.json")
	if err != nil {
		fmt.Println("Error reading matchers.json:", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(file, &matchers); err != nil {
		fmt.Println("Error unmarshaling matchers:", err)
		os.Exit(1)
	}
}

const banner = `
     4     |         |         
-----------+---------+-----------
           |    O    |         
-----------+---------+-----------
           |         |    four   

The 404 Page not found, tech enumeration toolkit
    	- Developed by 0x9747 (inspired by 0xdf)
`
