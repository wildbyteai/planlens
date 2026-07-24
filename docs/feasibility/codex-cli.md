# Codex CLI feasibility record

- Status: blocked — active Codex home contains global instructions
- Support status: not qualified
- Qualification date: 2026-07-24
- Platform: `darwin/arm64`
- CLI version: `codex-cli 0.146.0-alpha.3.1`
- Public fixture SHA-256: `7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429`
- Access capability: `constrained`; the hardened candidate stopped before authentication and review

The published record contains no account identity, credentials, prompt transcript, reasoning events, or final model response.

## What was verified

- PlanLens detected the exact tested CLI version and rejected version or control mismatches before authentication.
- The exact CLI accepted `--strict-config`, `--ask-for-approval never`, `--sandbox read-only`, `--ephemeral`, `--ignore-user-config`, and `--ignore-rules`.
- `project_doc_max_bytes=0` suppressed a public project `AGENTS.md` canary, and `web_search="disabled"` disabled native web search.
- Stable features that can expose files, commands, browser or computer control, plugins, skills, MCP elicitation, or subagents were explicitly disabled, and the CLI reported each tested feature as disabled before authentication.
- The adapter checks the active `CODEX_HOME` root with file metadata only. If `AGENTS.md` or `AGENTS.override.md` exists, including as a symbolic link, it stops before authentication or review without reading, moving, deleting, or rewriting the file.
- The same fail-closed check runs again immediately before each review, preventing a global instruction file added after discovery from entering the model context.
- Review subprocess configuration isolates `HOME`, `USERPROFILE`, and temporary directories while preserving only the selected `CODEX_HOME` for official CLI authentication. PlanLens does not inspect, copy, link, export, or store credentials.

## Result

The hardened candidate returned `blocked` with the sanitized failure class `global_instructions_present`. The active Codex home on this machine contains a global instruction file, and the tested CLI has no verified native control that suppresses it while retaining the existing authentication state. PlanLens therefore stopped before authentication and did not send the public fixture to a model.

To continue qualification, the user must prepare a dedicated `CODEX_HOME` that contains neither global instruction filename and authenticate it through the official Codex login flow. PlanLens does not start that login flow or transfer credentials. The same immutable candidate and public fixture must then be rerun explicitly.

An earlier public-fixture run reached a final response before this stricter global-instruction gate was added. Because the adapter behavior changed, that result does not qualify the current candidate.

## Access boundary

The adapter remains classified as `constrained`, not `enforced`. Read-only mode, disabled features, an isolated process home, a project-instruction canary, and an instruction-free Codex home reduce unintended access, but they are not a material-only operating-system sandbox.

## Unresolved limitations

- Successful real final-response capture remains unverified for the hardened candidate.
- Qualification requires a separately authenticated, instruction-free Codex home prepared through the official CLI.
- Authentication is checked by exit status only; account identity and raw diagnostics are not retained.
- The controls do not establish the absence of provider-managed system context.
- This record covers one CLI version, platform, date, and fixed public fixture; it does not establish universal compatibility.
- Codex CLI controls may change in later releases, so each supported version still requires requalification.

## Official documentation

- [Non-interactive mode](https://developers.openai.com/codex/non-interactive-mode)
- [Authentication](https://developers.openai.com/codex/auth)
- [Command-line reference](https://developers.openai.com/codex/cli/reference)
- [Configuration reference](https://developers.openai.com/codex/config-reference)
- [AGENTS.md instructions](https://developers.openai.com/codex/guides/agents-md)
