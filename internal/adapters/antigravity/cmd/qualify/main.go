package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/antigravity"
)

func main() {
	os.Exit(run())
}

func run() int {
	flags := flag.NewFlagSet("qualify-antigravity", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	fixturePath := flags.String("fixture", "", "public plan fixture to review")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *fixturePath == "" {
		fmt.Fprintln(os.Stderr, "fixture is required")
		return 2
	}

	fixture, err := os.ReadFile(*fixturePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read public fixture: failed")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	adapter, err := antigravity.Discover(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover Antigravity CLI: %v\n", err)
		return 1
	}
	result, err := adapter.Review(ctx, antigravity.ReviewRequest{
		Plan:    fixture,
		Timeout: 3 * time.Minute,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "qualify Antigravity CLI: %v\n", err)
		return 1
	}
	record, err := antigravity.BuildFeasibilityRecord(
		time.Now(),
		runtime.GOOS,
		runtime.GOARCH,
		fixture,
		result,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build feasibility record: %v\n", err)
		return 1
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(record); err != nil {
		fmt.Fprintln(os.Stderr, "write feasibility record: failed")
		return 1
	}
	return 0
}
