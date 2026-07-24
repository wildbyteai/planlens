package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const defaultVersion = "0.146.0-alpha.3.1"

const defaultRootHelp = `Codex CLI
  --ask-for-approval <APPROVAL_POLICY>
`

const defaultExecHelp = `Run Codex non-interactively
  --config <key=value>
  --strict-config
  --disable <FEATURE>
  --sandbox <SANDBOX_MODE>
  --cd <DIR>
  --skip-git-repo-check
  --ephemeral
  --ignore-user-config
  --ignore-rules
  --output-schema <FILE>
  --json
  --output-last-message <FILE>
  --color <COLOR>
`

const defaultFinalMessage = `{"schema_version":"1","findings":[]}`

const capabilityCanary = "PLANLENS_PUBLIC_AGENTS_CANARY_MUST_NOT_REACH_MODEL"

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

func main() {
	home := fakeControlHome()
	args := os.Args[1:]
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatal("read stdin")
	}
	cwd, err := os.Getwd()
	if err != nil {
		fatal("read cwd")
	}

	record := invocation{
		Args:        args,
		CWD:         cwd,
		Stdin:       string(input),
		Home:        os.Getenv("HOME"),
		UserProfile: os.Getenv("USERPROFILE"),
		CodexHome:   os.Getenv("CODEX_HOME"),
		Temp:        os.Getenv("TEMP"),
		Tmp:         os.Getenv("TMP"),
		TmpDir:      os.Getenv("TMPDIR"),
	}
	if schemaPath := valueAfter(args, "--output-schema"); schemaPath != "" {
		schema, err := os.ReadFile(schemaPath)
		if err != nil {
			fatal("read output schema")
		}
		record.OutputSchema = string(schema)
	}
	appendInvocation(filepath.Join(home, "fake-codex-invocations.jsonl"), record)

	if os.Getenv("PLANLENS_TEST_SECRET") != "" {
		if err := os.WriteFile(filepath.Join(home, "fake-codex-secret-leaked"), []byte("leaked"), 0o600); err != nil {
			fatal("record leaked secret")
		}
	}

	switch {
	case slices.Equal(args, []string{"--version"}):
		fmt.Printf("codex-cli %s\n", readOptional(filepath.Join(home, "fake-codex-version.txt"), defaultVersion))
	case slices.Equal(args, []string{"--help"}):
		if _, err := os.Stat(filepath.Join(home, "fake-codex-block-root-help")); err == nil {
			if err := os.WriteFile(filepath.Join(home, "fake-codex-root-help-started"), []byte("started"), 0o600); err != nil {
				fatal("record root help start")
			}
			time.Sleep(time.Hour)
		}
		fmt.Print(readOptional(filepath.Join(home, "fake-codex-root-help.txt"), defaultRootHelp))
	case slices.Equal(args, []string{"exec", "--help"}):
		fmt.Print(readOptional(filepath.Join(home, "fake-codex-exec-help.txt"), defaultExecHelp))
	case isFeatureList(args):
		printFeatureList(home, args)
	case isPromptInputProbe(args):
		printPromptInputProbe(home, cwd, args)
	case slices.Equal(args, []string{"login", "status"}):
		if _, err := os.Stat(filepath.Join(home, "fake-codex-authenticated")); err != nil {
			fmt.Fprintln(os.Stderr, "not logged in: owner@example.com secret-token")
			os.Exit(1)
		}
	default:
		runReview(home, args)
	}
}

func isFeatureList(args []string) bool {
	return len(args) >= 2 && slices.Equal(args[len(args)-2:], []string{"features", "list"})
}

func printFeatureList(home string, args []string) {
	forcedEnabled := readOptional(filepath.Join(home, "fake-codex-enabled-feature.txt"), "")
	for _, feature := range valuesAfter(args, "--disable") {
		enabled := "false"
		if feature == forcedEnabled {
			enabled = "true"
		}
		fmt.Printf("%s stable %s\n", feature, enabled)
	}
}

func isPromptInputProbe(args []string) bool {
	return len(args) >= 2 && slices.Equal(args[:2], []string{"debug", "prompt-input"})
}

func printPromptInputProbe(home, cwd string, args []string) {
	if !sameDirectory(home, cwd) ||
		!sameDirectory(os.Getenv("HOME"), cwd) ||
		!sameDirectory(os.Getenv("USERPROFILE"), cwd) ||
		!sameDirectory(os.Getenv("CODEX_HOME"), filepath.Join(cwd, ".codex")) {
		fatal("capability probe was not isolated from the user Codex home")
	}
	for _, key := range []string{"TEMP", "TMP", "TMPDIR"} {
		if !sameDirectory(os.Getenv(key), cwd) {
			fatal("capability probe temporary directory was not isolated")
		}
	}
	configs := valuesAfter(args, "--config")
	if !slices.Contains(configs, "project_doc_max_bytes=0") || !slices.Contains(configs, `web_search="disabled"`) {
		fatal("required config isolation controls are missing")
	}
	prompt := args[len(args)-1]
	executable, err := os.Executable()
	if err != nil {
		fatal("resolve fake executable")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(executable), "fake-codex-leak-capability-canary")); err == nil {
		fmt.Printf(`[{"role":"user","content":%q},{"role":"developer","content":%q}]`, prompt, capabilityCanary)
		return
	}
	projectInstructions, projectErr := os.ReadFile(filepath.Join(cwd, "AGENTS.md"))
	globalInstructionsAbsent := true
	for _, name := range []string{"AGENTS.override.md", "AGENTS.md"} {
		if _, err := os.Lstat(filepath.Join(os.Getenv("CODEX_HOME"), name)); !os.IsNotExist(err) {
			globalInstructionsAbsent = false
		}
	}
	if projectErr == nil && globalInstructionsAbsent &&
		strings.Contains(string(projectInstructions), capabilityCanary) {
		fmt.Printf(`[{"role":"user","content":%q}]`, prompt)
		return
	}
	fatal("capability probe did not isolate the project AGENTS canary")
}

func fakeControlHome() string {
	if codexHome := os.Getenv("CODEX_HOME"); codexHome != "" {
		return filepath.Dir(codexHome)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fatal("resolve fake home")
	}
	return home
}

func runReview(home string, args []string) {
	if _, err := os.Stat(filepath.Join(home, "fake-codex-run-fail")); err == nil {
		fmt.Fprintln(os.Stderr, "review failed for owner@example.com with secret-token")
		os.Exit(1)
	}
	outputPath := valueAfter(args, "--output-last-message")
	if outputPath == "" {
		fatal("missing output-last-message")
	}
	finalMessage := readOptional(filepath.Join(home, "fake-codex-final.json"), defaultFinalMessage)
	if err := os.WriteFile(outputPath, []byte(finalMessage), 0o600); err != nil {
		fatal("write final message")
	}
	fmt.Println(`{"type":"item.completed","item":{"type":"reasoning","text":"private reasoning marker","schema_version":"999"}}`)
}

func appendInvocation(path string, value invocation) {
	encoded, err := json.Marshal(value)
	if err != nil {
		fatal("encode invocation")
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		fatal("open invocation log")
	}
	defer file.Close()
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		fatal("write invocation log")
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

func readOptional(path, fallback string) string {
	contents, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fallback
		}
		fatal("read " + filepath.Base(path))
	}
	return strings.TrimSpace(string(contents))
}

func sameDirectory(left, right string) bool {
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo)
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
