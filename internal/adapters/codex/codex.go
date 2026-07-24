package codex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/wildbyteai/planlens/internal/review"
)

const (
	ReviewerID       = "codex"
	TestedCLIVersion = "0.146.0-alpha.3.1"

	maxFinalResponseBytes = 1 << 20
	capabilityProbePrompt = "planlens-public-capability-probe"
	capabilityProbeCanary = "PLANLENS_PUBLIC_AGENTS_CANARY_MUST_NOT_REACH_MODEL"
)

const AccessConstrained = review.AccessConstrained

var (
	ErrCLINotFound               = errors.New("codex CLI not found")
	ErrVersionProbeFailed        = errors.New("codex CLI version probe failed")
	ErrUnsupportedVersion        = errors.New("unsupported codex CLI version")
	ErrUnsupportedCapabilities   = errors.New("codex CLI lacks required controls")
	ErrGlobalInstructionsPresent = errors.New("codex home contains global instructions")
	ErrNotAuthenticated          = errors.New("codex CLI is not authenticated")
	ErrExecutionFailed           = errors.New("codex review process failed")
	ErrInvalidFinalResponse      = errors.New("codex final response is invalid")
	ErrCleanupFailed             = errors.New("codex review workspace cleanup failed")
)

var cliVersionPattern = regexp.MustCompile(`(?m)^codex-cli[\t ]+([0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?)\s*$`)

var requiredConfigOverrides = []string{
	"project_doc_max_bytes=0",
	"shell_environment_policy.inherit=none",
	`web_search="disabled"`,
}

var disabledFeatures = []string{
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

type Discovery struct {
	Adapter          *Adapter
	CLIVersion       string
	AccessCapability review.AccessCapability
}

type Adapter struct {
	executable string
	version    string
	codexHome  string
}

func Discover(ctx context.Context) (Discovery, error) {
	discovery := Discovery{AccessCapability: AccessConstrained}
	executable, err := exec.LookPath("codex")
	if err != nil {
		return discovery, ErrCLINotFound
	}

	versionOutput, err := captureCommand(ctx, executable, "--version")
	if err != nil {
		return discovery, err
	}
	version, err := parseCLIVersion(versionOutput)
	if err != nil {
		return discovery, err
	}
	discovery.CLIVersion = version
	if version != TestedCLIVersion {
		return discovery, fmt.Errorf("%w: found %s, tested candidate is %s", ErrUnsupportedVersion, version, TestedCLIVersion)
	}

	codexHome := existingCodexHome()
	if err := requireNoGlobalInstructions(codexHome); err != nil {
		return discovery, err
	}
	if err := requireCapabilities(ctx, executable); err != nil {
		return discovery, err
	}
	if err := requireAuthentication(ctx, executable, codexHome); err != nil {
		return discovery, err
	}

	discovery.Adapter = &Adapter{executable: executable, version: version, codexHome: codexHome}
	return discovery, nil
}

func New(ctx context.Context) (*Adapter, error) {
	discovery, err := Discover(ctx)
	if err != nil {
		return nil, err
	}
	return discovery.Adapter, nil
}

func (adapter *Adapter) ID() string {
	return ReviewerID
}

func (adapter *Adapter) Version() string {
	return adapter.version
}

func (adapter *Adapter) AccessCapability() review.AccessCapability {
	return AccessConstrained
}

func (adapter *Adapter) Run(ctx context.Context, request review.Request) (result review.Result, runErr error) {
	if err := requireNoGlobalInstructions(adapter.codexHome); err != nil {
		return review.Result{}, err
	}
	workspace, err := os.MkdirTemp("", "planlens-codex-review-*")
	if err != nil {
		return review.Result{}, fmt.Errorf("prepare Codex review: %w", ErrExecutionFailed)
	}
	defer func() {
		if err := os.RemoveAll(workspace); err != nil {
			result = review.Result{}
			runErr = errors.Join(runErr, ErrCleanupFailed)
		}
	}()

	schemaPath := filepath.Join(workspace, "response-schema.json")
	if err := os.WriteFile(schemaPath, []byte(responseSchema), 0o600); err != nil {
		return review.Result{}, fmt.Errorf("prepare Codex review schema: %w", ErrExecutionFailed)
	}
	finalResponsePath := filepath.Join(workspace, "final-response.json")

	args := []string{
		"--ask-for-approval", "never",
		"exec",
		"--strict-config",
	}
	for _, override := range requiredConfigOverrides {
		args = append(args, "--config", override)
	}
	for _, feature := range disabledFeatures {
		args = append(args, "--disable", feature)
	}
	args = append(args,
		"--sandbox", "read-only",
		"--cd", workspace,
		"--skip-git-repo-check",
		"--ephemeral",
		"--ignore-user-config",
		"--ignore-rules",
		"--output-schema", schemaPath,
		"--output-last-message", finalResponsePath,
		"--json",
		"--color", "never",
		"-",
	)
	command := exec.CommandContext(ctx, adapter.executable, args...)
	command.Dir = workspace
	command.Env = reviewEnvironment(workspace, adapter.codexHome)
	command.Stdin = bytes.NewReader(promptFor(request.Plan))
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return review.Result{}, fmt.Errorf("run Codex review: %w", ctx.Err())
		}
		return review.Result{}, ErrExecutionFailed
	}

	contents, err := readBoundedFile(finalResponsePath, maxFinalResponseBytes)
	if err != nil {
		return review.Result{}, ErrInvalidFinalResponse
	}
	result, err = review.DecodeResult(bytes.NewReader(contents))
	if err != nil {
		return review.Result{}, ErrInvalidFinalResponse
	}
	result.Metadata = review.Metadata{
		CLIVersion:       adapter.version,
		AccessCapability: review.AccessConstrained,
	}
	return result, nil
}

func requireCapabilities(ctx context.Context, executable string) error {
	rootHelp, err := captureCommand(ctx, executable, "--help")
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return ErrUnsupportedCapabilities
	}
	if !bytes.Contains(rootHelp, []byte("--ask-for-approval")) {
		return fmt.Errorf("%w: missing --ask-for-approval", ErrUnsupportedCapabilities)
	}

	execHelp, err := captureCommand(ctx, executable, "exec", "--help")
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return ErrUnsupportedCapabilities
	}
	required := []string{
		"--config",
		"--strict-config",
		"--disable",
		"--sandbox",
		"--cd",
		"--skip-git-repo-check",
		"--ephemeral",
		"--ignore-user-config",
		"--ignore-rules",
		"--output-schema",
		"--json",
		"--output-last-message",
		"--color",
	}
	for _, option := range required {
		if !bytes.Contains(execHelp, []byte(option)) {
			return fmt.Errorf("%w: missing %s", ErrUnsupportedCapabilities, option)
		}
	}
	if err := requireDisabledFeatureControls(ctx, executable); err != nil {
		return err
	}
	if err := requireConfigIsolationControls(ctx, executable); err != nil {
		return err
	}
	return nil
}

func requireDisabledFeatureControls(ctx context.Context, executable string) error {
	args := make([]string, 0, 2*(len(requiredConfigOverrides)+len(disabledFeatures))+2)
	for _, override := range requiredConfigOverrides {
		args = append(args, "--config", override)
	}
	for _, feature := range disabledFeatures {
		args = append(args, "--disable", feature)
	}
	args = append(args, "features", "list")
	output, err := captureCommand(ctx, executable, args...)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return ErrUnsupportedCapabilities
	}

	effectiveState := make(map[string]string, len(disabledFeatures))
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			effectiveState[fields[0]] = fields[2]
		}
	}
	for _, feature := range disabledFeatures {
		if effectiveState[feature] != "false" {
			return fmt.Errorf("%w: cannot disable feature %s", ErrUnsupportedCapabilities, feature)
		}
	}
	return nil
}

func requireConfigIsolationControls(ctx context.Context, executable string) (probeErr error) {
	workspace, err := os.MkdirTemp("", "planlens-codex-capability-*")
	if err != nil {
		return fmt.Errorf("%w: create config capability workspace", ErrUnsupportedCapabilities)
	}
	defer func() {
		if err := os.RemoveAll(workspace); err != nil {
			probeErr = errors.Join(probeErr, ErrCleanupFailed)
		}
	}()
	if err := os.Mkdir(filepath.Join(workspace, ".codex"), 0o700); err != nil {
		return fmt.Errorf("%w: prepare isolated Codex home", ErrUnsupportedCapabilities)
	}
	if err := os.WriteFile(filepath.Join(workspace, "AGENTS.md"), []byte(capabilityProbeCanary+"\n"), 0o600); err != nil {
		return fmt.Errorf("%w: write public capability canary", ErrUnsupportedCapabilities)
	}

	args := []string{"debug", "prompt-input"}
	for _, override := range requiredConfigOverrides {
		args = append(args, "--config", override)
	}
	for _, feature := range disabledFeatures {
		args = append(args, "--disable", feature)
	}
	args = append(args, capabilityProbePrompt)
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = workspace
	command.Env = capabilityProbeEnvironment(workspace)
	command.Stderr = io.Discard
	var stdout bytes.Buffer
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrUnsupportedCapabilities
	}
	if !bytes.Contains(stdout.Bytes(), []byte(capabilityProbePrompt)) || bytes.Contains(stdout.Bytes(), []byte(capabilityProbeCanary)) {
		return ErrUnsupportedCapabilities
	}
	return nil
}

func requireNoGlobalInstructions(codexHome string) error {
	if codexHome == "" {
		return ErrGlobalInstructionsPresent
	}
	for _, name := range []string{"AGENTS.override.md", "AGENTS.md"} {
		_, err := os.Lstat(filepath.Join(codexHome, name))
		switch {
		case err == nil:
			return ErrGlobalInstructionsPresent
		case errors.Is(err, os.ErrNotExist):
			continue
		default:
			return ErrGlobalInstructionsPresent
		}
	}
	return nil
}

func requireAuthentication(ctx context.Context, executable, codexHome string) error {
	command := exec.CommandContext(ctx, executable, "login", "status")
	command.Env = authenticationEnvironment(codexHome)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrNotAuthenticated
	}
	return nil
}

func captureCommand(ctx context.Context, executable string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = subprocessEnvironment()
	command.Stderr = io.Discard
	var stdout bytes.Buffer
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrVersionProbeFailed
	}
	return stdout.Bytes(), nil
}

func parseCLIVersion(output []byte) (string, error) {
	match := cliVersionPattern.FindSubmatch(output)
	if len(match) != 2 {
		return "", ErrVersionProbeFailed
	}
	version := string(match[1])
	return version, nil
}

func subprocessEnvironment() []string {
	keys := append([]string{}, sharedEnvironmentKeys...)
	keys = append(keys,
		"CODEX_HOME",
		"HOME",
		"SHELL",
		"TEMP",
		"TMP",
		"TMPDIR",
		"USERPROFILE",
	)
	return inheritedEnvironment(keys)
}

func authenticationEnvironment(codexHome string) []string {
	keys := append([]string{}, sharedEnvironmentKeys...)
	keys = append(keys,
		"HOME",
		"SHELL",
		"TEMP",
		"TMP",
		"TMPDIR",
		"USERPROFILE",
	)
	environment := inheritedEnvironment(keys)
	environment = append(environment, "CODEX_HOME="+codexHome)
	sort.Strings(environment)
	return environment
}

func reviewEnvironment(workspace, codexHome string) []string {
	return isolatedEnvironment(workspace, codexHome)
}

func capabilityProbeEnvironment(workspace string) []string {
	return isolatedEnvironment(workspace, filepath.Join(workspace, ".codex"))
}

func isolatedEnvironment(workspace, codexHome string) []string {
	environment := inheritedEnvironment(sharedEnvironmentKeys)
	if codexHome != "" {
		environment = append(environment, "CODEX_HOME="+codexHome)
	}
	environment = append(environment,
		"HOME="+workspace,
		"TEMP="+workspace,
		"TMP="+workspace,
		"TMPDIR="+workspace,
		"USERPROFILE="+workspace,
	)
	sort.Strings(environment)
	return environment
}

func existingCodexHome() string {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil || userHome == "" {
			return ""
		}
		home = filepath.Join(userHome, ".codex")
	}
	absoluteHome, err := filepath.Abs(home)
	if err != nil {
		return ""
	}
	return absoluteHome
}

var sharedEnvironmentKeys = []string{
	"ALL_PROXY",
	"HTTPS_PROXY",
	"HTTP_PROXY",
	"LANG",
	"LC_ALL",
	"LC_CTYPE",
	"NO_PROXY",
	"PATH",
	"PATHEXT",
	"SSL_CERT_DIR",
	"SSL_CERT_FILE",
	"SYSTEMROOT",
	"WINDIR",
	"all_proxy",
	"http_proxy",
	"https_proxy",
	"no_proxy",
}

func inheritedEnvironment(keys []string) []string {
	var environment []string
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			environment = append(environment, key+"="+value)
		}
	}
	sort.Strings(environment)
	return environment
}

func promptFor(plan []byte) []byte {
	var prompt bytes.Buffer
	prompt.WriteString("Review only the frozen plan below. Identify concrete risks, missing decisions, and unsupported assumptions.\n")
	prompt.WriteString("Do not inspect files, run commands, search the web, or use any material outside this prompt.\n")
	prompt.WriteString("Return only a response matching the supplied JSON schema. Base every finding on evidence in the plan.\n\n")
	prompt.WriteString("<frozen-plan>\n")
	prompt.Write(plan)
	if len(plan) == 0 || plan[len(plan)-1] != '\n' {
		prompt.WriteByte('\n')
	}
	prompt.WriteString("</frozen-plan>\n")
	return prompt.Bytes()
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil || int64(len(contents)) > limit {
		return nil, errors.New("final response exceeds limit")
	}
	return contents, nil
}

const responseSchema = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["schema_version", "findings"],
  "properties": {
    "schema_version": {
      "type": "string",
      "enum": ["1"]
    },
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["severity", "title", "evidence", "impact", "suggested_action"],
        "properties": {
          "severity": {
            "type": "string",
            "enum": ["blocking", "major", "minor"]
          },
          "title": {"type": "string"},
          "evidence": {"type": "string"},
          "impact": {"type": "string"},
          "suggested_action": {"type": "string"}
        }
      }
    }
  }
}`
