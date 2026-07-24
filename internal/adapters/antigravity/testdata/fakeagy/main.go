package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const finalResponse = `PLANLENS_ANTIGRAVITY_FEASIBILITY_OK
Finding: The rollout has no rollback decision.
Severity: major
Evidence: The plan deploys to all users and defines no rollback trigger or owner.
Impact: A failed deployment could remain active without clear responsibility.
Suggested action: Add a rollback trigger, owner, and recovery verification.
Fixture marker: planlens-public-fixture-end`

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		version := "1.1.6"
		home := homeDirectory()
		blockIfRequested(home, "version")
		if override, err := os.ReadFile(filepath.Join(home, "fake-agy-version")); err == nil {
			version = strings.TrimSpace(string(override))
		}
		fmt.Println(version)
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		home := homeDirectory()
		blockIfRequested(home, "help")
		missing := ""
		if override, err := os.ReadFile(filepath.Join(home, "fake-agy-missing-control")); err == nil {
			missing = strings.TrimSpace(string(override))
		}
		var output io.Writer = os.Stdout
		if _, err := os.Stat(filepath.Join(home, "fake-agy-help-stderr")); err == nil {
			output = os.Stderr
		}
		fmt.Fprintln(output, "--model")
		for _, option := range []string{"--mode", "--sandbox", "--print", "--print-timeout", "--log-file"} {
			if option != missing {
				fmt.Fprintln(output, option)
			}
		}
		return
	}
	if os.Getenv("PLANLENS_TEST_SECRET") != "" {
		fail("reviewer inherited an unrelated environment value")
	}

	values, flags := parseArgs(os.Args[1:])
	if values["--mode"] != "plan" || !flags["--sandbox"] {
		fail("reviewer was not started in plan mode with the sandbox enabled")
	}
	if values["--print-timeout"] == "" || values["--log-file"] == "" || values["--print"] == "" {
		fail("required non-interactive arguments are missing")
	}
	for _, forbidden := range []string{"--continue", "--conversation", "--dangerously-skip-permissions", "--new-project", "--model", "--agent"} {
		if flags[forbidden] || values[forbidden] != "" {
			fail("fresh-session contract was violated")
		}
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		fail("cannot inspect reviewer working directory")
	}
	if !strings.HasPrefix(filepath.Base(workingDirectory), "planlens-antigravity-") {
		fail("reviewer did not receive a fresh workspace")
	}
	if !sameDirectory(filepath.Dir(values["--log-file"]), workingDirectory) {
		fail("CLI log was not redirected into the temporary workspace")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fail("cannot inspect fake home")
	}
	if _, err := os.Stat(filepath.Join(home, "fake-agy-auth-required")); err == nil {
		fmt.Fprintln(os.Stderr, "Authentication required: PRIVATE_AUTH_DETAIL")
		os.Exit(1)
	}
	if !strings.Contains(values["--print"], "Deploy the new parser to all users") ||
		!strings.Contains(values["--print"], "<!-- planlens-public-fixture-end -->") {
		fail("reviewer did not receive the complete public fixture")
	}
	if _, err := os.Stat(filepath.Join(workingDirectory, "plan.md")); !os.IsNotExist(err) {
		fail("public fixture should not require a file-read permission flow")
	}

	fmt.Fprintln(os.Stderr, "reasoning: PRIVATE_REASONING")
	fmt.Println(finalResponse)
}

func homeDirectory() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func blockIfRequested(home, probe string) {
	if _, err := os.Stat(filepath.Join(home, "fake-agy-block-"+probe)); err != nil {
		return
	}
	if err := os.WriteFile(filepath.Join(home, "fake-agy-"+probe+"-started"), []byte("yes"), 0o600); err != nil {
		fail("cannot record probe start")
	}
	time.Sleep(30 * time.Second)
}

func parseArgs(arguments []string) (map[string]string, map[string]bool) {
	values := map[string]string{}
	flags := map[string]bool{}
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--sandbox" {
			flags[argument] = true
			continue
		}
		if strings.HasPrefix(argument, "--") && index+1 < len(arguments) {
			values[argument] = arguments[index+1]
			index++
			continue
		}
		flags[argument] = true
	}
	return values, flags
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func sameDirectory(first, second string) bool {
	firstInfo, firstErr := os.Stat(first)
	secondInfo, secondErr := os.Stat(second)
	return firstErr == nil && secondErr == nil && os.SameFile(firstInfo, secondInfo)
}
