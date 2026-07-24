package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/codex/qualification"
)

func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "qualification could not determine its working directory")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	record := qualification.Run(ctx, workingDirectory, time.Now())

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(record); err != nil {
		fmt.Fprintln(os.Stderr, "qualification could not encode its sanitized record")
		os.Exit(1)
	}
	if record.Result != qualification.ResultPassed {
		os.Exit(1)
	}
}
