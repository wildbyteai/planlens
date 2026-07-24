package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/wildbyteai/planlens/internal/review"
)

const ID = "claude"

const reviewPrompt = `Review the plan supplied on standard input as an independent plan reviewer. Return only the structured result required by the JSON schema. Identify concrete issues with evidence, impact, and a suggested action. Do not execute tools, modify files, or request additional access. A sound plan may have zero findings.`

const resultSchema = `{"type":"object","additionalProperties":false,"properties":{"schema_version":{"const":"1"},"findings":{"type":"array","maxItems":100,"items":{"type":"object","additionalProperties":false,"properties":{"severity":{"type":"string","enum":["blocking","major","minor"]},"title":{"type":"string","minLength":1,"maxLength":500},"evidence":{"type":"string","minLength":1,"maxLength":4000},"impact":{"type":"string","minLength":1,"maxLength":4000},"suggested_action":{"type":"string","minLength":1,"maxLength":4000}},"required":["severity","title","evidence","impact","suggested_action"]}}},"required":["schema_version","findings"]}`

var ErrAuthenticationRequired = errors.New("reviewer authentication required")

type reviewer struct {
	executable string
}

func New() (review.Reviewer, error) {
	executable, err := exec.LookPath("claude")
	if err != nil {
		return nil, fmt.Errorf("discover Claude Code CLI: %w", err)
	}
	return reviewer{executable: executable}, nil
}

func (reviewer) ID() string {
	return ID
}

func (adapter reviewer) Run(ctx context.Context, request review.Request) (result review.Result, runErr error) {
	workDir, err := os.MkdirTemp("", "planlens-claude-review-*")
	if err != nil {
		return review.Result{}, fmt.Errorf("create Claude reviewer workspace: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			result = review.Result{}
			runErr = errors.Join(runErr, fmt.Errorf("remove Claude reviewer workspace: %w", err))
		}
	}()

	versionOutput, err := adapter.output(ctx, workDir, "--version")
	if err != nil {
		return review.Result{}, fmt.Errorf("detect Claude Code version: %w", err)
	}
	version, err := parseVersion(versionOutput)
	if err != nil {
		return review.Result{}, err
	}

	authOutput, err := adapter.output(ctx, workDir, "auth", "status", "--json")
	if err != nil {
		return review.Result{}, fmt.Errorf("detect Claude Code authentication: %w", err)
	}
	var authStatus struct {
		LoggedIn bool `json:"loggedIn"`
	}
	if err := json.Unmarshal(authOutput, &authStatus); err != nil {
		return review.Result{}, fmt.Errorf("decode Claude Code authentication status: %w", err)
	}
	if !authStatus.LoggedIn {
		return review.Result{}, ErrAuthenticationRequired
	}

	command := exec.CommandContext(ctx, adapter.executable, reviewArgs()...)
	command.Dir = workDir
	command.Env = environment()
	command.Stdin = bytes.NewReader(request.Plan)

	var stdout bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		return review.Result{}, fmt.Errorf("Claude Code review process failed: %w", err)
	}

	var response struct {
		IsError          bool            `json:"is_error"`
		StructuredOutput json.RawMessage `json:"structured_output"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return review.Result{}, fmt.Errorf("decode Claude Code final response: %w", err)
	}
	if response.IsError {
		return review.Result{}, fmt.Errorf("Claude Code reported a review failure")
	}
	if len(response.StructuredOutput) == 0 || string(response.StructuredOutput) == "null" {
		return review.Result{}, fmt.Errorf("Claude Code final response did not include structured_output")
	}

	result, err = review.DecodeResult(bytes.NewReader(response.StructuredOutput))
	if err != nil {
		return review.Result{}, err
	}
	result.Metadata = review.Metadata{
		CLIVersion:       version,
		AccessCapability: review.AccessConstrained,
	}
	return result, nil
}

func (adapter reviewer) output(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, adapter.executable, args...)
	command.Dir = workDir
	command.Env = environment()
	command.Stderr = io.Discard
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func parseVersion(output []byte) (string, error) {
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return "", fmt.Errorf("Claude Code version output was empty")
	}
	parts := strings.Split(fields[0], ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("unrecognized Claude Code version %q", fields[0])
	}
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return "", fmt.Errorf("unrecognized Claude Code version %q", fields[0])
		}
	}
	return fields[0], nil
}

func reviewArgs() []string {
	return []string{
		"--print",
		"--safe-mode",
		"--no-session-persistence",
		"--no-chrome",
		"--disable-slash-commands",
		"--permission-mode", "dontAsk",
		"--tools", "",
		"--output-format", "json",
		"--json-schema", resultSchema,
		reviewPrompt,
	}
}

func environment() []string {
	allowed := map[string]bool{
		"CLAUDE_CONFIG_DIR": true,
		"HOME":              true,
		"LANG":              true,
		"LC_ALL":            true,
		"LC_CTYPE":          true,
		"LOGNAME":           true,
		"PATH":              true,
		"SHELL":             true,
		"TEMP":              true,
		"TMP":               true,
		"TMPDIR":            true,
		"USER":              true,
		"XDG_CONFIG_HOME":   true,
	}

	var filtered []string
	for _, entry := range os.Environ() {
		name, _, ok := strings.Cut(entry, "=")
		if ok && allowed[name] {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
