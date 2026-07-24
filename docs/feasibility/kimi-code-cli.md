# Kimi Code CLI feasibility

- Status: blocked — executable configuration and session persistence are not safely controllable
- Support status: not qualified
- Qualification date: 2026-07-24
- Platform: `darwin/arm64`
- CLI package: `@moonshot-ai/kimi-code`
- CLI version: `0.29.1`
- Public fixture SHA-256: `7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429`
- Access capability: fake-CLI protocol path only: `constrained`; real CLI path blocked before model execution
- Material scope: the fixed public fixture only; other material is rejected before Kimi starts

## What was verified

- PlanLens requires `KIMI_CODE_HOME` to explicitly name an absolute, existing, dedicated directory. It does not fall back to the user's normal home directory.
- Before version discovery and again before every review, PlanLens rejects `AGENTS.md`, `SYSTEM.md`, `config.toml`, `mcp.json`, and `plugins/installed.json` under that directory. Regular files, directories, and symbolic-link entries at those paths are all treated as present; their contents are not read.
- PlanLens discovered the fixed `kimi` command and obtained version `0.29.1` without reading CLI configuration or credentials.
- Before constructing the prompt argument, the adapter requires the exact fixed public-fixture SHA-256. Non-public or changed plans are rejected without starting a Kimi review process.
- The repository pins `testdata/public-plan.md` to LF checkout semantics. A regression test performs a Windows-style `core.autocrlf=true` checkout and verifies the raw fixture retains SHA-256 `7d70e4d6686b0903b1a739f0b6b73dfc4519d61d84bad9042e8e3cdfe4acd429`.
- Version discovery and review both use fresh temporary working directories outside the real repository. `HOME`, `USERPROFILE`, `APPDATA`, `LOCALAPPDATA`, `TEMP`, `TMP`, and `TMPDIR` point inside the temporary workspace; only the dedicated provider data root named by `KIMI_CODE_HOME` is retained.
- PlanLens started a fresh non-interactive prompt-mode process. It did not use session resume, continuation, `--auto`, or `--yolo` options.
- The review invocation selected an explicit Agent definition with empty tool and subagent lists, replaced automatically discovered user and project skills with an empty directory, disabled Kimi telemetry, and set `KIMI_LOG_LEVEL=off` to reduce ordinary Kimi logging.
- The adapter requested `stream-json`, checked the version metadata emitted by Kimi, discarded meta events, rejected tool-call or tool-result events, and normalized only the final assistant content.
- Fake-CLI tests verified frozen public-fixture delivery, fresh-process arguments, isolated home and temporary directories during both version discovery and review, `config.toml` and other fixed-load-path rejection, empty tools and skills, synchronized context cancellation semantics, final-response normalization, and immediate disposal of resume metadata.

## Result

The revised adapter rejects any `KIMI_CODE_HOME` containing `config.toml`. In Kimi Code CLI `0.29.1`, that file can provide model configuration while also enabling hooks and `extra_skill_dirs`. Accepting it would therefore allow executable behavior or additional context to enter the review, while inspecting selected fields would require PlanLens to read provider configuration that may contain credentials. PlanLens does neither.

As a result, a real official login home that contains `config.toml` cannot qualify with this candidate. This is a hard fail-closed boundary, not an instruction asking users to avoid hooks. No new real model review was performed, the earlier attempt did not reach a model, and the adapter remains blocked and not qualified. Qualification would require a native CLI control or provider layout that separates authentication and model selection from hooks and extra context, followed by a successful rerun of the immutable candidate and public fixture.

There is a second independent hard blocker. Kimi Code CLI `0.29.1` print execution in both v1 and v2 persists session wire records under `KIMI_CODE_HOME/sessions/.../wire.jsonl`. Those records can contain the prompt, accumulated context, assistant output, and thinking content. The tested CLI exposes no official no-session or ephemeral control that prevents this write.

PlanLens does not treat post-run deletion as immediate reasoning disposal: the material has already been persisted, deletion can fail, and deleting provider session files after execution cannot prove that no recoverable copy existed. `KIMI_LOG_LEVEL=off` only reduces ordinary logs and does not disable `wire.jsonl` persistence. Even if the `config.toml` boundary were resolved, real Kimi support would remain blocked until the official CLI provides a no-persistence control or authentication can be established in a genuinely isolated temporary provider home without copying or exposing user credentials.

## Access boundary

The fake-CLI protocol path is classified as `constrained`, not `enforced`. Empty model tools and subagents, an empty skills directory, a temporary workspace, isolated home paths, and fixed-load-path rejection reduce the review surface, but they are not an operating-system sandbox. This classification does not qualify real Kimi support: executable configuration blocks the current authentication path, and unavoidable session-wire persistence independently blocks the required disposal semantics.

PlanLens does not inspect, sanitize, or selectively copy `config.toml`. Presence alone, including a symbolic link, is rejected during discovery and checked again immediately before each review. The same fail-closed rule applies to the other known user-level instruction and extension entry points.

## Unresolved limitations

- Successful real final-response capture remains unverified because the current CLI does not safely separate model configuration from hooks and extra skills in `config.toml`.
- An official login home containing `config.toml` is rejected; user assurance about its contents is not accepted as a substitute for an enforceable CLI control.
- Kimi `0.29.1` persists prompt, context, assistant, and thinking records in `sessions/.../wire.jsonl`; no tested official option prevents this persistence.
- `KIMI_LOG_LEVEL=off` does not solve session-wire persistence, and post-run deletion is not accepted as proof of immediate disposal.
- Custom Agent selection currently requires Kimi's experimental feature flag.
- Kimi prompt mode requires prompt text in a process argument. The feasibility adapter therefore supports only the fixed public fixture; private or sensitive plans are rejected.
- PlanLens does not inspect or remove provider-managed credentials, logs, or session state in the dedicated home.
- Version handling is intentionally limited to the exact tested candidate. Version `0.29.1` is not qualified unless native safe-separation and no-persistence controls exist, or genuinely safe temporary authentication isolation becomes possible, and a real review then succeeds.

## Official sources

- [Kimi Code CLI getting started](https://www.kimi.com/code/docs/en/kimi-code-cli/guides/getting-started.html)
- [Kimi command reference](https://www.kimi.com/code/docs/en/kimi-code-cli/reference/kimi-command.html)
- [Environment variables](https://www.kimi.com/code/docs/en/kimi-code-cli/configuration/env-vars.html)
- [Custom Agents](https://www.kimi.com/code/docs/en/kimi-code-cli/customization/agents.html)
- [MCP](https://www.kimi.com/code/docs/en/kimi-code-cli/customization/mcp.html)
- [Hooks](https://www.kimi.com/code/docs/en/kimi-code-cli/customization/hooks.html)
- [Plugins](https://www.kimi.com/code/docs/en/kimi-code-cli/customization/plugins.html)
- [Data and privacy](https://www.kimi.com/code/docs/en/kimi-code-cli/guides/data-and-privacy.html)

This record describes one controlled attempt on the stated CLI version and platform. It is not a perpetual compatibility claim, legal opinion, or Moonshot AI endorsement.
