package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const expectedPlanDigest = "7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429"

func main() {
	switch {
	case slices.Equal(os.Args[1:], []string{"--version"}):
		writeMarker("version")
		fmt.Println("2.1.218 (Claude Code)")
	case slices.Equal(os.Args[1:], []string{"auth", "status", "--json"}):
		requireMarker("version")
		writeMarker("auth")
		fmt.Print(`{"loggedIn":true,"authMethod":"oauth_token","email":"must-not-be-retained@example.invalid"}`)
	default:
		review()
	}
}

func review() {
	requireMarker("version")
	requireMarker("auth")

	if os.Getenv("PLANLENS_TEST_SECRET") != "" {
		fail("Claude inherited an unrelated parent environment value")
	}
	requireArgs(os.Args[1:])

	plan, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail("read plan: " + err.Error())
	}
	normalizedPlan := strings.ReplaceAll(string(plan), "\r\n", "\n")
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(normalizedPlan)))
	if digest != expectedPlanDigest {
		fail("Claude did not receive the complete public plan fixture")
	}

	fmt.Print(`{"type":"result","subtype":"success","is_error":false,"result":"must not be retained","session_id":"must-not-be-retained","structured_output":{"schema_version":"1","findings":[{"severity":"major","title":"The rollout has no rollback decision","evidence":"The plan schedules deployment but defines no rollback trigger or owner.","impact":"A failed deployment could remain active while responsibility is unclear.","suggested_action":"Add a rollback trigger, named owner, and recovery verification step."}]}}`)
}

func requireArgs(args []string) {
	requiredFlags := []string{
		"--print",
		"--safe-mode",
		"--no-session-persistence",
		"--no-chrome",
		"--disable-slash-commands",
		"--permission-mode",
		"dontAsk",
		"--tools",
		"",
		"--output-format",
		"json",
		"--json-schema",
	}
	for _, required := range requiredFlags {
		if !slices.Contains(args, required) {
			fail("missing required Claude argument " + fmt.Sprintf("%q", required))
		}
	}

	for _, forbidden := range []string{"--continue", "--resume", "--dangerously-skip-permissions", "--bg", "--background"} {
		if slices.Contains(args, forbidden) {
			fail("forbidden Claude argument " + forbidden)
		}
	}

	schemaIndex := slices.Index(args, "--json-schema")
	if schemaIndex < 0 || schemaIndex+1 >= len(args) {
		fail("missing Claude JSON schema")
	}
	var schema map[string]any
	if err := json.Unmarshal([]byte(args[schemaIndex+1]), &schema); err != nil {
		fail("invalid Claude JSON schema")
	}
}

func markerPath(name string) string {
	executable, err := os.Executable()
	if err != nil {
		fail("locate fake Claude executable")
	}
	return filepath.Join(filepath.Dir(executable), ".fake-claude-"+name)
}

func writeMarker(name string) {
	if err := os.WriteFile(markerPath(name), []byte("ok"), 0o600); err != nil {
		fail("write fake Claude marker")
	}
}

func requireMarker(name string) {
	if _, err := os.Stat(markerPath(name)); err != nil {
		fail("missing fake Claude " + name + " check")
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
