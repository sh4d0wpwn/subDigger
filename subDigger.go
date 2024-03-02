package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type SubdomainResult struct {
	Subdomains []string
	Mutex      sync.Mutex
}

func (sr *SubdomainResult) Add(subdomains []string) {
	sr.Mutex.Lock()
	defer sr.Mutex.Unlock()
	for _, subdomain := range subdomains {
		if !contains(sr.Subdomains, subdomain) {
			sr.Subdomains = append(sr.Subdomains, subdomain)
		}
	}
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, exists := set[item]
	return exists
}

func executeExternalTool(domain, toolName string, args []string, results *SubdomainResult) {
	cmd := exec.Command(toolName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing %s: %s\n", toolName, err)
		return
	}
	// Correct use of the provided parseSubdomains function
	subdomains := parseSubdomains(string(output), domain)
	results.Add(subdomains)
}

// Existing parseSubdomains function as provided

func parseSubdomains(rawOutput, domain string) []string {
	regexPattern := fmt.Sprintf(`([a-zA-Z0-9_-]+\.)+%s`, regexp.QuoteMeta(domain))
	subdomainRegex := regexp.MustCompile(regexPattern)

	matches := subdomainRegex.FindAllString(rawOutput, -1)

	// Deduplicate matches
	uniqueMatches := make(map[string]bool)
	for _, match := range matches {
		uniqueMatches[match] = true
	}

	var subdomains []string
	for subdomain := range uniqueMatches {
		subdomains = append(subdomains, subdomain)
	}

	return subdomains
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func fetchJLDC(domain string, results *SubdomainResult) {
	url := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", domain)
	res, err := httpGet(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching JLDC data: %s\n", err)
		return
	}
	var subdomains []string
	if err := json.Unmarshal(res, &subdomains); err == nil {
		results.Add(subdomains)
	}
}

func fetchCRTSH(domain string, results *SubdomainResult) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
	res, err := httpGet(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching crt.sh data: %s\n", err)
		return
	}
	var data []struct {
		NameValue string `json:"name_value"`
	}
	if err := json.Unmarshal(res, &data); err == nil {
		var subdomains []string
		for _, d := range data {
			subdomains = append(subdomains, strings.Split(d.NameValue, "\n")...)
		}
		results.Add(subdomains)
	}
}

func fetchCertSpotter(domain string, results *SubdomainResult) {
	url := fmt.Sprintf("https://api.certspotter.com/v1/issuances?domain=%s&include_subdomains=true&expand=dns_names", domain)
	res, err := httpGet(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching CertSpotter data: %s\n", err)
		return
	}
	var data []struct {
		DNSNames []string `json:"dns_names"`
	}
	if err := json.Unmarshal(res, &data); err == nil {
		for _, d := range data {
			results.Add(d.DNSNames)
		}
	}
}

func unique(strings []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, s := range strings {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./subDigger <domain>")
		os.Exit(1)
	}
	domain := os.Args[1]
	results := &SubdomainResult{Subdomains: []string{}}

	var wg sync.WaitGroup

	// Add external tools execution with goroutines
	tools := []struct {
		name string
		args []string
	}{
		{"sublist3r", []string{"-d", domain}},
		{"subfinder", []string{"-d", domain}},
		{"assetfinder", []string{"-d", domain}},
		{"findomain", []string{"-t", domain}},
	}

	for _, tool := range tools {
		wg.Add(1)
		go func(tool struct {
			name string
			args []string
		}) {
			defer wg.Done()
			executeExternalTool(domain, tool.name, tool.args, results)
		}(tool)
	}

	// API fetches with goroutines
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchJLDC(domain, results)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchCRTSH(domain, results)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetchCertSpotter(domain, results)
	}()

	wg.Wait()

	// Deduplicate and sort results
	uniqueSubdomains := unique(results.Subdomains)
	sort.Strings(uniqueSubdomains)
	for _, subdomain := range uniqueSubdomains {
		fmt.Println(subdomain)
	}
}
