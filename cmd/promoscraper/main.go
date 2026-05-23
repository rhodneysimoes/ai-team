package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rhodneysimoes/ai-team/internal/scraper"
	"github.com/rhodneysimoes/ai-team/internal/sites"
)

func main() {
	var (
		sitesPath   = flag.String("sites", "sites.md", "path to markdown sites file")
		outputPath  = flag.String("output", "", "optional JSON output file")
		timeout     = flag.Duration("timeout", 15*time.Second, "request timeout")
		concurrency = flag.Int("concurrency", 4, "maximum concurrent requests")
	)
	flag.Parse()

	if err := run(*sitesPath, *outputPath, *timeout, *concurrency); err != nil {
		log.Fatal(err)
	}
}

func run(sitesPath, outputPath string, timeout time.Duration, concurrency int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	definitions, err := sites.LoadMarkdown(sitesPath)
	if err != nil {
		return fmt.Errorf("load sites: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	results, err := scraper.Collect(ctx, client, definitions, scraper.Options{
		Concurrency: concurrency,
	})
	if err != nil {
		return fmt.Errorf("collect promotions: %w", err)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("encode results: %w", err)
	}
	data = append(data, '\n')

	if outputPath == "" {
		_, err = os.Stdout.Write(data)
		if err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}

	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return fmt.Errorf("write output %q: %w", outputPath, err)
	}
	return nil
}
