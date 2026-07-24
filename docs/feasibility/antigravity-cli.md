# Antigravity CLI feasibility

## Status

- Technical feasibility: `passed_conditional`
- Stable adapter status: `not_stable`
- Integration: independent community subprocess integration; not endorsed by Google
- Provider-terms gate: not performed; it may be performed on or after `2026-07-30`

This record documents technical feasibility through the locally installed official `agy` CLI. It does not determine whether Google permits a third-party shell wrapper, does not qualify repeated reliability, and must not be used as a stable compatibility claim.

## Sanitized qualification evidence

| Field | Result |
|---|---|
| Checked at | `2026-07-24T19:16:57Z` |
| CLI | `agy` |
| CLI version | `1.1.6` |
| Platform | `darwin/arm64` |
| Public fixture SHA-256 | `7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429` |
| Command discovery | Passed for the fixed command, exact tested version `1.1.6`, and required native controls |
| Authentication | Available, inferred only from a successful public-fixture review; PlanLens did not directly inspect or export credentials |
| Fresh non-interactive process | Passed with `--print`; no continue or conversation option was used |
| Permission controls | `--mode plan` and `--sandbox` |
| Access capability | `constrained` |
| Final response | Captured; required finding and fixture markers were present |
| Final response SHA-256 | `d2e0e3348e2edf4b1689227f710dfcfd01d7fa280be30f3488d4002eb3df0345` |
| Intermediate reasoning | Not retained by PlanLens; only print-mode stdout was treated as final and stderr was discarded |

PlanLens did not directly inspect or export credentials. This sanitized record contains no account identifier, credential, private transcript, CLI log, model identity, or machine-specific path.

## Verified boundary

The current official interface requires the prompt as the value of `--print`. The successful qualification therefore embedded only the fixed public fixture in the process argument.

- Sensitive or private materials are not supported by this feasibility adapter.
- Only the exact tested CLI version `1.1.6` is accepted; other versions are not treated as compatible automatically.
- The adapter rejects any material whose SHA-256 does not exactly match the fixed public fixture.
- A file-based attempt was not accepted as qualification because read-only file access could enter an interactive permission flow.
- `--dangerously-skip-permissions` was not used and is not part of the adapter.
- `plan` mode restricts the Agent to read-only tools, and the terminal sandbox constrains terminal commands, but this test did not prove a general cross-platform operating-system boundary. The access level is therefore `constrained`, not `enforced`.

The controlled run on `2026-07-24` established command discovery, existing authentication availability, fresh headless execution, complete public-fixture delivery, and final-response capture on `darwin/arm64`. It did not qualify repeated reliability, other platforms, provider authorization, or stable support.

## Authentication and retention

PlanLens does not implement Antigravity login or directly inspect or export credentials. A successful review is the only evidence recorded for existing authentication availability. Authentication failure output is classified in memory and is not returned as a transcript.

The official CLI may maintain its own product logs or history under its own behavior and policies. This adapter redirects its requested CLI log into the temporary review workspace and removes that workspace after the process exits, but it does not claim control over every record the official CLI may maintain independently.

## Technical command sources consulted

The following Google documentation was consulted on `2026-07-24` only to define and verify the technical command invocation:

- [Antigravity CLI overview](https://antigravity.google/docs/cli)
- [Using AGY CLI](https://antigravity.google/docs/cli/features)
- [CLI reference](https://antigravity.google/docs/cli/reference)
- [Execution modes](https://antigravity.google/docs/cli/execution-modes)
- [Terminal sandboxing](https://antigravity.google/docs/cli/sandbox)

The Google and Antigravity terms review required by the release gate has not been performed. It must not be performed before `2026-07-30`, and this record makes no conclusion about provider permission. Until that later review and the remaining qualification gates are complete, this adapter cannot be presented as stable.
