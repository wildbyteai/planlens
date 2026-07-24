package antigravity

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestReviewReportsWorkspaceCleanupFailure(t *testing.T) {
	repositoryRoot := cleanupTestRepositoryRoot(t)
	executableDirectory := t.TempDir()
	executableName := "agy"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executable := filepath.Join(executableDirectory, executableName)
	goExecutable := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goExecutable += ".exe"
	}
	command := exec.Command(goExecutable, "build", "-o", executable, "./internal/adapters/antigravity/testdata/fakeagy")
	command.Dir = repositoryRoot
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fake agy: %v\n%s", err, output)
	}

	t.Setenv("PATH", executableDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	adapter, err := Discover(context.Background())
	if err != nil {
		t.Fatalf("discover Antigravity CLI: %v", err)
	}
	adapter.removeWorkspace = func(path string) error {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		return errors.New("forced cleanup failure")
	}

	fixture, err := os.ReadFile(filepath.Join(repositoryRoot, "testdata", "public-plan.md"))
	if err != nil {
		t.Fatalf("read public fixture: %v", err)
	}
	result, err := adapter.Review(context.Background(), ReviewRequest{
		Plan:    fixture,
		Timeout: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected cleanup failure")
	}
	if result != (ReviewResult{}) {
		t.Fatalf("cleanup failure returned a usable result: %#v", result)
	}
	var runError *RunError
	if !errors.As(err, &runError) || runError.Kind != FailureCleanup {
		t.Fatalf("expected cleanup failure kind, got %T: %v", err, err)
	}
}

func cleanupTestRepositoryRoot(t *testing.T) string {
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
