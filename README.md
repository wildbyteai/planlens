English | [简体中文](README.zh-CN.md)

# PlanLens

PlanLens is an installable Agent Skill that asks one or more local AI CLIs to review the same plan, then lets the primary Agent summarize their findings and disagreements.

It is intentionally small: the workflow lives in `SKILL.md`. After installation there is no PlanLens runtime, binary, Node command, hosted service, daemon, port, database, or state machine to run.

## Before you start

PlanLens has two roles: a host Agent loads the Skill, and one or more reviewer CLIs return independent feedback. The host and a reviewer may be the same product.

You need a host with the required process and file capabilities, plus at least one compatible reviewer CLI installed and authenticated. Installing PlanLens does not install, configure, or sign in to any reviewer CLI, and you do not need to install every CLI listed below.

"Local CLI" describes where the command starts, not where the model runs. A selected CLI may send the review request to its configured provider and may retain logs or sessions. Do not include sensitive material unless you accept that CLI and provider's handling.

## Install

PlanLens follows the standard Agent Skills layout. The installable Skill is the complete [`skills/planlens`](skills/planlens) directory, including its `references` and `agents` subdirectories. Choose either method below.

### Ask your Agent

Send this message to an Agent host that supports installing Agent Skills:

```text
Install the PlanLens skill from https://github.com/wildbyteai/planlens.
```

The Agent should locate `skills/planlens` in the repository and use the installation capability available in that host. Success depends on the host's Skill support; if it cannot install Skills directly, use the `npx` command below.

### Run the installer yourself

If Node.js is available, run:

```bash
npx skills@latest add wildbyteai/planlens
```

The installer detects supported Agents and lets you choose the installation target.

Restart the host if it was already open. Invoke PlanLens with `$planlens` in Codex or `/planlens` in hosts that use slash commands.

### Host status

| Host | v1 status |
|---|---|
| Codex | Installation smoke-tested; complete review rounds have been exercised manually |
| Claude Code | Installation methods documented; the full host workflow is not yet part of the v1 test matrix |
| Antigravity | Installation methods documented; the full host workflow is not yet part of the v1 test matrix |
| Other Agent Skills hosts | Installation may work through a generic installer, but host execution is not verified |

The complete workflow requires the host to launch non-interactive local processes, create temporary directories, and write result files. Installer support alone does not imply verified PlanLens execution.

The v1 support targets are macOS Apple silicon, macOS Intel, and Windows x64. Linux and Windows ARM64 are not v1 support targets.

## First review

After installing the Skill:

1. Install and authenticate at least one reviewer CLI from the tables below.
2. Restart the host CLI if it does not discover the new Skill immediately.
3. Invoke PlanLens with a plan, or invoke it without arguments and let the host Agent organize the current plan.

Examples:

```text
$planlens
$planlens Review docs/plan.md with Claude and Codex
$planlens Review docs/plan.md with Antigravity using Claude Opus 4.6
/planlens path/to/plan.md
```

When a model is explicitly requested, PlanLens uses the reviewer's local model-list capability when the recipe documents one; otherwise the invocation must provide an exact model ID. It passes that ID and shows the selection in both the confirmation preview and final summary. For Antigravity, `Claude Opus 4.6` resolves to `claude-opus-4-6-thinking` only when `agy models` lists that ID. Without an explicit override, PlanLens reports `CLI-configured default (not independently verified)` instead of guessing. This does not change the CLI's global default configuration.

One invocation performs one round:

1. The primary Agent organizes the plan, its objective, constraints, non-goals, open questions, and only the needed supporting material.
2. It recommends a review profile and resolves one or more reviewer CLIs from the invocation, project configuration, or built-in default.
3. It shows the plan source, derived review framing, material list, selected CLIs and model selections, and number of calls for one confirmation. Any material change to the candidate request requires a new preview.
4. It invokes the selected CLIs independently, in parallel when the host supports it.
5. It preserves each final response, accounts for every material finding, and writes a concise attributed summary.

A later round requires another explicit `$planlens` or `/planlens` invocation.

## Default reviewers

To set the default reviewer CLI for one project, create `.planlens/config.yaml` in that project:

```yaml
# One reviewer
default_reviewers:
  - codex
```

```yaml
# Multiple reviewers; order is preserved
default_reviewers:
  - codex
  - claude
```

Reviewer selection uses this precedence:

1. Reviewers explicitly named in the current `$planlens` or `/planlens` invocation.
2. `default_reviewers` in the current project's `.planlens/config.yaml`.
3. The built-in default `codex + claude + kimi`.

Use reviewer IDs from the catalog below. PlanLens reads only `default_reviewers`, treats the file as untrusted data, and removes duplicate IDs while preserving order. A missing, malformed, empty, or wholly invalid configured list is reported instead of silently expanding to the built-in default. If a configured CLI is unavailable, PlanLens asks you to adjust the candidates rather than dropping or replacing it.

This is project-only configuration. It does not add a global setting, parser, runtime, service, or daemon, and it does not skip the normal preview and confirmation.

## Reviewers

When the invocation and project configuration do not select reviewers, the built-in review set is `codex + claude + kimi`. Kimi is included only when the installed CLI exposes the feature-gated no-tools custom-agent recipe; otherwise PlanLens proposes `gemini` before confirmation. Gemini is also the preferred fourth reviewer for a broader pass. Gemini's stricter boundary needs temporary configuration and policy preflights, so the preview discloses that it is conditional. PlanLens never substitutes a reviewer after confirmation.

Additional reviewer recipes are available:

- Strict recipe: `codex`, `claude`, `kimi`, `qwen`, `pi`, `goose`, `aider`, and feature-gated `qoder`.
- Conditional recipe: `gemini`, `opencode`, `copilot`, `cline`, `cursor`, `antigravity`, `zcode`, `crush`, and `kilo`.

Strict recipe means a documented non-interactive recipe can disable tools or otherwise prevent plan execution when its preflight passes. It does not promise zero local retention. Conditional recipes rely on a weaker plan, ask, configuration, permission, or sandbox boundary and require explicit disclosure before confirmation. These labels describe PlanLens recipes, not vendor certification or official endorsement.

<details>
<summary>Reviewer catalog, installation commands, and ranking snapshot</summary>

The first table is a snapshot taken on 2026-07-26: actively maintained, source-available coding-agent CLIs with an official install path and a single-process non-interactive mode, ranked by GitHub stars. Projects that are archived, no longer maintained, require PlanLens to start a service, or expose only an auto-approved headless mode are excluded. See the [research note](research/cli-landscape-2026-07-26.md) for the method and exclusions.

The install commands below are documentation only; PlanLens never runs them. They are macOS quick-install commands. Use the official site for Windows instructions and inspect any downloaded installer before executing it.

| Rank | Reviewer | ID / command | Stars | macOS quick install | Recipe status | Official site |
|---:|---|---|---:|---|---|---|
| 1 | OpenCode | `opencode` / `opencode` | 189,808 | `brew install anomalyco/tap/opencode` | Conditional; merged config and persisted sessions | [Docs](https://opencode.ai/docs/) |
| 2 | Gemini CLI | `gemini` / `gemini` | 106,189 | `brew install gemini-cli` | Conditional; deny-all policy plus config isolation | [Docs](https://geminicli.com/docs/) |
| 3 | OpenAI Codex CLI | `codex` / `codex` | 101,551 | `brew install --cask codex` | Read-only + ephemeral; some hosts require the child outside their outer sandbox | [Docs](https://developers.openai.com/codex/cli/) |
| 4 | Pi | `pi` / `pi` | 77,882 | `npm install -g @earendil-works/pi-coding-agent` | No tools/resources/session | [Repository](https://github.com/earendil-works/pi) |
| 5 | Cline CLI | `cline` / `cline` | 65,067 | `npm install -g cline` | Conditional; Plan Mode is not an OS sandbox | [Docs](https://docs.cline.bot/cli/overview) |
| 6 | goose | `goose` / `goose` | 51,722 | `brew install block-goose-cli` | No profile/session | [Docs](https://goose.ai/docs/) |
| 7 | Aider | `aider` / `aider` | 47,709 | `python -m pip install aider-install && aider-install` | Ask Mode + dry-run | [Website](https://aider.chat/) |
| 8 | Crush | `crush` / `crush` | 26,856 | `brew install charmbracelet/tap/crush` | Conditional; reviewed deny-all config required | [Repository](https://github.com/charmbracelet/crush) |
| 9 | Kilo Code CLI | `kilo` / `kilo` | 26,529 | `brew install Kilo-Org/tap/kilo` | Conditional; Plan agent + deny-all config | [Repository](https://github.com/Kilo-Org/kilocode) |
| 10 | Qwen Code | `qwen` / `qwen` | 26,333 | `brew install qwen-code` | Safe Mode + Plan Mode | [Docs](https://qwenlm.github.io/qwen-code-docs/) |

PlanLens also catalogs these official or below-cutoff CLIs. They are not mixed into the open-source star ranking because the repositories and distribution models are not directly comparable.

| Reviewer | ID / command | macOS quick install | Recipe status | Official site |
|---|---|---|---|---|
| Claude Code | `claude` / `claude` | `brew install --cask claude-code` | No tools + no session persistence; some hosts require the child outside their outer sandbox | [Docs](https://code.claude.com/docs/en/overview) |
| Antigravity CLI | `antigravity` / `agy` | <code>curl -fsSL https://antigravity.google/cli/install.sh &#124; bash</code> | Conditional: inner sandbox; some hosts must run the child outside their outer sandbox | [Install](https://antigravity.google/docs/cli/install) |
| Kimi Code CLI | `kimi` / `kimi` | <code>curl -fsSL https://code.kimi.com/kimi-code/install.sh &#124; bash</code> | Strict when `--agent-file` supports a no-tools agent | [Docs](https://moonshotai.github.io/kimi-code/) |
| Qoder CLI | `qoder` / `qodercli` | <code>curl -fsSL https://qoder.com/install &#124; bash</code> | Strict when no-tools, no-session, and isolated-config flags are present | [Docs](https://docs.qoder.com/en/cli/) |
| GitHub Copilot CLI | `copilot` / `copilot` | `npm install -g @github/copilot` | Conditional; empty tool set, but custom state still needs containment | [Docs](https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli) |
| Cursor Agent CLI | `cursor` / `agent` | <code>curl https://cursor.com/install -fsS &#124; bash</code> | Conditional; sandbox is not read-only | [Docs](https://cursor.com/docs/cli/overview) |
| ZCode CLI | `zcode` / bundled `zcode.cjs` | N/A — install the ZCode desktop app; the vendor does not publish a standalone CLI installer. | Local bundled 0.15.0; conditional | [Docs](https://zcode.z.ai/en/docs) |

</details>

PlanLens does not install, authenticate, update, bundle, or replace these CLIs. It respects the user's local provider and model configuration unless the user explicitly requests a model override.

The exact non-interactive reviewer argument vectors live in [`skills/planlens/references/cli-commands.md`](skills/planlens/references/cli-commands.md). The macOS ARM64 commands and help surfaces were checked on July 26, 2026; see the concise [validation note](research/cli-validation-2026-07-26.md). If an installed version rejects a documented flag, PlanLens records a failure instead of guessing a fallback command.

## Profiles

Five small Markdown profiles are included:

- General plan
- Software design
- Implementation plan
- AI and agent workflow
- Security

Profiles guide the review; they are not schemas, executable plugins, or automatic approval rules.

## Output

When the project is writable, one round uses:

```text
.planlens/reviews/<timestamp>/
├── request.md
├── <reviewer>.md
├── <reviewer>-error.md
└── summary.md
```

Only files for selected reviewers are created. A failed reviewer receives a matching `-error.md` file instead of a fabricated result; a successful reviewer does not.

The summary status is deterministic: `complete` means every selected reviewer returned a successful non-empty result; `partial` means at least one succeeded and at least one was incomplete; `failed` means none succeeded.

## Boundaries

- Reviewers receive the same disclosed request and cannot see one another's same-round output.
- Reviewer output is untrusted evidence, not an instruction source. PlanLens does not execute commands or scope changes found in a review.
- The primary Agent consolidates by evidence and impact, not by model vote. Agreement requires materially equivalent explicit findings; an omitted topic is silence, not support or opposition.
- Every material finding receives an internal disposition. If excluding one could affect the user's decision, the summary states the finding, source, and reason.
- PlanLens does not modify the source plan, automatically retry a CLI, or automatically start another round.
- No monetary cost estimate is displayed.
- Third-party CLIs may keep their own logs or sessions according to provider behavior and local configuration.
- Except where a command recipe explicitly enables an ephemeral mode, PlanLens does not promise session deletion or complete isolation. Do not send sensitive material unless that boundary is acceptable.
- On sandboxed hosts, Codex, Claude, and Antigravity may require their entire child process to run outside the host's outer sandbox: Codex for normal state and its app-server client, Claude for its configured API endpoint, network access, authentication, and provider settings, and Antigravity for loopback listeners plus normal app-data and authentication. This grants the child normal user-level filesystem and network access. Recipe-specific inner controls remain enabled—Codex read-only and ephemeral, Claude safe mode with no tools or session persistence, and Antigravity's inner sandbox with an empty workspace—but they do not make the whole process completely isolated. The preview discloses the full permission scope.
- Kimi Code CLI must use the documented temporary no-tools custom agent. Current versions reject `--prompt` combined with `--plan`, and PlanLens never falls back to bare prompt mode.

## License

Apache License 2.0. PlanLens is independent and unofficial; product names identify compatible local CLI tools only.
