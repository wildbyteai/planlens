package qualification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/codex"
	"github.com/wildbyteai/planlens/internal/review"
)

func TestRunUsesOnlyFrozenPublicFixtureAndEmitsSanitizedMetadata(t *testing.T) {
	root := repositoryRoot(t)
	wantPlan, err := os.ReadFile(filepath.Join(root, publicFixturePath))
	if err != nil {
		t.Fatal(err)
	}
	reviewer := &capturingReviewer{result: review.Result{
		SchemaVersion: "1",
		Findings: []review.Finding{{
			Severity:        review.SeverityMajor,
			Title:           "private model wording",
			Evidence:        "private model evidence",
			Impact:          "private model impact",
			SuggestedAction: "private model action",
		}},
	}}
	discover := func(context.Context) (candidate, error) {
		return candidate{
			reviewer:    reviewer,
			cliVersion:  "0.146.0-alpha.3.1",
			accessLevel: string(codex.AccessConstrained),
		}, nil
	}

	record := run(context.Background(), root, qualificationDate(), discover)
	if got, want := record.Result, ResultPassed; got != want {
		t.Fatalf("qualification result = %q, want %q", got, want)
	}
	if got, want := record.CLIVersion, "0.146.0-alpha.3.1"; got != want {
		t.Fatalf("CLI version = %q, want %q", got, want)
	}
	if got, want := record.Platform, runtimePlatform(); got != want {
		t.Fatalf("platform = %q, want %q", got, want)
	}
	if got, want := record.QualificationDate, "2026-07-24"; got != want {
		t.Fatalf("qualification date = %q, want %q", got, want)
	}
	if got, want := string(reviewer.request.Plan), string(wantPlan); got != want {
		t.Fatalf("reviewer did not receive exact frozen public fixture\nwant:\n%s\ngot:\n%s", want, got)
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		root,
		"Public Parser Rollout Plan",
		"private model wording",
		"private model evidence",
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("sanitized record retained forbidden material %q: %s", forbidden, encoded)
		}
	}
}

func TestRunRejectsChangedFixtureBeforeStartingReviewer(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "testdata"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, publicFixturePath), []byte("not the frozen fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	discover := func(context.Context) (candidate, error) {
		called = true
		return candidate{}, nil
	}

	record := run(context.Background(), root, qualificationDate(), discover)
	if called {
		t.Fatal("qualification discovered Codex after fixed fixture validation failed")
	}
	if got, want := record.Result, ResultBlocked; got != want {
		t.Fatalf("qualification result = %q, want %q", got, want)
	}
	if !slicesContain(record.UnresolvedLimitations, "qualification blocked: fixture_changed") {
		t.Fatalf("limitations do not contain sanitized fixture failure: %#v", record.UnresolvedLimitations)
	}
}

func TestRunClassifiesFailureWithoutRetainingRawDiagnostic(t *testing.T) {
	discover := func(context.Context) (candidate, error) {
		return candidate{
			cliVersion:  "0.146.0-alpha.3.1",
			accessLevel: string(codex.AccessConstrained),
		}, fmt.Errorf("%w: owner@example.com secret-token", codex.ErrNotAuthenticated)
	}

	record := run(context.Background(), repositoryRoot(t), qualificationDate(), discover)
	if got, want := record.Result, ResultBlocked; got != want {
		t.Fatalf("qualification result = %q, want %q", got, want)
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"owner@example.com", "secret-token"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("sanitized record retained raw diagnostic %q: %s", forbidden, encoded)
		}
	}
	if !slicesContain(record.UnresolvedLimitations, "qualification blocked: not_authenticated") {
		t.Fatalf("limitations do not contain sanitized authentication failure: %#v", record.UnresolvedLimitations)
	}
}

func TestRunClassifiesGlobalInstructionsWithoutReadingThem(t *testing.T) {
	discover := func(context.Context) (candidate, error) {
		return candidate{
			cliVersion:  "0.146.0-alpha.3.1",
			accessLevel: string(codex.AccessConstrained),
		}, codex.ErrGlobalInstructionsPresent
	}

	record := run(context.Background(), repositoryRoot(t), qualificationDate(), discover)
	if got, want := record.Result, ResultBlocked; got != want {
		t.Fatalf("qualification result = %q, want %q", got, want)
	}
	if !slicesContain(record.UnresolvedLimitations, "qualification blocked: global_instructions_present") {
		t.Fatalf("limitations do not contain sanitized global-instruction failure: %#v", record.UnresolvedLimitations)
	}
}

type capturingReviewer struct {
	request review.Request
	result  review.Result
	err     error
}

func (reviewer *capturingReviewer) ID() string {
	return codex.ReviewerID
}

func (reviewer *capturingReviewer) Run(_ context.Context, request review.Request) (review.Result, error) {
	reviewer.request = request
	return reviewer.result, reviewer.err
}

func qualificationDate() time.Time {
	return time.Date(2026, time.July, 24, 12, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(workingDirectory, "..", "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("locate repository root: %v", err)
	}
	return root
}

func slicesContain(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
