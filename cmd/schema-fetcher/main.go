package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	url := flag.String("url", "", "The URL of the GraphQL schema to fetch.")
	output := flag.String("output", "api/graphql/schema.graphql", "The path to save the schema file.")
	timeout := flag.Duration("timeout", 30*time.Second, "The timeout for the HTTP request.")
	flag.Parse()

	if *url == "" {
		exitWithErr("The --url flag is required.")
	}

	if err := os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		exitWithErr(fmt.Sprintf("failed to create output directory: %v", err))
	}

	file, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		exitWithErr(fmt.Sprintf("failed to open output file: %v", err))
	}
	defer file.Close()

	client := &http.Client{
		Timeout: *timeout,
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
