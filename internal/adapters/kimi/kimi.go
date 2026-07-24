package kimi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wildbyteai/planlens/internal/review"
)

const (
	ID                     = "kimi"
	testedCandidateVersion = "0.29.1"
	publicFixtureDigest    = "7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429"
)

var semanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`)

var implicitLoadPaths = []string{
	"AGENTS.md",
	"SYSTEM.md",
	"config.toml",
	"mcp.json",
	filepath.Join("plugins", "installed.json"),
}

type reviewer struct {
	executable string
	version    string
	kimiHome   string
}

func New(ctx context.Context) (review.Reviewer, error) {
	executable, err := exec.LookPath("kimi")
	if err != nil {
		return nil, errors.New("Kimi Code CLI is not installed or is not on PATH")
	}
	kimiHome, err := configuredKimiHome()
	if err != nil {
		return nil, err
	}
	if err := requireSafeKimiHome(kimiHome); err != nil {
		return nil, err
	}

	output, err := captureVersion(ctx, executable, kimiHome)
	if err != nil {
		return nil, fmt.Errorf("detect Kimi Code CLI version: %w", err)
	}
	version := strings.TrimSpace(string(output))
	if !semanticVersion.MatchString(version) {
		return nil, fmt.Errorf("detect Kimi Code CLI version: unexpected output %q", version)
	}
	if version != testedCandidateVersion {
		return nil, fmt.Errorf("Kimi Code CLI version %s is untested by this feasibility adapter; tested candidate is %s", version, testedCandidateVersion)
	}

	return reviewer{executable: executable, version: version, kimiHome: kimiHome}, nil
}

func configuredKimiHome() (string, error) {
	kimiHome := os.Getenv("KIMI_CODE_HOME")
	if kimiHome == "" {
		return "", errors.New("KIMI_CODE_HOME must explicitly name a dedicated Kimi Code CLI home")
	}
	if !filepath.IsAbs(kimiHome) {
		return "", errors.New("KIMI_CODE_HOME must be an absolute path")
	}
	kimiHome = filepath.Clean(kimiHome)
	info, err := os.Stat(kimiHome)
	if err != nil {
		return "", fmt.Errorf("KIMI_CODE_HOME must name an existing directory: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("KIMI_CODE_HOME must name an existing directory")
	}
	return kimiHome, nil
}

func requireSafeKimiHome(kimiHome string) error {
	for _, relativePath := range implicitLoadPaths {
		_, err := os.Lstat(filepath.Join(kimiHome, relativePath))
		switch {
		case err == nil:
			return fmt.Errorf("KIMI_CODE_HOME contains an implicit-load path %q", relativePath)
		case errors.Is(err, os.ErrNotExist):
			continue
		default:
			return fmt.Errorf("cannot safely inspect KIMI_CODE_HOME implicit-load path %q: %w", relativePath, err)
		}
	}
	return nil
}

func (reviewer) ID() string {
	return ID
}

func (adapter reviewer) Run(ctx context.Context, request review.Request) (result review.Result, runErr error) {
	planDigest := fmt.Sprintf("%x", sha256.Sum256(request.Plan))
	if planDigest != publicFixtureDigest {
		return review.Result{}, errors.New("Kimi Code CLI feasibility adapter only accepts the fixed public fixture")
	}
	if err := requireSafeKimiHome(adapter.kimiHome); err != nil {
		return review.Result{}, err
	}

	workspace, err := os.MkdirTemp("", "planlens-kimi-*")
	if err != nil {
		return review.Result{}, fmt.Errorf("create Kimi review workspace: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workspace); err != nil {
			result = review.Result{}
			runErr = errors.Join(runErr, fmt.Errorf("remove Kimi review workspace: %w", err))
		}
	}()

	skillsDirectory := filepath.Join(workspace, "skills")
	if err := os.Mkdir(skillsDirectory, 0o700); err != nil {
		return review.Result{}, fmt.Errorf("create empty Kimi skills directory: %w", err)
	}
	agentPath := filepath.Join(workspace, "planlens-reviewer.md")
	if err := os.WriteFile(agentPath, []byte(agentDefinition), 0o600); err != nil {
		return review.Result{}, fmt.Errorf("create Kimi reviewer definition: %w", err)
	}

	command := exec.CommandContext(
		ctx,
		adapter.executable,
		"--output-format", "stream-json",
		"--skills-dir", skillsDirectory,
		"--agent-file", agentPath,
		"--prompt", reviewPrompt(request.Plan),
	)
	command.Dir = workspace
	command.Env = append(
		isolatedEnvironment(workspace, adapter.kimiHome),
		"KIMI_CODE_EXPERIMENTAL_FLAG=1",
	)

	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		return review.Result{}, fmt.Errorf("capture Kimi Code CLI final response: %w", err)
	}
	if err := command.Start(); err != nil {
		if ctx.Err() != nil {
			return review.Result{}, fmt.Errorf("start Kimi Code CLI review: %w", ctx.Err())
		}
		return review.Result{}, fmt.Errorf("start Kimi Code CLI review: %w", err)
	}

	finalResponse, responseErr := finalAssistantResponse(stdout, adapter.version)
	if responseErr != nil {
		_, _ = io.Copy(io.Discard, stdout)
	}
	if err := command.Wait(); err != nil {
		if ctx.Err() != nil {
			return review.Result{}, fmt.Errorf("run Kimi Code CLI review: %w", ctx.Err())
		}
		if authenticationRequired(stderr.String()) {
			return review.Result{}, errors.New("Kimi Code CLI authentication is required; run `kimi login` through the official CLI, then retry")
		}
		return review.Result{}, fmt.Errorf("Kimi Code CLI review failed: %w", err)
	}
	if responseErr != nil {
		return review.Result{}, responseErr
	}
	result, err = review.DecodeResult(strings.NewReader(finalResponse))
	if err != nil {
		return review.Result{}, err
	}
	result.Metadata = review.Metadata{
		CLIVersion:       adapter.version,
		AccessCapability: review.AccessConstrained,
	}
	return result, nil
}

func inheritedEnvironment(keys ...string) []string {
	environment := make([]string, 0, len(keys))
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			environment = append(environment, key+"="+value)
		}
	}
	return environment
}

func captureVersion(ctx context.Context, executable, kimiHome string) (output []byte, probeErr error) {
	workspace, err := os.MkdirTemp("", "planlens-kimi-version-*")
	if err != nil {
		return nil, fmt.Errorf("create isolated version workspace: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workspace); err != nil {
			output = nil
			probeErr = errors.Join(probeErr, fmt.Errorf("remove isolated version workspace: %w", err))
		}
	}()

	command := exec.CommandContext(ctx, executable, "--version")
	command.Dir = workspace
	command.Env = isolatedEnvironment(workspace, kimiHome)
	output, err = command.Output()
	if err != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return output, err
}

func isolatedEnvironment(workspace, kimiHome string) []string {
	environment := inheritedEnvironment(
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
	)
	return append(environment,
		"APPDATA="+workspace,
		"HOME="+workspace,
		"KIMI_CODE_HOME="+kimiHome,
		"KIMI_DISABLE_TELEMETRY=1",
		"KIMI_LOG_LEVEL=off",
		"LOCALAPPDATA="+workspace,
		"TEMP="+workspace,
		"TMP="+workspace,
		"TMPDIR="+workspace,
		"USERPROFILE="+workspace,
	)
}

func authenticationRequired(stderr string) bool {
	message := strings.ToLower(stderr)
	return strings.Contains(message, "no model configured") ||
		strings.Contains(message, "use /login to sign in") ||
		strings.Contains(message, "authentication required") ||
		strings.Contains(message, "not authenticated")
}

func finalAssistantResponse(output io.Reader, expectedVersion string) (string, error) {
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var finalResponse string
	versionSeen := false
	for scanner.Scan() {
		var event streamEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return "", fmt.Errorf("decode Kimi Code CLI event: %w", err)
		}
		switch event.Role {
		case "meta":
			if event.Type == "system.version" {
				if event.Version != expectedVersion {
					return "", fmt.Errorf("validate Kimi Code CLI event: version changed from %s to %s", expectedVersion, event.Version)
				}
				versionSeen = true
			}
		case "assistant":
			if len(event.ToolCalls) != 0 {
				return "", errors.New("validate Kimi Code CLI event: reviewer attempted a tool call")
			}
			if strings.TrimSpace(event.Content) != "" {
				finalResponse = event.Content
			}
		case "tool":
			return "", errors.New("validate Kimi Code CLI event: reviewer produced a tool result")
		default:
			return "", fmt.Errorf("validate Kimi Code CLI event: unexpected role %q", event.Role)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read Kimi Code CLI events: %w", err)
	}
	if !versionSeen {
		return "", errors.New("validate Kimi Code CLI events: version metadata is missing")
	}
	if strings.TrimSpace(finalResponse) == "" {
		return "", errors.New("validate Kimi Code CLI events: final assistant response is missing")
	}
	return finalResponse, nil
}

func reviewPrompt(plan []byte) string {
	return `Review only the bounded plan below. Do not infer or use any prior conversation, hidden project context, or another reviewer's result.

Return exactly one JSON object with this shape and no Markdown:
{"schema_version":"1","findings":[{"severity":"blocking|major|minor","title":"...","evidence":"...","impact":"...","suggested_action":"..."}]}

Use an empty findings array when there are no findings. Base every finding on evidence in the plan.

<plan>
` + string(plan) + `
</plan>`
}

type streamEvent struct {
	Role      string          `json:"role"`
	Type      string          `json:"type"`
	Version   string          `json:"version"`
	Content   string          `json:"content"`
	ToolCalls json.RawMessage `json:"tool_calls"`
}

const agentDefinition = `---
name: planlens-reviewer
description: Independently review one bounded plan without tools or subagents.
tools: []
subagents: []
---

You are an independent plan reviewer. Inspect only the material supplied in the current user prompt. You have no tools and must not request file, shell, network, browser, skill, memory, or subagent access. Return only the requested final JSON object.
`
