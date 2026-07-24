package qualification

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/codex"
	"github.com/wildbyteai/planlens/internal/review"
)

const (
	publicFixturePath   = "testdata/public-plan.md"
	publicFixtureDigest = "7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429"

	ResultPassed  = "passed"
	ResultBlocked = "blocked"
)

type Record struct {
	SchemaVersion         string   `json:"schema_version"`
	Adapter               string   `json:"adapter"`
	QualificationDate     string   `json:"qualification_date"`
	CLIVersion            string   `json:"cli_version"`
	Platform              string   `json:"platform"`
	Result                string   `json:"result"`
	AccessCapability      string   `json:"access_capability"`
	UnresolvedLimitations []string `json:"unresolved_limitations"`
	OfficialDocumentation []string `json:"official_documentation"`
}

type candidate struct {
	reviewer    review.Reviewer
	cliVersion  string
	accessLevel string
}

type discoverFunc func(context.Context) (candidate, error)

func Run(ctx context.Context, startDirectory string, now time.Time) Record {
	repositoryRoot, err := findRepositoryRoot(startDirectory)
	if err != nil {
		return addBlock(baseRecord(now), "repository_not_found")
	}
	return run(ctx, repositoryRoot, now, discoverCodex)
}

func run(ctx context.Context, repositoryRoot string, now time.Time, discover discoverFunc) Record {
	record := baseRecord(now)
	fixture, err := os.ReadFile(filepath.Join(repositoryRoot, publicFixturePath))
	if err != nil {
		return addBlock(record, "fixture_unavailable")
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(fixture))
	if digest != publicFixtureDigest {
		return addBlock(record, "fixture_changed")
	}

	discovered, err := discover(ctx)
	if discovered.cliVersion != "" {
		record.CLIVersion = discovered.cliVersion
	}
	if discovered.accessLevel != "" {
		record.AccessCapability = discovered.accessLevel
	}
	if err != nil {
		return addBlock(record, classifyFailure(err))
	}
	if discovered.reviewer == nil {
		return addBlock(record, "adapter_unavailable")
	}
	if _, err := discovered.reviewer.Run(ctx, review.Request{Plan: fixture}); err != nil {
		return addBlock(record, classifyFailure(err))
	}
	record.Result = ResultPassed
	return record
}

func discoverCodex(ctx context.Context) (candidate, error) {
	discovery, err := codex.Discover(ctx)
	discovered := candidate{
		cliVersion:  discovery.CLIVersion,
		accessLevel: string(discovery.AccessCapability),
	}
	if discovery.Adapter != nil {
		discovered.reviewer = discovery.Adapter
	}
	return discovered, err
}

func baseRecord(now time.Time) Record {
	return Record{
		SchemaVersion:     "1",
		Adapter:           codex.ReviewerID,
		QualificationDate: now.UTC().Format("2006-01-02"),
		CLIVersion:        "unavailable",
		Platform:          runtimePlatform(),
		Result:            ResultBlocked,
		AccessCapability:  string(codex.AccessConstrained),
		UnresolvedLimitations: []string{
			"Codex read-only mode constrains writes but is not a material-only operating-system sandbox.",
			"Authentication is checked by exit status only; account identity and raw diagnostics are not retained.",
			"This result covers one CLI version, platform, date, and fixed public fixture; it does not establish universal compatibility.",
			"Codex CLI controls may change in later releases, so each supported version still requires requalification.",
		},
		OfficialDocumentation: []string{
			"https://developers.openai.com/codex/non-interactive-mode",
			"https://developers.openai.com/codex/auth",
			"https://developers.openai.com/codex/cli/reference",
			"https://developers.openai.com/codex/config-reference",
			"https://developers.openai.com/codex/guides/agents-md",
		},
	}
}

func addBlock(record Record, failureClass string) Record {
	record.Result = ResultBlocked
	record.UnresolvedLimitations = append(record.UnresolvedLimitations, "qualification blocked: "+failureClass)
	return record
}

func classifyFailure(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, codex.ErrCLINotFound):
		return "cli_not_found"
	case errors.Is(err, codex.ErrVersionProbeFailed):
		return "version_probe_failed"
	case errors.Is(err, codex.ErrUnsupportedVersion):
		return "unsupported_version"
	case errors.Is(err, codex.ErrGlobalInstructionsPresent):
		return "global_instructions_present"
	case errors.Is(err, codex.ErrUnsupportedCapabilities):
		return "required_controls_unavailable"
	case errors.Is(err, codex.ErrNotAuthenticated):
		return "not_authenticated"
	case errors.Is(err, codex.ErrCleanupFailed):
		return "workspace_cleanup_failed"
	case errors.Is(err, codex.ErrExecutionFailed):
		return "review_process_failed"
	case errors.Is(err, codex.ErrInvalidFinalResponse):
		return "invalid_final_response"
	default:
		return "unknown_failure"
	}
}

func runtimePlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

func findRepositoryRoot(startDirectory string) (string, error) {
	current, err := filepath.Abs(startDirectory)
	if err != nil {
		return "", errors.New("resolve repository root")
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("repository root not found")
		}
		current = parent
	}
}
