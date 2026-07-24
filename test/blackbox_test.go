package blackbox_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestUserCanReviewPublicPlanWithExternalReviewer(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	binaryDirectory := t.TempDir()
	planlens := buildCommand(t, repositoryRoot, binaryDirectory, "planlens", "./cmd/planlens")
	buildCommand(t, repositoryRoot, binaryDirectory, "planlens-simulated-reviewer", "./test/fixtures/mockreviewer")

	want := `PlanLens review result
Reviewer: simulated
Status: complete

MAJOR: The rollout has no rollback decision
Evidence: The plan schedules deployment but defines no rollback trigger or owner.
Impact: A failed deployment could remain active while responsibility is unclear.
Suggested action: Add a rollback trigger, named owner, and recovery verification step.
`
	first := runReview(t, repositoryRoot, planlens)
	if first != want {
		t.Fatalf("unexpected review output\nwant:\n%s\ngot:\n%s", want, first)
	}

	second := runReview(t, repositoryRoot, planlens)
	if second != first {
		t.Fatalf("review output changed between identical invocations\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestUserCanReviewPublicPlanWithClaudeCodeCLI(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	binaryDirectory := t.TempDir()
	planlens := buildCommand(t, repositoryRoot, binaryDirectory, "planlens", "./cmd/planlens")
	buildCommand(t, repositoryRoot, binaryDirectory, "claude", "./test/fixtures/fakeclaude")

	want := `PlanLens review result
Reviewer: claude
CLI version: 2.1.218
Access capability: constrained
Status: complete

MAJOR: The rollout has no rollback decision
Evidence: The plan schedules deployment but defines no rollback trigger or owner.
Impact: A failed deployment could remain active while responsibility is unclear.
Suggested action: Add a rollback trigger, named owner, and recovery verification step.
`
	got := runReviewWithReviewer(t, repositoryRoot, planlens, "claude", binaryDirectory)
	if got != want {
		t.Fatalf("unexpected Claude Code review output\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUserCanReviewPublicPlanWithAntigravityCLI(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	binaryDirectory := t.TempDir()
	planlens := buildCommand(t, repositoryRoot, binaryDirectory, "planlens", "./cmd/planlens")
	buildCommand(t, repositoryRoot, binaryDirectory, "agy", "./internal/adapters/antigravity/testdata/fakeagy")

	want := `PlanLens review result
Reviewer: antigravity
CLI version: 1.1.6
Access capability: constrained
Status: complete

MAJOR: The rollout has no rollback decision.
Evidence: The plan deploys to all users and defines no rollback trigger or owner.
Impact: A failed deployment could remain active without clear responsibility.
Suggested action: Add a rollback trigger, owner, and recovery verification.
`
	got := runReviewWithReviewer(t, repositoryRoot, planlens, "antigravity", binaryDirectory)
	if got != want {
		t.Fatalf("unexpected Antigravity review output\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestUserCanReviewPublicPlanWithCodexCLI(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	binaryDirectory := t.TempDir()
	planlens := buildCommand(t, repositoryRoot, binaryDirectory, "planlens", "./cmd/planlens")
	buildCommand(t, repositoryRoot, binaryDirectory, "codex", "./internal/adapters/codex/testdata/fakecodex")

	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "fake-codex-authenticated"), []byte("yes"), 0o600); err != nil {
		t.Fatal(err)
	}
	finalResponse := `{"schema_version":"1","findings":[{"severity":"major","title":"The rollout has no rollback decision","evidence":"The plan schedules deployment but defines no rollback trigger or owner.","impact":"A failed deployment could remain active while responsibility is unclear.","suggested_action":"Add a rollback trigger, named owner, and recovery verification step."}]}`
	if err := os.WriteFile(filepath.Join(home, "fake-codex-final.json"), []byte(finalResponse), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))

	want := `PlanLens review result
Reviewer: codex
CLI version: 0.146.0-alpha.3.1
Access capability: constrained
Status: complete

MAJOR: The rollout has no rollback decision
Evidence: The plan schedules deployment but defines no rollback trigger or owner.
Impact: A failed deployment could remain active while responsibility is unclear.
Suggested action: Add a rollback trigger, named owner, and recovery verification step.
`
	got := runReviewWithReviewer(t, repositoryRoot, planlens, "codex", binaryDirectory)
	if got != want {
		t.Fatalf("unexpected Codex review output\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func runReview(t *testing.T, repositoryRoot, planlens string) string {
	return runReviewWithReviewer(t, repositoryRoot, planlens, "simulated", "")
}

func runReviewWithReviewer(t *testing.T, repositoryRoot, planlens, reviewer, commandDirectory string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := exec.CommandContext(
		ctx,
		planlens,
		"review",
		"--plan", filepath.Join(repositoryRoot, "testdata", "public-plan.md"),
		"--reviewer", reviewer,
	)
	command.Dir = repositoryRoot
	command.Env = setEnvironment(os.Environ(), "PLANLENS_TEST_SECRET", "must-not-reach-reviewer")
	if commandDirectory != "" {
		command.Env = setEnvironment(command.Env, "PATH", commandDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	output, err := command.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("PlanLens did not exit after the review completed")
	}
	if err != nil {
		t.Fatalf("PlanLens review failed: %v\n%s", err, output)
	}
	return string(output)
}

func setEnvironment(environment []string, name, value string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return append(filtered, prefix+value)
}

func buildCommand(t *testing.T, repositoryRoot, outputDirectory, name, packagePath string) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	outputPath := filepath.Join(outputDirectory, name)
	goCommand := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goCommand += ".exe"
	}
	command := exec.Command(goCommand, "build", "-o", outputPath, packagePath)
	command.Dir = repositoryRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build %s: %v\n%s", packagePath, err, output)
	}
	return outputPath
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(workingDirectory, ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("locate repository root: %v", err)
	}
	return root
}
