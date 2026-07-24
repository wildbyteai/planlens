package antigravity

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/wildbyteai/planlens/internal/review"
)

const (
	FeasibilityPassedConditional    = "passed_conditional"
	StabilityNotStable              = "not_stable"
	MaterialTransportPublicArgument = "public_fixture_in_print_argument"
)

type FeasibilityRecord struct {
	SchemaVersion               string                     `json:"schema_version"`
	CheckedAt                   string                     `json:"checked_at"`
	Reviewer                    string                     `json:"reviewer"`
	CLI                         string                     `json:"cli"`
	CLIVersion                  string                     `json:"cli_version"`
	Platform                    string                     `json:"platform"`
	Architecture                string                     `json:"architecture"`
	FixtureSHA256               string                     `json:"fixture_sha256"`
	MaterialTransport           string                     `json:"material_transport"`
	SensitiveMaterialsSupported bool                       `json:"sensitive_materials_supported"`
	CommandDiscovery            string                     `json:"command_discovery"`
	Authentication              AuthenticationStatus       `json:"authentication"`
	AuthenticationBasis         string                     `json:"authentication_basis"`
	FreshNonInteractive         bool                       `json:"fresh_non_interactive"`
	Permissions                 QualificationPermissions   `json:"permissions"`
	FinalResponse               QualificationFinalResponse `json:"final_response"`
	ReasoningRetention          string                     `json:"reasoning_retention"`
	Status                      string                     `json:"status"`
	Stability                   string                     `json:"stability"`
	TermsReviewNotBefore        string                     `json:"terms_review_not_before"`
	Integration                 string                     `json:"integration"`
	OfficialSources             []OfficialSource           `json:"official_sources"`
}

type QualificationPermissions struct {
	Mode             string                  `json:"mode"`
	Sandbox          bool                    `json:"sandbox"`
	AccessCapability review.AccessCapability `json:"access_capability"`
	Rationale        string                  `json:"rationale"`
}

type QualificationFinalResponse struct {
	Captured               bool   `json:"captured"`
	SHA256                 string `json:"sha256"`
	ExpectedMarkersPresent bool   `json:"expected_markers_present"`
}

type OfficialSource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func BuildFeasibilityRecord(
	checkedAt time.Time,
	platform string,
	architecture string,
	fixture []byte,
	result ReviewResult,
) (FeasibilityRecord, error) {
	if digest(fixture) != publicFixtureDigest {
		return FeasibilityRecord{}, fmt.Errorf("qualification only accepts the fixed public fixture")
	}
	if result.Reviewer != ReviewerID || result.CLIVersion != qualifiedVersion {
		return FeasibilityRecord{}, fmt.Errorf("incomplete Antigravity reviewer identity")
	}
	if result.Authentication != AuthenticationAvailable {
		return FeasibilityRecord{}, fmt.Errorf("Antigravity authentication was not demonstrated by a successful review")
	}
	if result.AccessCapability != AccessConstrained {
		return FeasibilityRecord{}, fmt.Errorf("unexpected Antigravity access capability %q", result.AccessCapability)
	}
	if _, err := normalizeFinalResponse(result.FinalResponse); err != nil {
		return FeasibilityRecord{}, fmt.Errorf("Antigravity final response cannot be normalized: %w", err)
	}

	return FeasibilityRecord{
		SchemaVersion:               "1",
		CheckedAt:                   checkedAt.UTC().Format(time.RFC3339),
		Reviewer:                    ReviewerID,
		CLI:                         "agy",
		CLIVersion:                  result.CLIVersion,
		Platform:                    platform,
		Architecture:                architecture,
		FixtureSHA256:               digest(fixture),
		MaterialTransport:           MaterialTransportPublicArgument,
		SensitiveMaterialsSupported: false,
		CommandDiscovery:            "passed",
		Authentication:              result.Authentication,
		AuthenticationBasis:         "successful public-fixture review through the existing official CLI session; PlanLens did not directly inspect or export credentials",
		FreshNonInteractive:         true,
		Permissions: QualificationPermissions{
			Mode:             "plan",
			Sandbox:          true,
			AccessCapability: result.AccessCapability,
			Rationale:        "plan mode is read-only and the terminal sandbox is enabled, but no cross-platform OS boundary was independently proven",
		},
		FinalResponse: QualificationFinalResponse{
			Captured:               true,
			SHA256:                 digest([]byte(result.FinalResponse)),
			ExpectedMarkersPresent: true,
		},
		ReasoningRetention:   "only documented print-mode stdout was treated as final; stderr was discarded and was not persisted",
		Status:               FeasibilityPassedConditional,
		Stability:            StabilityNotStable,
		TermsReviewNotBefore: "2026-07-30",
		Integration:          "independent community subprocess integration; not endorsed by Google",
		OfficialSources: []OfficialSource{
			{Title: "Antigravity CLI overview", URL: "https://antigravity.google/docs/cli"},
			{Title: "Using AGY CLI", URL: "https://antigravity.google/docs/cli/features"},
			{Title: "CLI reference", URL: "https://antigravity.google/docs/cli/reference"},
			{Title: "Execution modes", URL: "https://antigravity.google/docs/cli/execution-modes"},
			{Title: "Terminal sandboxing", URL: "https://antigravity.google/docs/cli/sandbox"},
		},
	}, nil
}

func missingQualificationMarkers(response string) []string {
	normalized := strings.ToLower(response)
	normalized = strings.NewReplacer("*", "", "`", "", "#", "").Replace(normalized)
	var missing []string
	for _, marker := range []string{
		"planlens_antigravity_feasibility_ok",
		"finding:",
		"severity:",
		"evidence:",
		"impact:",
		"suggested action:",
		"fixture marker:",
		"planlens-public-fixture-end",
	} {
		if !strings.Contains(normalized, marker) {
			missing = append(missing, marker)
		}
	}
	return missing
}

func digest(content []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(content))
}
