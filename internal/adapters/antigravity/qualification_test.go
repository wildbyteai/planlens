package antigravity_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wildbyteai/planlens/internal/adapters/antigravity"
)

func TestFeasibilityRecordIsSanitizedAndConditional(t *testing.T) {
	checkedAt := time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC)
	record, err := antigravity.BuildFeasibilityRecord(
		checkedAt,
		"darwin",
		"arm64",
		[]byte(publicPlanFixture),
		antigravity.ReviewResult{
			Reviewer:         antigravity.ReviewerID,
			CLIVersion:       "1.1.6",
			Authentication:   antigravity.AuthenticationAvailable,
			AccessCapability: antigravity.AccessConstrained,
			FinalResponse:    expectedFinalResponse,
		},
	)
	if err != nil {
		t.Fatalf("build feasibility record: %v", err)
	}
	if record.Status != antigravity.FeasibilityPassedConditional {
		t.Fatalf("unexpected status: %q", record.Status)
	}
	if record.Stability != antigravity.StabilityNotStable {
		t.Fatalf("unexpected stability: %q", record.Stability)
	}
	if record.TermsReviewNotBefore != "2026-07-30" {
		t.Fatalf("unexpected terms review date: %q", record.TermsReviewNotBefore)
	}
	if record.MaterialTransport != antigravity.MaterialTransportPublicArgument || record.SensitiveMaterialsSupported {
		t.Fatalf(
			"unexpected material transport boundary: transport=%q sensitive=%t",
			record.MaterialTransport,
			record.SensitiveMaterialsSupported,
		)
	}
	if !record.FinalResponse.Captured || !record.FinalResponse.ExpectedMarkersPresent || record.FinalResponse.SHA256 == "" {
		t.Fatalf("incomplete final response evidence: %#v", record.FinalResponse)
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(encoded)
	for _, forbidden := range []string{
		expectedFinalResponse,
		"PRIVATE_REASONING",
		"PRIVATE_AUTH_DETAIL",
		"/Users/",
	} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("feasibility record contains forbidden detail %q", forbidden)
		}
	}
}

func TestFeasibilityRecordAcceptsMarkdownFormattingAroundRequiredLabels(t *testing.T) {
	formattedFinal := `PLANLENS_ANTIGRAVITY_FEASIBILITY_OK
**Finding:** The rollout has no rollback decision.
**Severity:** major
**Evidence:** The plan deploys to all users without a rollback trigger.
**Impact:** Recovery ownership is unclear.
**Suggested action:** Add a rollback trigger and owner.
**Fixture marker:** ` + "`planlens-public-fixture-end`"

	_, err := antigravity.BuildFeasibilityRecord(
		time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC),
		"darwin",
		"arm64",
		[]byte(publicPlanFixture),
		antigravity.ReviewResult{
			Reviewer:         antigravity.ReviewerID,
			CLIVersion:       "1.1.6",
			Authentication:   antigravity.AuthenticationAvailable,
			AccessCapability: antigravity.AccessConstrained,
			FinalResponse:    formattedFinal,
		},
	)
	if err != nil {
		t.Fatalf("accept Markdown-formatted required labels: %v", err)
	}
}

func TestFeasibilityRecordRejectsChangedFixture(t *testing.T) {
	_, err := antigravity.BuildFeasibilityRecord(
		time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC),
		"darwin",
		"arm64",
		[]byte("changed fixture"),
		antigravity.ReviewResult{
			Reviewer:         antigravity.ReviewerID,
			CLIVersion:       "1.1.6",
			Authentication:   antigravity.AuthenticationAvailable,
			AccessCapability: antigravity.AccessConstrained,
			FinalResponse:    expectedFinalResponse,
		},
	)
	if err == nil || !strings.Contains(err.Error(), "fixed public fixture") {
		t.Fatalf("expected fixed-fixture error, got %v", err)
	}
}

func TestFeasibilityRecordRejectsFinalResponseThatCannotBeNormalized(t *testing.T) {
	invalidFinal := `PLANLENS_ANTIGRAVITY_FEASIBILITY_OK
Finding: The rollout has no rollback decision.
Severity: critical
Evidence: The plan deploys to all users without a rollback trigger.
Impact: Recovery ownership is unclear.
Suggested action: Add a rollback trigger and owner.
Fixture marker: planlens-public-fixture-end`

	_, err := antigravity.BuildFeasibilityRecord(
		time.Date(2026, time.July, 24, 16, 0, 0, 0, time.UTC),
		"darwin",
		"arm64",
		[]byte(publicPlanFixture),
		antigravity.ReviewResult{
			Reviewer:         antigravity.ReviewerID,
			CLIVersion:       "1.1.6",
			Authentication:   antigravity.AuthenticationAvailable,
			AccessCapability: antigravity.AccessConstrained,
			FinalResponse:    invalidFinal,
		},
	)
	if err == nil || !strings.Contains(err.Error(), "cannot be normalized") {
		t.Fatalf("expected normalization error, got %v", err)
	}
}
