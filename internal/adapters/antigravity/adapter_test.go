package antigravity_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/antigravity"
	"github.com/wildbyteai/planlens/internal/review"
)

func TestDiscoverAndReviewWithOfficialCLIContract(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PLANLENS_TEST_SECRET", "must-not-reach-reviewer")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adapter, err := antigravity.Discover(ctx)
	if err != nil {
		t.Fatalf("discover Antigravity CLI: %v", err)
	}
	if adapter.ID() != antigravity.ReviewerID {
		t.Fatalf("unexpected reviewer identity: %q", adapter.ID())
	}
	if adapter.Version() != "1.1.6" {
		t.Fatalf("unexpected CLI version: %q", adapter.Version())
	}

	result, err := adapter.Review(ctx, antigravity.ReviewRequest{
		Plan:    []byte(publicPlanFixture),
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("review public fixture: %v", err)
	}
	if result.Authentication != antigravity.AuthenticationAvailable {
		t.Fatalf("unexpected authentication status: %q", result.Authentication)
	}
	if result.AccessCapability != antigravity.AccessConstrained {
		t.Fatalf("unexpected access capability: %q", result.AccessCapability)
	}
	if result.FinalResponse != expectedFinalResponse {
		t.Fatalf("unexpected final response\nwant:\n%s\ngot:\n%s", expectedFinalResponse, result.FinalResponse)
	}
	if strings.Contains(result.FinalResponse, "PRIVATE_REASONING") {
		t.Fatal("intermediate reasoning reached the final response")
	}
}

func TestDiscoverRejectsUnqualifiedVersion(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-version"), []byte("1.2.0"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := antigravity.Discover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("expected unsupported-version error, got %v", err)
	}
}

func TestDiscoverRejectsMissingNativeControl(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-missing-control"), []byte("--sandbox"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := antigravity.Discover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "missing --sandbox") {
		t.Fatalf("expected missing-control error, got %v", err)
	}
}

func TestDiscoverRejectsMissingCompleteModeOption(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-missing-control"), []byte("--mode"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := antigravity.Discover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "missing --mode") {
		t.Fatalf("expected missing-mode error despite --model lookalike, got %v", err)
	}
}

func TestDiscoverRejectsMissingCompletePrintOption(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-missing-control"), []byte("--print"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := antigravity.Discover(context.Background())
	if err == nil || !strings.Contains(err.Error(), "missing --print") {
		t.Fatalf("expected missing-print error despite --print-timeout lookalike, got %v", err)
	}
}

func TestDiscoverAcceptsNativeControlsReportedOnStderr(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-help-stderr"), []byte("yes"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := antigravity.Discover(context.Background()); err != nil {
		t.Fatalf("discover controls reported on stderr: %v", err)
	}
}

func TestDiscoverPreservesCanceledContextFromVersionProbe(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := antigravity.Discover(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled discovery, got %T: %v", err, err)
	}
	if strings.Contains(err.Error(), "command failed") {
		t.Fatalf("canceled version probe was misclassified: %v", err)
	}
}

func TestDiscoverPreservesDeadlineFromVersionProbe(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := antigravity.Discover(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline-exceeded discovery, got %T: %v", err, err)
	}
	if strings.Contains(err.Error(), "command failed") {
		t.Fatalf("expired version probe was misclassified: %v", err)
	}
}

func TestDiscoverPreservesCanceledContextFromCapabilityProbe(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-block-help"), []byte("yes"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := antigravity.Discover(ctx)
		result <- err
	}()
	waitForFile(t, filepath.Join(home, "fake-agy-help-started"))
	cancel()

	err := <-result
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled discovery, got %T: %v", err, err)
	}
	if strings.Contains(err.Error(), "command failed") {
		t.Fatalf("canceled capability probe was misclassified: %v", err)
	}
}

func TestDiscoverPreservesDeadlineFromCapabilityProbe(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-block-help"), []byte("yes"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := antigravity.Discover(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline-exceeded discovery, got %T: %v", err, err)
	}
	if strings.Contains(err.Error(), "command failed") {
		t.Fatalf("expired capability probe was misclassified: %v", err)
	}
}

func TestRunNormalizesFinalResponseIntoSharedProtocol(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	adapter, err := antigravity.Discover(ctx)
	if err != nil {
		t.Fatalf("discover Antigravity CLI: %v", err)
	}
	result, err := adapter.Run(ctx, review.Request{Plan: []byte(publicPlanFixture)})
	if err != nil {
		t.Fatalf("run Antigravity review: %v", err)
	}
	if got, want := result.SchemaVersion, "1"; got != want {
		t.Fatalf("schema version = %q, want %q", got, want)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %#v, want one", result.Findings)
	}
	finding := result.Findings[0]
	if finding.Severity != review.SeverityMajor || finding.Title != "The rollout has no rollback decision." {
		t.Fatalf("unexpected normalized finding: %#v", finding)
	}
	if got, want := result.Metadata.CLIVersion, "1.1.6"; got != want {
		t.Fatalf("CLI version = %q, want %q", got, want)
	}
	if got, want := result.Metadata.AccessCapability, review.AccessConstrained; got != want {
		t.Fatalf("access capability = %q, want %q", got, want)
	}
}

func TestReviewClassifiesMissingAuthenticationWithoutReturningCLIOutput(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "fake-agy-auth-required"), []byte("yes"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adapter, err := antigravity.Discover(ctx)
	if err != nil {
		t.Fatalf("discover Antigravity CLI: %v", err)
	}
	_, err = adapter.Review(ctx, antigravity.ReviewRequest{
		Plan:    []byte(publicPlanFixture),
		Timeout: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected authentication failure")
	}
	var runError *antigravity.RunError
	if !errors.As(err, &runError) {
		t.Fatalf("expected typed run error, got %T: %v", err, err)
	}
	if runError.Kind != antigravity.FailureAuthRequired {
		t.Fatalf("unexpected failure kind: %q", runError.Kind)
	}
	if strings.Contains(err.Error(), "PRIVATE_AUTH_DETAIL") {
		t.Fatal("CLI authentication output leaked through the public error")
	}
}

func TestReviewClassifiesCanceledContext(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))

	adapter, err := antigravity.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover Antigravity CLI: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = adapter.Review(ctx, antigravity.ReviewRequest{
		Plan:    []byte(publicPlanFixture),
		Timeout: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected canceled review")
	}
	var runError *antigravity.RunError
	if !errors.As(err, &runError) || runError.Kind != antigravity.FailureCanceled {
		t.Fatalf("expected canceled failure kind, got %T: %v", err, err)
	}
}

func TestReviewRejectsChangedMaterialBeforeStartingCLI(t *testing.T) {
	fakeDirectory := buildFakeAGY(t)
	t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	adapter, err := antigravity.Discover(ctx)
	if err != nil {
		t.Fatalf("discover Antigravity CLI: %v", err)
	}
	_, err = adapter.Review(ctx, antigravity.ReviewRequest{Plan: []byte("changed or sensitive material")})
	if err == nil || !strings.Contains(err.Error(), "public fixture") {
		t.Fatalf("expected public-fixture boundary error, got %v", err)
	}
}

func buildFakeAGY(t *testing.T) string {
	t.Helper()

	repositoryRoot := repositoryRoot(t)
	fakeSource := filepath.Join(repositoryRoot, "internal", "adapters", "antigravity", "testdata", "fakeagy", "main.go")
	if _, err := os.ReadFile(fakeSource); err != nil {
		t.Fatalf("read fake agy source: %v", err)
	}
	outputDirectory := t.TempDir()
	name := "agy"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	outputPath := filepath.Join(outputDirectory, name)
	goCommand := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goCommand += ".exe"
	}
	command := exec.Command(goCommand, "build", "-o", outputPath, "./internal/adapters/antigravity/testdata/fakeagy")
	command.Dir = repositoryRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build fake agy: %v\n%s", err, output)
	}
	return outputDirectory
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(workingDirectory, "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("locate repository root: %v", err)
	}
	return root
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", filepath.Base(path))
}

const publicPlanFixture = `# Public Parser Rollout Plan

## Objective

Replace the existing document parser without interrupting document imports.

## Decision

Deploy the new parser to all users in one release window.

## Plan

1. Deploy the new parser.
2. Observe import errors for 30 minutes.
3. Mark the rollout complete.

## Constraints

- Existing documents must remain readable.
- The rollout must finish within one maintenance window.

## Acceptance criteria

- New documents import successfully.
- Existing documents still open successfully.

<!-- planlens-public-fixture-end -->
`

const expectedFinalResponse = `PLANLENS_ANTIGRAVITY_FEASIBILITY_OK
Finding: The rollout has no rollback decision.
Severity: major
Evidence: The plan deploys to all users and defines no rollback trigger or owner.
Impact: A failed deployment could remain active without clear responsibility.
Suggested action: Add a rollback trigger, owner, and recovery verification.
Fixture marker: planlens-public-fixture-end`
