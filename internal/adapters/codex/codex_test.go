package codex_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/codex"
	"github.com/wildbyteai/planlens/internal/review"
)

const supportedVersion = "0.146.0-alpha.3.1"

func TestDiscoverFindsSupportedAuthenticatedCodex(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)

	discovery, err := codex.Discover(testContext(t))
	if err != nil {
		t.Fatalf("discover Codex: %v", err)
	}
	if discovery.Adapter == nil {
		t.Fatal("discover Codex returned no adapter")
	}
	if got, want := discovery.Adapter.ID(), codex.ReviewerID; got != want {
		t.Fatalf("adapter ID = %q, want %q", got, want)
	}
	if got, want := discovery.CLIVersion, supportedVersion; got != want {
		t.Fatalf("CLI version = %q, want %q", got, want)
	}
	if got, want := discovery.AccessCapability, codex.AccessConstrained; got != want {
		t.Fatalf("access capability = %q, want %q", got, want)
	}
	if _, err := os.Stat(fake.secretLeakPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unrelated parent secret reached Codex subprocess: %v", err)
	}
}

func TestDiscoverRejectsUnsupportedVersionBeforeAuthentication(t *testing.T) {
	fake := prepareFakeCodex(t, "0.145.9", false)

	discovery, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrUnsupportedVersion) {
		t.Fatalf("discover error = %v, want ErrUnsupportedVersion", err)
	}
	if got, want := discovery.CLIVersion, "0.145.9"; got != want {
		t.Fatalf("reported CLI version = %q, want %q", got, want)
	}
	for _, invocation := range readInvocations(t, fake.invocationPath) {
		if slices.Equal(invocation.Args, []string{"login", "status"}) {
			t.Fatal("unsupported CLI version should be rejected before authentication probe")
		}
	}
}

func TestDiscoverRejectsUntestedNewerVersionBeforeAuthentication(t *testing.T) {
	fake := prepareFakeCodex(t, "0.146.1", false)

	discovery, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrUnsupportedVersion) {
		t.Fatalf("discover error = %v, want ErrUnsupportedVersion", err)
	}
	if got, want := discovery.CLIVersion, "0.146.1"; got != want {
		t.Fatalf("reported CLI version = %q, want %q", got, want)
	}
	for _, invocation := range readInvocations(t, fake.invocationPath) {
		if slices.Equal(invocation.Args, []string{"login", "status"}) {
			t.Fatal("untested CLI version should be rejected before authentication probe")
		}
	}
}

func TestDiscoverRejectsMissingNativeControl(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	writeFile(t, filepath.Join(fake.home, "fake-codex-exec-help.txt"), []byte(`--config
--strict-config
--sandbox
--cd
--skip-git-repo-check
--ephemeral
--ignore-user-config
--ignore-rules
--output-schema
--json
--output-last-message
--color
`))

	_, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrUnsupportedCapabilities) {
		t.Fatalf("discover error = %v, want ErrUnsupportedCapabilities", err)
	}
}

func TestDiscoverRejectsFeatureThatDoesNotDisable(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	writeFile(t, filepath.Join(fake.home, "fake-codex-enabled-feature.txt"), []byte("shell_tool"))

	_, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrUnsupportedCapabilities) {
		t.Fatalf("discover error = %v, want ErrUnsupportedCapabilities", err)
	}
	for _, invocation := range readInvocations(t, fake.invocationPath) {
		if slices.Equal(invocation.Args, []string{"login", "status"}) {
			t.Fatal("Codex authentication should not be probed after a required feature remains enabled")
		}
	}
}

func TestDiscoverRejectsConfigControlsThatExposeProjectInstructions(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	writeFile(t, filepath.Join(fake.binDirectory, "fake-codex-leak-capability-canary"), []byte("leak"))

	_, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrUnsupportedCapabilities) {
		t.Fatalf("discover error = %v, want ErrUnsupportedCapabilities", err)
	}
	for _, invocation := range readInvocations(t, fake.invocationPath) {
		if slices.Equal(invocation.Args, []string{"login", "status"}) {
			t.Fatal("Codex authentication should not be probed after project instructions remain visible")
		}
	}
}

func TestDiscoverRejectsGlobalInstructionFilesBeforeAuthentication(t *testing.T) {
	for _, name := range []string{"AGENTS.md", "AGENTS.override.md"} {
		t.Run(name, func(t *testing.T) {
			fake := prepareFakeCodex(t, supportedVersion, true)
			codexHome := filepath.Join(fake.home, ".codex")
			if err := os.MkdirAll(codexHome, 0o700); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(codexHome, name), []byte("must not be read\n"))

			_, err := codex.Discover(testContext(t))
			if !errors.Is(err, codex.ErrGlobalInstructionsPresent) {
				t.Fatalf("discover error = %v, want ErrGlobalInstructionsPresent", err)
			}
			for _, invocation := range readInvocations(t, fake.invocationPath) {
				if slices.Equal(invocation.Args, []string{"login", "status"}) ||
					(len(invocation.Args) >= 2 && slices.Equal(invocation.Args[:2], []string{"debug", "prompt-input"})) ||
					valueAfter(invocation.Args, "--output-last-message") != "" {
					t.Fatalf("global instructions were rejected after an auth or material execution: %q", invocation.Args)
				}
			}
		})
	}
}

func TestDiscoverTreatsGlobalInstructionSymlinkAsPresent(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	codexHome := filepath.Join(fake.home, ".codex")
	if err := os.MkdirAll(codexHome, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(fake.home, "instruction-target")
	writeFile(t, target, []byte("must not be read\n"))
	if err := os.Symlink(target, filepath.Join(codexHome, "AGENTS.md")); err != nil {
		t.Skipf("create global instruction symlink: %v", err)
	}

	_, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrGlobalInstructionsPresent) {
		t.Fatalf("discover error = %v, want ErrGlobalInstructionsPresent", err)
	}
}

func TestDiscoverPreservesCapabilityProbeContextFailure(t *testing.T) {
	tests := []struct {
		name             string
		context          func() (context.Context, context.CancelFunc)
		cancelAfterStart bool
		want             error
	}{
		{
			name: "cancellation",
			context: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			cancelAfterStart: true,
			want:             context.Canceled,
		},
		{
			name: "deadline",
			context: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 2*time.Second)
			},
			want: context.DeadlineExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := prepareFakeCodex(t, supportedVersion, true)
			writeFile(t, filepath.Join(fake.home, "fake-codex-block-root-help"), []byte("block"))

			ctx, cancel := test.context()
			defer cancel()
			errChannel := make(chan error, 1)
			go func() {
				_, err := codex.Discover(ctx)
				errChannel <- err
			}()

			waitForFile(t, filepath.Join(fake.home, "fake-codex-root-help-started"))
			if test.cancelAfterStart {
				cancel()
			}

			select {
			case err := <-errChannel:
				if !errors.Is(err, test.want) {
					t.Fatalf("discover error = %v, want %v", err, test.want)
				}
				if errors.Is(err, codex.ErrUnsupportedCapabilities) {
					t.Fatalf("capability probe context failure was misclassified: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Discover did not return after capability probe context failure")
			}
		})
	}
}

func TestDiscoverDiscardsAuthenticationDiagnostics(t *testing.T) {
	prepareFakeCodex(t, supportedVersion, false)

	_, err := codex.Discover(testContext(t))
	if !errors.Is(err, codex.ErrNotAuthenticated) {
		t.Fatalf("discover error = %v, want ErrNotAuthenticated", err)
	}
	for _, sensitive := range []string{"owner@example.com", "secret-token"} {
		if strings.Contains(err.Error(), sensitive) {
			t.Fatalf("authentication error retained raw diagnostic %q: %v", sensitive, err)
		}
	}
}

func TestRunStartsFreshConstrainedReviewAndNormalizesOnlyFinalMessage(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	want := review.Result{
		SchemaVersion: "1",
		Findings: []review.Finding{{
			Severity:        review.SeverityMajor,
			Title:           "No rollback trigger",
			Evidence:        "The plan deploys globally before defining a rollback decision.",
			Impact:          "A failed rollout could remain active.",
			SuggestedAction: "Add a rollback trigger and owner.",
		}},
	}
	finalMessage, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(fake.home, "fake-codex-final.json"), finalMessage)

	discovery, err := codex.Discover(testContext(t))
	if err != nil {
		t.Fatalf("discover Codex: %v", err)
	}
	plan := []byte("# Frozen public plan\n\nDeploy once, then observe errors.\n")
	first, err := discovery.Adapter.Run(testContext(t), review.Request{Plan: plan})
	if err != nil {
		t.Fatalf("run first review: %v", err)
	}
	second, err := discovery.Adapter.Run(testContext(t), review.Request{Plan: plan})
	if err != nil {
		t.Fatalf("run second review: %v", err)
	}
	if !resultsEqual(first, want) || !resultsEqual(second, want) {
		t.Fatalf("normalized results differ from final message\nfirst: %#v\nsecond: %#v\nwant: %#v", first, second, want)
	}
	for index, result := range []review.Result{first, second} {
		if got, want := result.Metadata.CLIVersion, supportedVersion; got != want {
			t.Fatalf("review %d CLI version = %q, want %q", index+1, got, want)
		}
		if got, want := result.Metadata.AccessCapability, review.AccessConstrained; got != want {
			t.Fatalf("review %d access capability = %q, want %q", index+1, got, want)
		}
	}

	runs := reviewInvocations(readInvocations(t, fake.invocationPath))
	if got, wantCount := len(runs), 2; got != wantCount {
		t.Fatalf("review invocation count = %d, want %d", got, wantCount)
	}
	if valueAfter(runs[0].Args, "--cd") == valueAfter(runs[1].Args, "--cd") {
		t.Fatalf("review invocations reused workspace %q", valueAfter(runs[0].Args, "--cd"))
	}
	for index, invocation := range runs {
		assertReviewInvocation(t, invocation, plan, filepath.Join(fake.home, ".codex"))
		if _, err := os.Stat(valueAfter(invocation.Args, "--cd")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("review %d workspace still exists after completion: %v", index+1, err)
		}
	}
	if _, err := os.Stat(fake.secretLeakPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unrelated parent secret reached Codex subprocess: %v", err)
	}
}

func TestRunRejectsInvalidFinalMessageWithoutRetainingEventStream(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	writeFile(t, filepath.Join(fake.home, "fake-codex-final.json"), []byte(`{"schema_version":"1","findings":[]}{"extra":true}`))

	discovery, err := codex.Discover(testContext(t))
	if err != nil {
		t.Fatalf("discover Codex: %v", err)
	}
	_, err = discovery.Adapter.Run(testContext(t), review.Request{Plan: []byte("public plan")})
	if !errors.Is(err, codex.ErrInvalidFinalResponse) {
		t.Fatalf("run error = %v, want ErrInvalidFinalResponse", err)
	}
	if strings.Contains(err.Error(), "private reasoning marker") {
		t.Fatalf("run error retained reasoning event: %v", err)
	}
}

func TestRunDiscardsFailedProcessDiagnostics(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	writeFile(t, filepath.Join(fake.home, "fake-codex-run-fail"), []byte("fail"))

	discovery, err := codex.Discover(testContext(t))
	if err != nil {
		t.Fatalf("discover Codex: %v", err)
	}
	_, err = discovery.Adapter.Run(testContext(t), review.Request{Plan: []byte("public plan")})
	if !errors.Is(err, codex.ErrExecutionFailed) {
		t.Fatalf("run error = %v, want ErrExecutionFailed", err)
	}
	for _, sensitive := range []string{"owner@example.com", "secret-token"} {
		if strings.Contains(err.Error(), sensitive) {
			t.Fatalf("run error retained raw diagnostic %q: %v", sensitive, err)
		}
	}
}

func TestRunRechecksGlobalInstructionsBeforeStartingReview(t *testing.T) {
	fake := prepareFakeCodex(t, supportedVersion, true)
	discovery, err := codex.Discover(testContext(t))
	if err != nil {
		t.Fatalf("discover Codex: %v", err)
	}
	codexHome := filepath.Join(fake.home, ".codex")
	if err := os.MkdirAll(codexHome, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(codexHome, "AGENTS.override.md"), []byte("must not be read\n"))

	_, err = discovery.Adapter.Run(testContext(t), review.Request{Plan: []byte("public plan")})
	if !errors.Is(err, codex.ErrGlobalInstructionsPresent) {
		t.Fatalf("run error = %v, want ErrGlobalInstructionsPresent", err)
	}
	if got := len(reviewInvocations(readInvocations(t, fake.invocationPath))); got != 0 {
		t.Fatalf("review invocation count = %d, want 0", got)
	}
}

type fakeCodex struct {
	home           string
	binDirectory   string
	invocationPath string
	secretLeakPath string
}

type invocation struct {
	Args         []string `json:"args"`
	CWD          string   `json:"cwd"`
	Stdin        string   `json:"stdin"`
	OutputSchema string   `json:"output_schema,omitempty"`
	Home         string   `json:"home"`
	UserProfile  string   `json:"user_profile"`
	CodexHome    string   `json:"codex_home"`
	Temp         string   `json:"temp"`
	Tmp          string   `json:"tmp"`
	TmpDir       string   `json:"tmpdir"`
}

func prepareFakeCodex(t *testing.T, version string, authenticated bool) fakeCodex {
	t.Helper()

	repositoryRoot := repositoryRoot(t)
	binDirectory := t.TempDir()
	binaryName := "codex"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(binDirectory, binaryName)
	goCommand := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goCommand += ".exe"
	}
	command := exec.Command(goCommand, "build", "-o", binaryPath, "./internal/adapters/codex/testdata/fakecodex")
	command.Dir = repositoryRoot
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fake Codex: %v\n%s", err, output)
	}

	home := t.TempDir()
	writeFile(t, filepath.Join(home, "fake-codex-version.txt"), []byte(version))
	if authenticated {
		writeFile(t, filepath.Join(home, "fake-codex-authenticated"), []byte("yes"))
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, ".codex"))
	t.Setenv("PATH", binDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PLANLENS_TEST_SECRET", "must-not-reach-codex")

	return fakeCodex{
		home:           home,
		binDirectory:   binDirectory,
		invocationPath: filepath.Join(home, "fake-codex-invocations.jsonl"),
		secretLeakPath: filepath.Join(home, "fake-codex-secret-leaked"),
	}
}

func assertReviewInvocation(t *testing.T, invocation invocation, plan []byte, wantCodexHome string) {
	t.Helper()

	if len(invocation.Args) < 3 {
		t.Fatalf("review args are incomplete: %q", invocation.Args)
	}
	if got, want := invocation.Args[:3], []string{"--ask-for-approval", "never", "exec"}; !slices.Equal(got, want) {
		t.Fatalf("review argument prefix = %q, want %q", got, want)
	}
	workspace := valueAfter(invocation.Args, "--cd")
	if workspace == "" {
		t.Fatalf("review args do not contain --cd: %q", invocation.Args)
	}
	if filepath.Base(workspace) != filepath.Base(invocation.CWD) {
		t.Fatalf("review process cwd = %q, want workspace equivalent to %q", invocation.CWD, workspace)
	}
	for name, got := range map[string]string{
		"HOME":        invocation.Home,
		"USERPROFILE": invocation.UserProfile,
		"TEMP":        invocation.Temp,
		"TMP":         invocation.Tmp,
		"TMPDIR":      invocation.TmpDir,
	} {
		if got != workspace {
			t.Fatalf("review %s = %q, want isolated workspace %q", name, got, workspace)
		}
	}
	if invocation.CodexHome != wantCodexHome {
		t.Fatalf("review CODEX_HOME = %q, want existing authentication home %q", invocation.CodexHome, wantCodexHome)
	}
	for flag, want := range map[string]string{
		"--sandbox":             "read-only",
		"--output-schema":       filepath.Join(workspace, "response-schema.json"),
		"--output-last-message": filepath.Join(workspace, "final-response.json"),
		"--color":               "never",
	} {
		if got := valueAfter(invocation.Args, flag); got != want {
			t.Fatalf("%s value = %q, want %q (args: %q)", flag, got, want, invocation.Args)
		}
	}
	if !slices.Contains(invocation.Args, "--strict-config") {
		t.Fatalf("review args do not contain --strict-config: %q", invocation.Args)
	}
	wantConfigs := []string{
		"project_doc_max_bytes=0",
		"shell_environment_policy.inherit=none",
		`web_search="disabled"`,
	}
	gotConfigs := valuesAfter(invocation.Args, "--config")
	slices.Sort(gotConfigs)
	if !slices.Equal(gotConfigs, wantConfigs) {
		t.Fatalf("review config overrides = %q, want %q", gotConfigs, wantConfigs)
	}
	wantDisabledFeatures := []string{
		"apps",
		"browser_use",
		"browser_use_external",
		"browser_use_full_cdp_access",
		"code_mode_host",
		"computer_use",
		"hooks",
		"image_generation",
		"in_app_browser",
		"multi_agent",
		"plugin_sharing",
		"plugins",
		"remote_plugin",
		"shell_snapshot",
		"shell_tool",
		"skill_mcp_dependency_install",
		"skill_search",
		"tool_call_mcp_elicitation",
		"tool_suggest",
		"unified_exec",
		"workspace_dependencies",
	}
	gotDisabledFeatures := valuesAfter(invocation.Args, "--disable")
	slices.Sort(gotDisabledFeatures)
	if !slices.Equal(gotDisabledFeatures, wantDisabledFeatures) {
		t.Fatalf("disabled Codex features = %q, want %q", gotDisabledFeatures, wantDisabledFeatures)
	}
	for _, flag := range []string{
		"--skip-git-repo-check",
		"--ephemeral",
		"--ignore-user-config",
		"--ignore-rules",
		"--json",
	} {
		if !slices.Contains(invocation.Args, flag) {
			t.Fatalf("review args do not contain %s: %q", flag, invocation.Args)
		}
	}
	for _, forbidden := range []string{"resume", "--add-dir", "--dangerously-bypass-approvals-and-sandbox", "--yolo"} {
		if slices.Contains(invocation.Args, forbidden) {
			t.Fatalf("review args contain forbidden option %q: %q", forbidden, invocation.Args)
		}
	}
	if got := invocation.Args[len(invocation.Args)-1]; got != "-" {
		t.Fatalf("last review argument = %q, want stdin marker", got)
	}
	if count := strings.Count(invocation.Stdin, string(plan)); count != 1 {
		t.Fatalf("frozen plan occurred %d times in prompt, want exactly once\nprompt:\n%s", count, invocation.Stdin)
	}
	if strings.Contains(strings.ToLower(invocation.Stdin), "previous reviewer") {
		t.Fatalf("prompt referenced another reviewer result:\n%s", invocation.Stdin)
	}
	var schema map[string]any
	if err := json.Unmarshal([]byte(invocation.OutputSchema), &schema); err != nil {
		t.Fatalf("output schema is not valid JSON: %v\n%s", err, invocation.OutputSchema)
	}
	if schema["additionalProperties"] != false {
		t.Fatalf("output schema does not reject unknown top-level properties: %#v", schema)
	}
}

func valueAfter(args []string, flag string) string {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == flag {
			return args[index+1]
		}
	}
	return ""
}

func valuesAfter(args []string, flag string) []string {
	var values []string
	for index := 0; index+1 < len(args); index++ {
		if args[index] == flag {
			values = append(values, args[index+1])
			index++
		}
	}
	return values
}

func reviewInvocations(invocations []invocation) []invocation {
	var reviews []invocation
	for _, invocation := range invocations {
		if valueAfter(invocation.Args, "--output-last-message") != "" {
			reviews = append(reviews, invocation)
		}
	}
	return reviews
}

func readInvocations(t *testing.T, path string) []invocation {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fake Codex invocations: %v", err)
	}
	defer file.Close()

	var invocations []invocation
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item invocation
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			t.Fatalf("decode fake Codex invocation: %v", err)
		}
		invocations = append(invocations, item)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan fake Codex invocations: %v", err)
	}
	return invocations
}

func resultsEqual(left, right review.Result) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}

func writeFile(t *testing.T, path string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write %s: %v", filepath.Base(path), err)
	}
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("wait for %s: %v", filepath.Base(path), err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", filepath.Base(path))
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
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
