package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const availableProviders = "Available: github, gitlab and none."

func main() {
	provider := flag.String("provider", "github", "The name of provider. "+availableProviders)
	url := flag.String("url", "", "The URL of the schema like this: first-debug/lk-graphql-schemas/master/schemas/user-provider/schema.graphql")
	output := flag.String("output", "", "The path to save the schema file.")
	timeout := flag.Duration("timeout", 30*time.Second, "The timeout for the HTTP request.")
	flag.Parse()

	if *url == "" {
		exitWithErr("The -url flag is required.")
	}
	if *output == "" {
		exitWithErr("The -output flag is required.")
	}

	if err := os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		exitWithErr(fmt.Sprintf("failed to create output directory: %v", err))
	}

	if info, _ := os.Stat(*output); info != nil && info.IsDir() {
		parts := strings.Split(*url, "/")
		*output += parts[len(parts)-1]
	}

	file, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		exitWithErr(fmt.Sprintf("failed to open output file: %v", err))
	}
	defer file.Close()

	client := &http.Client{
		Timeout: *timeout,
	}

	switch *provider {
	case "github":
		*url = "https://raw.githubusercontent.com/" + *url
	case "gitlab":
		*url = "https://gitlab.com/" + *url
	case "none":
	default:
		exitWithErr("Not available provider. " + availableProviders)
	}

	resp, err := client.Get(*url)
	if err != nil {
		exitWithErr(fmt.Sprintf("failed to fetch schema: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		exitWithErr(fmt.Sprintf("bad status: %s", resp.Status))
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		exitWithErr(fmt.Sprintf("failed to write schema to file: %v", err))
	}

	fmt.Printf("Schema successfully fetched from %s and saved to %s\n", *url, *output)
}

func exitWithErr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
