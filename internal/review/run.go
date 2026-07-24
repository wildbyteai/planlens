package review

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const SimulatedReviewerID = "simulated"

type Result struct {
	SchemaVersion string    `json:"schema_version"`
	Findings      []Finding `json:"findings"`
	Metadata      Metadata  `json:"-"`
}

type Metadata struct {
	CLIVersion       string
	AccessCapability AccessCapability
}

type AccessCapability string

const AccessConstrained AccessCapability = "constrained"

type Finding struct {
	Severity        Severity `json:"severity"`
	Title           string   `json:"title"`
	Evidence        string   `json:"evidence"`
	Impact          string   `json:"impact"`
	SuggestedAction string   `json:"suggested_action"`
}

type Severity string

const (
	SeverityBlocking Severity = "blocking"
	SeverityMajor    Severity = "major"
	SeverityMinor    Severity = "minor"
)

type Request struct {
	Plan []byte
}

type Reviewer interface {
	ID() string
	Run(context.Context, Request) (Result, error)
}

type simulatedReviewer struct {
	executable string
}

func NewSimulatedReviewer() (Reviewer, error) {
	planlensExecutable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locate PlanLens executable: %w", err)
	}

	reviewerName := "planlens-simulated-reviewer"
	if runtime.GOOS == "windows" {
		reviewerName += ".exe"
	}
	return simulatedReviewer{
		executable: filepath.Join(filepath.Dir(planlensExecutable), reviewerName),
	}, nil
}

func (simulatedReviewer) ID() string {
	return SimulatedReviewerID
}

func (reviewer simulatedReviewer) Run(ctx context.Context, request Request) (Result, error) {
	workDir, err := os.MkdirTemp("", "planlens-review-*")
	if err != nil {
		return Result{}, fmt.Errorf("create reviewer workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	command := exec.CommandContext(ctx, reviewer.executable)
	command.Stdin = bytes.NewReader(request.Plan)
	command.Env = []string{}
	command.Dir = workDir

	var stdout bytes.Buffer
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		return Result{}, fmt.Errorf("reviewer process failed: %w", err)
	}

	return DecodeResult(&stdout)
}

func DecodeResult(input io.Reader) (Result, error) {
	var result Result
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return Result{}, fmt.Errorf("decode final response: %w", err)
	}
	if err := requireEndOfInput(decoder); err != nil {
		return Result{}, err
	}
	if result.SchemaVersion != "1" {
		return Result{}, fmt.Errorf("validate final response: unsupported schema_version %q", result.SchemaVersion)
	}
	for index, finding := range result.Findings {
		if err := validate(finding); err != nil {
			return Result{}, fmt.Errorf("validate final response: finding %d: %w", index, err)
		}
	}
	return result, nil
}

func requireEndOfInput(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("decode final response: multiple JSON values")
	}
	return fmt.Errorf("decode final response: %w", err)
}

func validate(finding Finding) error {
	switch finding.Severity {
	case SeverityBlocking, SeverityMajor, SeverityMinor:
	default:
		return fmt.Errorf("severity must be blocking, major, or minor")
	}

	fields := []struct {
		name  string
		value string
	}{
		{name: "title", value: finding.Title},
		{name: "evidence", value: finding.Evidence},
		{name: "impact", value: finding.Impact},
		{name: "suggested_action", value: finding.SuggestedAction},
	}
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}
	return nil
}
