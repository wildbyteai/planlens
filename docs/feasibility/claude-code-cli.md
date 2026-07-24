# Claude Code CLI feasibility

- Status: passed
- Qualification date: 2026-07-24
- Platform: `darwin/arm64`
- CLI version: `2.1.218`
- Public fixture SHA-256: `7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429`
- Access capability: `constrained`

## What was verified

- PlanLens discovered the fixed `claude` command and obtained its version.
- Authentication availability was checked through the official `claude auth status --json` command. Account details were neither recorded nor displayed.
- The reviewer started as a fresh non-interactive `--print` process.
- Session persistence, Chrome integration, and slash commands were disabled for the review attempt.
- Claude Code safe mode was enabled, the built-in tool list was empty, and permission mode was `dontAsk`.
- The fixed public plan was supplied on standard input from a temporary review workspace.
- Claude Code returned a JSON-schema-constrained final response containing four findings.
- PlanLens normalized only `structured_output` into `review.Result`. It did not retain the outer response, session identifier, raw response text, or intermediate reasoning.

## Result

The real public-fixture qualification completed successfully. No account identifier, credential, private material, absolute filesystem path, or private transcript is included in this record.

## Access boundary

The adapter is classified as `constrained`, not `enforced`. A temporary workspace, reduced environment, safe mode, disabled tools, and non-persistent fresh invocation reduce unintended access, but they do not constitute an operating-system sandbox. Provider-managed authentication, network access, model selection, built-in behavior, and administrator policy remain under Claude Code control.

## Official sources

- [Run Claude Code programmatically](https://code.claude.com/docs/en/headless)
- [Claude Code CLI reference](https://code.claude.com/docs/en/cli-reference)
- [Manage permissions](https://code.claude.com/docs/en/permissions)
- [Authentication](https://code.claude.com/docs/en/authentication)

This record demonstrates one qualification on the stated CLI version and platform. It is not a perpetual compatibility claim, legal opinion, or Anthropic endorsement.
