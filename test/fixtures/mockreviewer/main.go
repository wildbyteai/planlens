package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

const expectedPlanDigest = "7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429"

func main() {
	if os.Getenv("PLANLENS_TEST_SECRET") != "" {
		fmt.Fprintln(os.Stderr, "reviewer inherited an unrelated parent environment value")
		os.Exit(1)
	}

	plan, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read plan:", err)
		os.Exit(1)
	}

	normalizedPlan := strings.ReplaceAll(string(plan), "\r\n", "\n")
	planDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(normalizedPlan)))
	if planDigest != expectedPlanDigest {
		fmt.Fprintln(os.Stderr, "reviewer did not receive the complete public plan fixture")
		os.Exit(1)
	}

	fmt.Print(`{"schema_version":"1","findings":[{"severity":"major","title":"The rollout has no rollback decision","evidence":"The plan schedules deployment but defines no rollback trigger or owner.","impact":"A failed deployment could remain active while responsibility is unclear.","suggested_action":"Add a rollback trigger, named owner, and recovery verification step."}]}`)
}
