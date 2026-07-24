package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	antigravityadapter "github.com/wildbyteai/planlens/internal/adapters/antigravity"
	claudeadapter "github.com/wildbyteai/planlens/internal/adapters/claude"
	"github.com/wildbyteai/planlens/internal/review"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "review" {
		fmt.Fprintln(stderr, "usage: planlens review --plan <path> --reviewer <simulated|claude|antigravity>")
		return 2
	}

	flags := flag.NewFlagSet("review", flag.ContinueOnError)
	flags.SetOutput(stderr)
	planPath := flags.String("plan", "", "path to the plan to review")
	reviewerID := flags.String("reviewer", "", "reviewer adapter to invoke")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *planPath == "" || *reviewerID == "" {
		fmt.Fprintln(stderr, "plan and reviewer are required")
		return 2
	}
	plan, err := os.ReadFile(*planPath)
	if err != nil {
		fmt.Fprintf(stderr, "read plan: %v\n", err)
		return 1
	}

	var reviewer review.Reviewer
	switch *reviewerID {
	case review.SimulatedReviewerID:
		reviewer, err = review.NewSimulatedReviewer()
	case claudeadapter.ID:
		reviewer, err = claudeadapter.New()
	case antigravityadapter.ReviewerID:
		reviewer, err = antigravityadapter.Discover(ctx)
	default:
		fmt.Fprintf(stderr, "unsupported reviewer %q\n", *reviewerID)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "prepare reviewer: %v\n", err)
		return 1
	}
	result, err := reviewer.Run(ctx, review.Request{Plan: plan})
	if err != nil {
		fmt.Fprintf(stderr, "run reviewer: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "PlanLens review result")
	fmt.Fprintf(stdout, "Reviewer: %s\n", reviewer.ID())
	if result.Metadata.CLIVersion != "" {
		fmt.Fprintf(stdout, "CLI version: %s\n", result.Metadata.CLIVersion)
	}
	if result.Metadata.AccessCapability != "" {
		fmt.Fprintf(stdout, "Access capability: %s\n", result.Metadata.AccessCapability)
	}
	fmt.Fprintln(stdout, "Status: complete")
	for _, finding := range result.Findings {
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "%s: %s\n", strings.ToUpper(string(finding.Severity)), finding.Title)
		fmt.Fprintf(stdout, "Evidence: %s\n", finding.Evidence)
		fmt.Fprintf(stdout, "Impact: %s\n", finding.Impact)
		fmt.Fprintf(stdout, "Suggested action: %s\n", finding.SuggestedAction)
	}
	return 0
}
