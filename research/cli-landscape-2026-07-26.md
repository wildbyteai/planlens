# AI coding agent CLI landscape — 2026-07-26

This note records the primary-source research used to decide which local AI CLIs PlanLens can safely call as plan reviewers. It is a dated snapshot, not a permanent popularity claim.

## Ranking method

- Rank only active, public-source projects whose official repository documents a locally installable terminal coding-agent CLI on macOS.
- Sort by the official repository's GitHub `stargazers_count`, queried through the GitHub API on 2026-07-26 at approximately 13:00 UTC. Star counts change continuously.
- Use the CLI-specific repository when a brand has separate products. For example, the 82k-star OpenHands brand repository is now Agent Canvas; the separate OpenHands CLI repository had 224 stars and is primarily in stability maintenance.
- Exclude archived projects, projects whose official README says they are no longer actively maintained, IDE-only products, general-purpose interpreters, benchmarks, SDKs, and research issue-solving harnesses.
- Do not mix proprietary CLIs into the GitHub-star ranking. They are listed separately as compatibility candidates.
- A high rank does not imply PlanLens reviewer suitability. Reviewer suitability requires a documented single-prompt mode and a credible way to prevent writes and command execution, or an explicit warning that the CLI needs an outer sandbox.

## PlanLens compatibility decision

- Default: `codex + claude + kimi`, with Kimi gated on the documented no-tools custom-agent feature.
- Preferred fourth reviewer or replacement for an unavailable default: `gemini`.
- Formal compatibility after command validation: `codex`, `claude`, `kimi`, `qwen`, `pi`, `goose`, `aider`, and feature-gated `qoder`.
- Conditional compatibility: `gemini`, `opencode`, `copilot`, `cline`, `cursor`, `antigravity`, `zcode`, `crush`, and `kilo`.

This file records the landscape and ranking work. The later argument and isolation validation in [cli-validation-2026-07-26.md](cli-validation-2026-07-26.md) supersedes any recipe detail below when the two differ.

Formal compatibility means a non-interactive recipe and its required preflight are documented. It does not promise that every CLI version or local configuration passes. Conditional compatibility requires explicit user selection and disclosure of a weaker boundary.

## Active open-source top 10

| Rank | CLI | Official repository | Stars at snapshot | macOS install | Executable | PlanLens status |
|---:|---|---|---:|---|---|---|
| 1 | OpenCode | [anomalyco/opencode](https://github.com/anomalyco/opencode) | 189,808 | `brew install anomalyco/tap/opencode` | `opencode` | Conditional; merged config and persisted sessions require containment |
| 2 | Gemini CLI | [google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli) | 106,189 | `brew install gemini-cli` | `gemini` | Conditional; deny-all policy and isolated CLI home required |
| 3 | OpenAI Codex CLI | [openai/codex](https://github.com/openai/codex) | 101,551 | `brew install --cask codex` | `codex` | Suitable; strongest built-in read-only/ephemeral boundary |
| 4 | Pi | [earendil-works/pi](https://github.com/earendil-works/pi) | 77,882 | `npm install -g @earendil-works/pi-coding-agent` | `pi` | Suitable with no-tools/no-resources/no-session flags |
| 5 | Cline CLI | [cline/cline](https://github.com/cline/cline) | 65,067 | `npm install -g cline` | `cline` | Conditional; Plan Mode is not an OS read-only sandbox |
| 6 | goose | [aaif-goose/goose](https://github.com/aaif-goose/goose) | 51,722 | `brew install block-goose-cli` | `goose` | Suitable with `--no-profile` and `--no-session` |
| 7 | Aider | [Aider-AI/aider](https://github.com/Aider-AI/aider) | 47,709 | `python -m pip install aider-install` then `aider-install` | `aider` | Suitable in Ask Mode plus dry-run and disabled automation |
| 8 | Crush | [charmbracelet/crush](https://github.com/charmbracelet/crush) | 26,856 | `brew install charmbracelet/tap/crush` | `crush` | Conditional; reviewed restrictive config required |
| 9 | Kilo Code CLI | [Kilo-Org/kilocode](https://github.com/Kilo-Org/kilocode) | 26,529 | `brew install Kilo-Org/tap/kilo` | `kilo` | Conditional; Plan agent plus deny-by-default config required |
| 10 | Qwen Code | [QwenLM/qwen-code](https://github.com/QwenLM/qwen-code) | 26,333 | `brew install qwen-code` | `qwen` | Suitable with Safe Mode and Plan approval mode |

The narrow gaps near ranks 8–10 mean those positions may change quickly; the timestamp above is part of the ranking.

## Reviewer recipe research for the top 10

The following are argument-vector recipes. A host Agent should pass arguments directly, not build one interpolated shell command. Unless noted otherwise, run each CLI in its own empty temporary working directory and pass the complete PlanLens request as prompt text or stdin.

### 1. OpenCode

Official sources: [CLI](https://opencode.ai/docs/cli/), [permissions](https://opencode.ai/docs/permissions/), [configuration](https://opencode.ai/docs/config/).

Set child-process environment values without shell interpolation:

```text
OPENCODE_CONFIG_CONTENT={"permission":"deny","share":"disabled"}
OPENCODE_DISABLE_DEFAULT_PLUGINS=1
OPENCODE_DISABLE_CLAUDE_CODE=1
```

```text
opencode
  --pure
  run
  --format default
  [--model <provider/model>]
  <request-text>
```

This denies tools rather than relying only on the built-in `plan` agent. Do not add `--auto`: it automatically approves actions that remain `ask`. OpenCode can retain sessions/configuration; no verified ephemeral session flag was found.

### 2. Gemini CLI

Official sources: [CLI reference](https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/cli-reference.md), [Plan Mode](https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/plan-mode.md), [policy engine](https://geminicli.com/docs/reference/policy-engine/).

Create a temporary admin policy:

```toml
[[rule]]
toolName = "*"
decision = "deny"
priority = 999
```

```text
gemini
  -p <request-text>
  --approval-mode plan
  --sandbox
  -e none
  --admin-policy <temporary-deny-all-policy.toml>
  --output-format text
  [-m <explicit-user-choice>]
```

Plan Mode keeps the review read-only. Current policy documentation says a CLI-supplied admin policy is ignored when a standard system administrator policy already exists, so PlanLens must inspect the documented system policy directory before selection. If it contains a `.toml` policy, do not use this recipe unless an independently verified outer sandbox supplies the missing boundary. When the supplemental policy is effective, its deny-all rule remains active even if headless Plan Mode attempts a mode transition. Gemini does not document a general no-persistence flag for this invocation.

### 3. OpenAI Codex CLI

Official sources: [repository/install](https://github.com/openai/codex), [CLI reference](https://developers.openai.com/codex/cli/reference).

```text
codex --ask-for-approval never exec
  --sandbox read-only
  --ephemeral
  --skip-git-repo-check
  -c project_doc_max_bytes=0
  -C <empty-temporary-directory>
  [-m <explicit-user-choice>]
  -
```

Send the request through stdin. `--sandbox read-only` prevents writes and `--ephemeral` avoids saving session state. `--output-last-message <file>` may be used when available.

### 4. Pi

Official source: [Pi coding-agent CLI reference](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/README.md).

```text
pi
  --no-session
  --no-tools
  --no-extensions
  --no-skills
  --no-prompt-templates
  --no-context-files
  --no-approve
  -p
  [--model <provider/model>]
  <request-text>
```

Pi has no built-in permission prompts and otherwise inherits the launching process's permissions, so the no-tools/resource flags are essential. `--no-session` is the documented ephemeral mode. `PI_SKIP_VERSION_CHECK=1` and `PI_TELEMETRY=0` can suppress startup version checks and install telemetry without affecting model-provider traffic.

### 5. Cline CLI

Official sources: [CLI reference](https://docs.cline.bot/cli/cli-reference), [headless mode](https://docs.cline.bot/cli/headless), [command permissions](https://docs.cline.bot/cli/command-permissions).

Candidate invocation:

```text
CLINE_COMMAND_PERMISSIONS={"allow":[],"deny":["*"],"allowRedirects":false}

cline
  --plan
  --auto-approve false
  --cwd <empty-temporary-directory>
  --hooks-dir <empty-hooks-directory>
  [--model <explicit-user-choice>]
  <request-text>
```

Cline's headless CLI otherwise defaults to auto-approval. Plan Mode is intended for analysis/planning, and the command permission deny rule blocks shell commands, but this is not equivalent to an OS-level read-only sandbox. Cline also maintains history/state unless an isolated data directory is used; isolating it may require an explicit existing configuration directory for authentication. Treat this recipe as conditional until tested against a supported installed version.

### 6. goose

Official sources: [installation](https://github.com/aaif-goose/goose/blob/main/documentation/docs/getting-started/installation.md), [running tasks](https://github.com/aaif-goose/goose/blob/main/documentation/docs/guides/running-tasks.md), [CLI source](https://github.com/aaif-goose/goose/blob/main/crates/goose-cli/src/cli.rs).

```text
goose run
  --no-profile
  --no-session
  --quiet
  [--provider <explicit-user-choice>]
  [--model <explicit-user-choice>]
  --text <request-text>
```

`--no-profile` disables default extensions; with no `--with-*` extensions supplied, the review is text-only. `--no-session` discards session storage. Do not confuse this with `goose review`: there, `--prompt` names a prompt file for code-range review; normal one-shot text uses `goose run --text`.

### 7. Aider

Official sources: [installation](https://aider.chat/docs/install.html), [options](https://aider.chat/docs/config/options.html), [Ask Mode](https://aider.chat/docs/usage/modes.html).

```text
aider
  --chat-mode ask
  --message <request-text>
  --dry-run
  --yes-always
  --no-git
  --no-auto-commits
  --no-auto-lint
  --no-auto-test
  --no-suggest-shell-commands
  --disable-playwright
  --chat-history-file /dev/null
  --input-history-file /dev/null
  --no-pretty
  --no-stream
  [--model <explicit-user-choice>]
```

Ask Mode answers without editing, while `--dry-run` provides an additional no-modification boundary. Important: Aider's short `-m` flag means `--message`, not model; use the full `--model` option.

### 8. Crush

Official sources: [README/install](https://github.com/charmbracelet/crush), [`run` implementation](https://github.com/charmbracelet/crush/blob/main/internal/cmd/run.go).

The actual one-shot syntax is:

```text
crush run
  --quiet
  [--model <provider/model>]
  <request-text>
```

No reliable one-line CLI switch was found to disable every tool or make the session ephemeral. `--yolo` only skips permission requests and is unsafe for review isolation. A custom data/config directory with explicit `options.disabled_tools` and `permissions.allowed_tools` would be required and must preserve provider authentication. Therefore do not add Crush as a default reviewer yet.

### 9. Kilo Code CLI

Official sources: [repository/install](https://github.com/Kilo-Org/kilocode), [agent overview](https://github.com/Kilo-Org/kilocode#agents).

The documented one-shot shape is:

```text
kilo run
  --agent plan
  [--model <provider/model>]
  <request-text>
```

The built-in Plan agent avoids file editing, but unqualified tool permissions can still remain `ask`. `--auto` approves those actions and must not be used as a review-isolation substitute. A dedicated deny-by-default PlanLens agent/config is required before Kilo should become a default reviewer.

### 10. Qwen Code

Official sources: [repository/install](https://github.com/QwenLM/qwen-code), [headless mode](https://github.com/QwenLM/qwen-code/blob/main/docs/users/features/headless.md), [CLI configuration](https://qwenlm.github.io/qwen-code-docs/en/users/configuration/cli/).

```text
qwen
  -p <request-text>
  --safe-mode
  --approval-mode plan
  --output-format text
  [-m <explicit-user-choice>]
```

Safe Mode disables context files, hooks, extensions, skills, MCP servers, custom subagents, and permission overrides from configuration; the CLI-specified Plan approval mode still applies. Qwen stores project-scoped JSONL sessions and does not document a general no-persistence flag.

## Additional proprietary or below-cutoff CLIs

These do not participate in the open-source star ranking. “Suitable” means only that a documented single-prompt reviewer recipe exists; it does not mean the vendor guarantees complete data isolation.

| CLI | Official source | macOS install | Executable | Reviewer conclusion |
|---|---|---|---|---|
| Claude Code | [setup](https://code.claude.com/docs/en/setup), [CLI reference](https://code.claude.com/docs/en/cli-reference) | `curl -fsSL https://claude.ai/install.sh \| bash` or `brew install --cask claude-code` | `claude` | Suitable: safe mode, no built-in tools, and no session persistence |
| GitHub Copilot CLI | [installation](https://docs.github.com/en/copilot/how-tos/copilot-cli/install-copilot-cli), [programmatic reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference) | `brew install copilot-cli` or `npm install -g @github/copilot` | `copilot` | Conditional; an empty available-tool set is supported, but no strict no-session flag exists |
| Qoder CLI | [quick start](https://docs.qoder.com/en/cli/quick-start), [usage](https://docs.qoder.com/en/cli/using-cli), [permissions](https://docs.qoder.com/en/cli/permissions) | `curl -fsSL https://qoder.com/install \| bash` | `qodercli` | Formal when no-tools, no-session-persistence, strict MCP, and isolated-config flags are present |
| Cursor Agent | [installation](https://cursor.com/docs/cli/installation), [parameters](https://cursor.com/docs/cli/reference/parameters), [headless](https://cursor.com/docs/cli/headless) | `curl https://cursor.com/install -fsS \| bash` | `agent` | Conditional: Ask/Plan Mode without `--force`; workspace sandbox is not read-only |
| Kimi Code CLI | [current repository](https://github.com/MoonshotAI/kimi-code), [command reference](https://github.com/MoonshotAI/kimi-code/blob/main/docs/en/reference/kimi-command.md), [agents](https://moonshotai.github.io/kimi-code/en/customization/agents.html) | `curl -fsSL https://code.kimi.com/kimi-code/install.sh \| bash` | `kimi` | Formal for versions exposing `--agent-file`; use an explicit no-tools/no-subagents custom agent |
| ZCode CLI | [product site](https://zcode.z.ai/) | Install the desktop app; no standalone CLI installer or guaranteed PATH command is published | bundled `zcode.cjs` | Conditional local compatibility only; `--prompt` defaults to yolo, so explicit `--mode plan` is mandatory |
| Amp | [official manual](https://ampcode.com/manual) | Use the installer documented by the official manual | `amp` | Do not add until a current official no-tools/read-only one-shot recipe is verified |
| Antigravity CLI | [official installation](https://antigravity.google/docs/cli/install) | `curl -fsSL https://antigravity.google/cli/install.sh \| bash` | `agy` | Conditional: public install is verified; keep `--mode plan --sandbox --print`, but its complete isolation boundary still needs versioned validation |

### Verified proprietary recipes

Claude Code:

```text
claude
  --print
  --safe-mode
  --no-session-persistence
  --no-chrome
  --disable-slash-commands
  --permission-mode dontAsk
  --tools ""
  --output-format text
  [--model <explicit-user-choice>]
  <request-text>
```

GitHub Copilot CLI, after confirming no user-configured MCP server has a persistent permission outside the deny set:

```text
copilot
  -p <request-text>
  -s
  --no-ask-user
  --disable-builtin-mcps
  --deny-tool "shell,write,read,url,memory"
  [--model <explicit-user-choice>]
```

Kimi Code CLI, gated by the presence of `--agent-file` rather than a guessed version floor:

```markdown
---
name: planlens-reviewer
description: Review the supplied plan without tools or delegation
tools: []
subagents: []
---

Review only the supplied PlanLens request. Do not use tools, read files, retrieve external material, delegate work, or perform the plan. Return only the final review.
```

```text
KIMI_CODE_EXPERIMENTAL_FLAG=1

kimi
  --prompt <request-text>
  --agent-file <temporary-agent-file>
  --skills-dir <empty-skills-directory>
  --output-format text
  [--model <explicit-user-choice>]
```

The custom agent body owns the system prompt and the empty `tools` and `subagents` lists prevent tool use and delegation. Kimi may still retain a local session or execute user-configured lifecycle hooks; disclose that boundary.

Qoder CLI:

```text
qodercli
  -p
  --permission-mode dont_ask
  --disallowed-tools "*"
  --output-format text
  [--model <explicit-user-choice>]
  <request-text>
```

Cursor Agent, limited boundary:

```text
agent
  -p
  --mode ask
  --sandbox enabled
  --output-format text
  [--model <explicit-user-choice>]
  <request-text>
```

Do not add `--force`, `--yolo`, `--continue`, `--resume`, or a session identifier.

ZCode, local version checked at 0.15.0:

```text
zcode
  --prompt <request-text>
  --mode plan
  --cwd <empty-temporary-directory>
  --no-color
```

The bundled ZCode 0.15.0 source supports the `ZCODE_MODEL`, `ZCODE_STORAGE_DIR`, and `ZCODE_SESSION_DB_PATH` environment variables. It still loads configurable resources and has no no-session flag, so restrictive config and temporary storage are required.

## Below-cutoff and excluded projects

| Project | Snapshot | Decision |
|---|---:|---|
| Plandex | 15,545 stars | Active candidate below the top-10 cutoff; not researched deeply enough in this pass for a safe reviewer recipe |
| Kimi Code CLI | 5,127 stars in the current `MoonshotAI/kimi-code` repository | Below the open-source top-10 cutoff but formally compatible through the version-gated no-tools custom-agent recipe; never use `--plan --prompt` or bare prompt mode |
| OpenHands CLI | 224 stars in `OpenHands/OpenHands-CLI` | CLI-specific repository is primarily stability-maintained; headless mode auto-approves actions, so do not run directly on the host as a reviewer |
| Continue | 35,118 stars | Official README says the repository is no longer actively maintained and is read-only; excluded despite its star count |
| GPT Pilot | 33,726 stars | Official README says it is unmaintained and discloses a credential-stealing supply-chain compromise from 2025-08-24 through 2026-06-11; do not recommend or support |
| Roo Code | 24,363 stars | Official repository is archived; excluded |
| Amazon Q Developer CLI | 1,982 stars | Official repository says active maintenance stopped and users should move to closed-source Kiro CLI; excluded |
| OpenHands brand repository | 82,132 stars | Now represents the broader Agent Canvas project, not the standalone CLI; do not use this count for CLI ranking |
| SWE-agent | 19,917 stars | Research/issue-solving harness rather than a general local coding-review assistant; out of scope |
| Open Interpreter | 67,303 stars | General-purpose computer/terminal interpreter, not primarily a coding-agent CLI; out of scope |

## Command mistakes to avoid

1. Current Kimi Code CLI rejects `--prompt` combined with `--plan`; use the version-gated `--agent-file` no-tools recipe and never fall back to bare prompt mode.
2. In Aider, `-m` means `--message`, not model. Use `--model` in full.
3. In goose, normal one-shot text is `goose run --text`; `goose review --prompt` treats the value as a prompt file path for code-range review.
4. OpenCode/Kilo `--auto`, Crush `--yolo`, OpenHands `--headless`, and ZCode's default prompt mode are automation/approval shortcuts, not read-only isolation.
5. A CLI's “plan” or “ask” mode is not automatically an OS sandbox. Only describe the exact boundary the vendor documents.
6. Do not claim ephemeral execution unless the CLI documents it (`codex --ephemeral`, `claude --no-session-persistence`, `pi --no-session`, and `goose --no-session` are verified examples).

## Recommended PlanLens expansion order

1. Keep the default small: `codex + claude + kimi`, with Kimi feature-gated.
2. Use `gemini` as the preferred fourth reviewer or pre-confirmation fallback when its admin-policy preflight passes.
3. Keep `qwen`, `pi`, `goose`, `aider`, and feature-gated `qoder` as additional formal recipes. Keep `gemini`, `opencode`, and `copilot` conditional.
4. Keep `cline`, `cursor`, `antigravity`, `zcode`, `crush`, and `kilo` conditional rather than presenting their plan or ask modes as full isolation.
5. Do not add `openhands` or `amp` until a current versioned safe recipe exists.
