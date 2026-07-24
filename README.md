# PlanLens

PlanLens is an installable Agent skill that asks one or more local AI CLIs to independently review the same bounded plan, then consolidates their findings while preserving evidence, sources, and disagreements.

## Status

**Pre-release — design and compatibility validation are in progress.**

There is no stable installation yet. The first public release will be `v1.0.0` after every required reviewer and platform combination passes compatibility, safety, and provider-usage checks.

## Target integrations

PlanLens is being designed for these host Agent environments:

- Codex
- Claude Code
- Antigravity

The planned v1 reviewer adapters are:

- Claude Code CLI — `claude`
- OpenAI Codex CLI — `codex`
- Antigravity CLI — `agy`
- Kimi Code CLI — `kimi`

These are target integrations, not current compatibility claims. Stable support will be documented with tested CLI versions, platforms, and validation dates.

## How it works

1. The host Agent prepares a bounded plan and an explicit material list.
2. The plan owner selects a review type and one or more reviewers.
3. Every reviewer starts in a fresh process and receives the same frozen materials.
4. PlanLens normalizes the final responses and consolidates findings by issue and evidence.
5. The plan owner accepts, rejects, defers, or clarifies the findings.
6. Another review round runs only after an explicit new invocation.

Users invoke the installed skill as `$planlens` or `/planlens`; they do not start an internal service or run the Go program manually.

## Core principles

- **Local CLI orchestration.** PlanLens invokes CLIs already installed and authenticated by the user. It does not call model APIs, accept API keys, or run a hosted relay.
- **Bounded materials.** Reviewers receive only the plan and supporting files explicitly selected for the round.
- **Independent reviews.** Reviewers cannot see another reviewer's findings from the same round.
- **Equivalent input.** Every reviewer in a round receives the same frozen review materials and review profile.
- **Review only.** Reviewers evaluate a plan; they do not modify its source files or make decisions for the user.
- **No model voting.** Findings are consolidated by root cause and evidence, not by majority vote or averaged scores.
- **Owner-controlled decisions.** Only the plan owner can declare a review converged or abandoned.
- **Traceability.** Reports retain reviewer identity, source findings, material hashes, disagreements, and run status.
- **No hidden iteration.** One invocation performs one review round. Retries, reviewer changes, upgrades, and later rounds are explicit.
- **Credential boundaries.** PlanLens checks whether a CLI is available and authenticated but does not read, copy, export, or store its credentials.

## Security boundary

PlanLens uses a temporary material workspace, a reduced subprocess environment, and supported CLI permission controls where available.

These measures reduce unintended access, but they are not an operating-system sandbox and do not prove that a third-party CLI cannot access files outside the review workspace. Sensitive materials must remain subject to the user's organization policies and each provider's data-handling terms.

Only final reviewer responses and necessary operational metadata are retained. Intermediate reasoning events are discarded rather than stored or included in reports.

## Target platforms for v1

Planned prebuilt releases:

- macOS on Apple silicon — `darwin/arm64`
- macOS on Intel — `darwin/amd64`
- Windows x64 — `windows/amd64`

Linux and Windows ARM64 are outside the v1 support scope. Unsupported platforms will not be silently mapped to another build.

## Third-party CLI boundaries

Third-party CLIs are not bundled with PlanLens. Users install and authenticate them through their official flows. PlanLens does not provide provider login, forward subscription credentials, use private service interfaces, or bypass account, quota, rate, region, permission, or safety controls.

The Antigravity adapter remains conditional because Google's current terms use broad language about third-party access. PlanLens distinguishes launching a user's authenticated official `agy` subprocess from directly using Antigravity credentials or private services, but Google has not endorsed that interpretation. The adapter may be paused or removed if provider guidance conflicts with it.

## License

PlanLens is licensed under the [Apache License 2.0](LICENSE).

## Disclaimer

PlanLens is an independent open-source project. It is not affiliated with, sponsored by, endorsed by, or an official integration of Anthropic, OpenAI, Google, Moonshot AI, or their products.

Product names and trademarks belong to their respective owners and are used only to identify compatible command-line tools. Third-party CLIs remain subject to their own licenses, subscriptions, authentication requirements, usage policies, service terms, and data practices.

AI-generated reviews may be incomplete or incorrect. PlanLens does not automatically approve a plan and does not replace human judgment, security review, legal advice, or professional assessment.
