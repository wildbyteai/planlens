package kimi_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/kimi"
	"github.com/wildbyteai/planlens/internal/review"
)

const publicFixtureSHA256 = "7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429"

func TestNewRequiresExplicitKimiCodeHome(t *testing.T) {
	prepareFakeKimi(t)
	t.Setenv("KIMI_CODE_HOME", "")

	_, err := kimi.New(context.Background())
	if err == nil || !strings.Contains(err.Error(), "KIMI_CODE_HOME") {
		t.Fatalf("New error = %v, want explicit KIMI_CODE_HOME requirement", err)
	}
}

func TestNewRequiresAbsoluteExistingKimiCodeHomeDirectory(t *testing.T) {
	tests := []struct {
		name string
		home func(*testing.T) string
	}{
		{
			name: "relative path",
			home: func(*testing.T) string { return "relative-kimi-home" },
		},
		{
			name: "missing path",
			home: func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing") },
		},
		{
			name: "regular file",
			home: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "kimi-home")
				if err := os.WriteFile(path, []byte("not a directory\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				return path
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prepareFakeKimi(t)
			t.Setenv("KIMI_CODE_HOME", test.home(t))

			_, err := kimi.New(context.Background())
			if err == nil || !strings.Contains(err.Error(), "KIMI_CODE_HOME") {
				t.Fatalf("New error = %v, want invalid KIMI_CODE_HOME rejection", err)
			}
		})
	}
}

func TestNewRejectsKimiCodeHomeWithImplicitLoadFiles(t *testing.T) {
	for _, relativePath := range []string{
		"AGENTS.md",
		"SYSTEM.md",
		"config.toml",
		"mcp.json",
		filepath.Join("plugins", "installed.json"),
	} {
		t.Run(relativePath, func(t *testing.T) {
			prepareFakeKimi(t)
			kimiHome := t.TempDir()
			writeFile(t, filepath.Join(kimiHome, relativePath), []byte("must not be loaded\n"))
			t.Setenv("KIMI_CODE_HOME", kimiHome)

			_, err := kimi.New(context.Background())
			if err == nil || !strings.Contains(err.Error(), relativePath) {
				t.Fatalf("New error = %v, want rejection naming %q", err, relativePath)
			}
		})
	}
}

func TestNewTreatsImplicitLoadSymlinksAsPresent(t *testing.T) {
	for _, relativePath := range []string{
		"AGENTS.md",
		"SYSTEM.md",
		"config.toml",
		"mcp.json",
		filepath.Join("plugins", "installed.json"),
	} {
		t.Run(relativePath, func(t *testing.T) {
			prepareFakeKimi(t)
			kimiHome := t.TempDir()
			linkPath := filepath.Join(kimiHome, relativePath)
			if err := os.MkdirAll(filepath.Dir(linkPath), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(t.TempDir(), "missing-target"), linkPath); err != nil {
				t.Skipf("create implicit-load symlink: %v", err)
			}
			t.Setenv("KIMI_CODE_HOME", kimiHome)

			_, err := kimi.New(context.Background())
			if err == nil || !strings.Contains(err.Error(), relativePath) {
				t.Fatalf("New error = %v, want symlink rejection naming %q", err, relativePath)
			}
		})
	}
}

func TestRunRechecksKimiCodeHomeBeforeStartingReview(t *testing.T) {
	prepareFakeKimi(t)
	kimiHome := t.TempDir()
	t.Setenv("KIMI_CODE_HOME", kimiHome)

	reviewer, err := kimi.New(context.Background())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	writeFile(t, filepath.Join(kimiHome, "config.toml"), []byte("added after discovery\n"))

	_, err = reviewer.Run(context.Background(), review.Request{Plan: publicPlan(t)})
	if err == nil || !strings.Contains(err.Error(), "config.toml") {
		t.Fatalf("Run error = %v, want late config.toml rejection", err)
	}
}

func TestNewAndRunUseSterileSubprocessEnvironment(t *testing.T) {
	prepareFakeKimi(t)
	kimiHome := t.TempDir()
	t.Setenv("KIMI_CODE_HOME", kimiHome)
	parentEnvironmentDirectory := t.TempDir()
	for _, name := range []string{
		"HOME",
		"USERPROFILE",
		"APPDATA",
		"LOCALAPPDATA",
		"TEMP",
		"TMP",
		"TMPDIR",
	} {
		t.Setenv(name, parentEnvironmentDirectory)
	}
	t.Setenv("PLANLENS_TEST_SECRET", "must-not-reach-kimi")

	reviewer, err := kimi.New(context.Background())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	result, err := reviewer.Run(context.Background(), review.Request{Plan: publicPlan(t)})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got, want := result.Metadata.CLIVersion, "0.29.1"; got != want {
		t.Fatalf("CLI version = %q, want %q", got, want)
	}
}

func TestSharedPublicFixtureSurvivesWindowsStyleGitCheckout(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	checkoutRoot := t.TempDir()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is required to verify checkout line-ending behavior")
	}
	command := exec.Command(
		git,
		"-c", "core.autocrlf=true",
		"-c", "core.eol=crlf",
		"checkout-index",
		"--force",
		"--prefix="+filepath.ToSlash(checkoutRoot)+"/",
		"--",
		"testdata/public-plan.md",
	)
	command.Dir = repositoryRoot
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("check out public fixture with Windows line-ending settings: %v\n%s", err, output)
	}

	plan, err := os.ReadFile(filepath.Join(checkoutRoot, "testdata", "public-plan.md"))
	if err != nil {
		t.Fatalf("read checked-out public fixture: %v", err)
	}
	if bytes.Contains(plan, []byte("\r\n")) {
		t.Fatal("Windows-style Git checkout converted the public fixture to CRLF")
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(plan))
	if digest != publicFixtureSHA256 {
		t.Fatalf("checked-out public fixture SHA-256 = %s, want %s", digest, publicFixtureSHA256)
	}
}

func TestNewPreservesVersionProbeContextFailure(t *testing.T) {
	tests := []struct {
		name       string
		newContext func() (context.Context, func())
		want       error
	}{
		{
			name: "cancellation",
			newContext: func() (context.Context, func()) {
				return context.WithCancel(context.Background())
			},
			want: context.Canceled,
		},
		{
			name: "deadline",
			newContext: func() (context.Context, func()) {
				ctx := newTriggeredContext()
				return ctx, func() { ctx.trigger(context.DeadlineExceeded) }
			},
			want: context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prepareFakeKimi(t)
			kimiHome := t.TempDir()
			writeFile(t, filepath.Join(kimiHome, "fake-kimi-block-version"), []byte("yes\n"))
			t.Setenv("KIMI_CODE_HOME", kimiHome)

			ctx, trigger := test.newContext()
			defer trigger()
			errChannel := make(chan error, 1)
			go func() {
				_, err := kimi.New(ctx)
				errChannel <- err
			}()

			waitForFile(t, filepath.Join(kimiHome, "fake-kimi-version-started"))
			trigger()
			select {
			case err := <-errChannel:
				if !errors.Is(err, test.want) {
					t.Fatalf("New error = %v, want %v", err, test.want)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("New did not return after version probe context failure")
			}
		})
	}
}

type triggeredContext struct {
	done chan struct{}
	once sync.Once
	mu   sync.RWMutex
	err  error
}

func newTriggeredContext() *triggeredContext {
	return &triggeredContext{done: make(chan struct{})}
}

func (ctx *triggeredContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (ctx *triggeredContext) Done() <-chan struct{} {
	return ctx.done
}

func (ctx *triggeredContext) Err() error {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.err
}

func (*triggeredContext) Value(any) any {
	return nil
}

func (ctx *triggeredContext) trigger(err error) {
	ctx.once.Do(func() {
		ctx.mu.Lock()
		ctx.err = err
		ctx.mu.Unlock()
		close(ctx.done)
	})
}

func TestRunPreservesCanceledContextWhenProcessCannotStart(t *testing.T) {
	prepareFakeKimi(t)
	kimiHome := t.TempDir()
	t.Setenv("KIMI_CODE_HOME", kimiHome)
	reviewer, err := kimi.New(context.Background())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = reviewer.Run(ctx, review.Request{Plan: publicPlan(t)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
}

func TestRunPreservesDeadlineBeforeAuthenticationClassification(t *testing.T) {
	prepareFakeKimi(t)
	kimiHome := t.TempDir()
	t.Setenv("KIMI_CODE_HOME", kimiHome)
	reviewer, err := kimi.New(context.Background())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	writeFile(t, filepath.Join(kimiHome, "fake-kimi-block-after-auth-diagnostic"), []byte("yes\n"))

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()
	_, err = reviewer.Run(ctx, review.Request{Plan: publicPlan(t)})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want context.DeadlineExceeded", err)
	}
	if strings.Contains(err.Error(), "authentication is required") {
		t.Fatalf("deadline was misclassified as authentication failure: %v", err)
	}
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
	t.Fatalf("timed out waiting for %s", path)
}

func publicPlan(t *testing.T) []byte {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repositoryRoot(t), "testdata", "public-plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func writeFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
}

func prepareFakeKimi(t *testing.T) string {
	t.Helper()

	repositoryRoot := repositoryRoot(t)
	binDirectory := t.TempDir()
	binaryName := "kimi"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(binDirectory, binaryName)
	goCommand := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goCommand += ".exe"
	}
	command := exec.Command(goCommand, "build", "-o", binaryPath, "./test/fixtures/fakekimi")
	command.Dir = repositoryRoot
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fake Kimi: %v\n%s", err, output)
	}
	t.Setenv("PATH", binDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	return binaryPath
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
