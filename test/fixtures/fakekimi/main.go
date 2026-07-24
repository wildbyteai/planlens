package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const expectedPlanDigest = "7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429"

func main() {
	kimiHome := verifyIsolation()
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		blockIfRequested(kimiHome, "version")
		fmt.Println("0.29.1")
		return
	}

	if os.Getenv("KIMI_CODE_EXPERIMENTAL_FLAG") != "1" {
		fail("experimental agent-file support was not enabled")
	}
	if os.Getenv("KIMI_DISABLE_TELEMETRY") != "1" {
		fail("telemetry was not disabled")
	}
	if hasArgument("--auto") || hasArgument("--yolo") || hasArgument("--continue") || hasArgument("--session") {
		fail("reviewer was started with an unsafe or stateful option")
	}
	if valueFor("--output-format") != "stream-json" {
		fail("reviewer did not request machine-readable final output")
	}
	if kimiHome == "" {
		fail("reviewer did not receive the dedicated Kimi home")
	}
	blockAfterAuthenticationDiagnosticIfRequested(kimiHome)

	workingDirectory, err := os.Getwd()
	if err != nil {
		fail("read working directory: " + err.Error())
	}
	agentPath := valueFor("--agent-file")
	skillsPath := valueFor("--skills-dir")
	if !isWithin(workingDirectory, agentPath) || !isWithin(workingDirectory, skillsPath) {
		fail("agent or skills path escaped the temporary review workspace")
	}
	agent, err := os.ReadFile(agentPath)
	if err != nil {
		fail("read agent file: " + err.Error())
	}
	agentText := string(agent)
	if !strings.Contains(agentText, "tools: []") || !strings.Contains(agentText, "subagents: []") {
		fail("agent file did not disable tools and subagents")
	}
	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		fail("read skills directory: " + err.Error())
	}
	if len(entries) != 0 {
		fail("reviewer received discovered skills")
	}

	prompt := valueFor("--prompt")
	if prompt == "" {
		prompt = valueFor("-p")
	}
	plan := between(prompt, "<plan>\n", "\n</plan>")
	if plan == "" {
		fail("reviewer did not receive one bounded plan block")
	}
	normalizedPlan := strings.ReplaceAll(plan, "\r\n", "\n")
	planDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(normalizedPlan)))
	if planDigest != expectedPlanDigest {
		fail("reviewer did not receive the complete public plan fixture")
	}
	if strings.Contains(prompt, "Reviewer: simulated") || strings.Count(prompt, "<plan>") != 1 {
		fail("reviewer received prior review state or duplicate material")
	}

	fmt.Println(`{"role":"meta","type":"system.version","version":"0.29.1"}`)
	fmt.Println(`{"role":"assistant","content":"{\"schema_version\":\"1\",\"findings\":[{\"severity\":\"major\",\"title\":\"The rollout has no rollback decision\",\"evidence\":\"The plan schedules deployment but defines no rollback trigger or owner.\",\"impact\":\"A failed deployment could remain active while responsibility is unclear.\",\"suggested_action\":\"Add a rollback trigger, named owner, and recovery verification step.\"}]}"}`)
	fmt.Println(`{"role":"meta","type":"session.resume_hint","session_id":"discard-me","command":"kimi -S discard-me","content":"discard this metadata"}`)
}

func verifyIsolation() string {
	if os.Getenv("PLANLENS_TEST_SECRET") != "" {
		fail("reviewer inherited an unrelated parent environment value")
	}
	if os.Getenv("KIMI_LOG_LEVEL") != "off" {
		fail("KIMI_LOG_LEVEL did not disable ordinary Kimi logging")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		fail("read working directory: " + err.Error())
	}
	for _, name := range []string{
		"HOME",
		"USERPROFILE",
		"APPDATA",
		"LOCALAPPDATA",
		"TEMP",
		"TMP",
		"TMPDIR",
	} {
		value := os.Getenv(name)
		if value == "" || !isWithinOrSame(workingDirectory, value) {
			fail(name + " was not isolated inside the temporary workspace")
		}
	}
	kimiHome := os.Getenv("KIMI_CODE_HOME")
	info, err := os.Stat(kimiHome)
	if kimiHome == "" || err != nil || !info.IsDir() {
		fail("KIMI_CODE_HOME did not name the dedicated existing directory")
	}
	return kimiHome
}

func blockIfRequested(kimiHome, phase string) {
	if _, err := os.Stat(filepath.Join(kimiHome, "fake-kimi-block-"+phase)); err != nil {
		return
	}
	if err := os.WriteFile(filepath.Join(kimiHome, "fake-kimi-"+phase+"-started"), []byte("yes\n"), 0o600); err != nil {
		fail("record " + phase + " start: " + err.Error())
	}
	time.Sleep(30 * time.Second)
}

func blockAfterAuthenticationDiagnosticIfRequested(kimiHome string) {
	if _, err := os.Stat(filepath.Join(kimiHome, "fake-kimi-block-after-auth-diagnostic")); err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, "authentication required: private diagnostic must be discarded")
	if err := os.WriteFile(filepath.Join(kimiHome, "fake-kimi-auth-diagnostic-started"), []byte("yes\n"), 0o600); err != nil {
		fail("record authentication diagnostic start: " + err.Error())
	}
	time.Sleep(30 * time.Second)
}

func hasArgument(want string) bool {
	for _, argument := range os.Args[1:] {
		if argument == want {
			return true
		}
	}
	return false
}

func valueFor(flag string) string {
	for index, argument := range os.Args[1:] {
		if argument == flag && index+2 <= len(os.Args[1:]) {
			return os.Args[index+2]
		}
	}
	return ""
}

func isWithin(directory, path string) bool {
	directory, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return false
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(directory, path)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func isWithinOrSame(directory, path string) bool {
	directory, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return false
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(directory, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func between(value, start, end string) string {
	startIndex := strings.Index(value, start)
	if startIndex < 0 {
		return ""
	}
	remaining := value[startIndex+len(start):]
	endIndex := strings.Index(remaining, end)
	if endIndex < 0 {
		return ""
	}
	return remaining[:endIndex]
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
