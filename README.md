# PlanLens

PlanLens is an installable Agent Skill that asks one or more local AI CLIs to review the same plan, then lets the primary Agent summarize their findings and disagreements.

It is intentionally small: the workflow lives in `SKILL.md`. There is no PlanLens runtime, binary, Node command, hosted service, daemon, port, database, or state machine.

## Before you start

PlanLens has two roles: a host Agent loads the Skill, and one or more reviewer CLIs return independent feedback. The host and a reviewer may be the same product.

You need a supported host and at least one compatible reviewer CLI installed and authenticated. Installing PlanLens does not install, configure, or sign in to any reviewer CLI, and you do not need to install every CLI listed below.

## Install

Install the complete [`skills/planlens`](skills/planlens) directory, including its `references` and `agents` subdirectories.

Use the `v1.1.0` tag for the published release. Replace `v1.1.0` with `main` if you intentionally want the current development version.

### Codex

Ask Codex:

```text
Use $skill-installer to install PlanLens from https://github.com/wildbyteai/planlens/tree/v1.1.0/skills/planlens.
```

Invoke it on a later turn with `$planlens`.

### Claude Code

Copy `skills/planlens` to `~/.claude/skills/planlens` for a personal installation, or to `.claude/skills/planlens` inside one project. Invoke it with `/planlens`.

### Antigravity

Copy `skills/planlens` to `~/.gemini/config/skills/planlens` for a personal installation, or to `.agents/skills/planlens` inside one project. Restart the CLI if it was already open, then invoke `/planlens`.

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
/planlens path/to/plan.md
```

One invocation performs one round:

1. The primary Agent organizes the plan and only the needed supporting material.
2. It recommends a review profile and lets the user select one or more local CLIs.
3. It shows the plan source, material list, selected CLIs, and number of calls for one confirmation.
4. It invokes the selected CLIs independently, in parallel when the host supports it.
5. It preserves each final response and writes a concise attributed summary.

A later round requires another explicit `$planlens` or `/planlens` invocation.

## Reviewers

The default review set is `codex + claude + kimi`. Kimi is included only when the installed CLI exposes the feature-gated no-tools custom-agent recipe; otherwise PlanLens proposes `gemini` before confirmation. Gemini is also the preferred fourth reviewer for a broader pass. Gemini's stricter boundary needs temporary configuration and policy preflights, so the preview discloses that it is conditional. PlanLens never substitutes a reviewer after confirmation.

The system remains wider than the default:

- Formal compatibility: `codex`, `claude`, `kimi`, `qwen`, `pi`, `goose`, `aider`, and feature-gated `qoder`.
- Conditional compatibility: `gemini`, `opencode`, `copilot`, `cline`, `cursor`, `antigravity`, `zcode`, `crush`, and `kilo`.

Formal compatibility means a documented non-interactive recipe can disable tools or otherwise prevent plan execution when its preflight passes. It does not promise zero local retention. Conditional reviewers rely on a weaker plan, ask, configuration, permission, or sandbox boundary and require explicit disclosure before confirmation.

The first table is a snapshot taken on 2026-07-26: actively maintained, source-available coding-agent CLIs with an official install path and a single-process non-interactive mode, ranked by GitHub stars. Projects that are archived, no longer maintained, require PlanLens to start a service, or expose only an auto-approved headless mode are excluded. See the [research note](research/cli-landscape-2026-07-26.md) for the method and exclusions.

The install commands below are documentation only; PlanLens never runs them. They are macOS quick-install commands. Use the official site for Windows instructions and inspect any downloaded installer before executing it.

| Rank | Reviewer | ID / command | Stars | macOS quick install | Recipe status | Official site |
|---:|---|---|---:|---|---|---|
| 1 | OpenCode | `opencode` / `opencode` | 189,808 | `brew install anomalyco/tap/opencode` | Conditional; merged config and persisted sessions | [Docs](https://opencode.ai/docs/) |
| 2 | Gemini CLI | `gemini` / `gemini` | 106,189 | `brew install gemini-cli` | Conditional; deny-all policy plus config isolation | [Docs](https://geminicli.com/docs/) |
| 3 | OpenAI Codex CLI | `codex` / `codex` | 101,551 | `brew install --cask codex` | Read-only + ephemeral | [Docs](https://developers.openai.com/codex/cli/) |
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
| Claude Code | `claude` / `claude` | `brew install --cask claude-code` | No tools + no session persistence | [Docs](https://code.claude.com/docs/en/overview) |
| Antigravity CLI | `antigravity` / `agy` | <code>curl -fsSL https://antigravity.google/cli/install.sh &#124; bash</code> | Conditional: Plan Mode + sandbox | [Install](https://antigravity.google/docs/cli/install) |
| Kimi Code CLI | `kimi` / `kimi` | <code>curl -fsSL https://code.kimi.com/kimi-code/install.sh &#124; bash</code> | Formal when `--agent-file` supports a no-tools agent | [Docs](https://moonshotai.github.io/kimi-code/) |
| Qoder CLI | `qoder` / `qodercli` | <code>curl -fsSL https://qoder.com/install &#124; bash</code> | Formal when no-tools, no-session, and isolated-config flags are present | [Docs](https://docs.qoder.com/en/cli/) |
| GitHub Copilot CLI | `copilot` / `copilot` | `npm install -g @github/copilot` | Conditional; empty tool set, but custom state still needs containment | [Docs](https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli) |
| Cursor Agent CLI | `cursor` / `agent` | <code>curl https://cursor.com/install -fsS &#124; bash</code> | Conditional; sandbox is not read-only | [Docs](https://cursor.com/docs/cli/overview) |
| ZCode CLI | `zcode` / bundled `zcode.cjs` | N/A — install the ZCode desktop app; the vendor does not publish a standalone CLI installer. | Local bundled 0.15.0; conditional | [Docs](https://zcode.z.ai/en/docs) |

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

## Boundaries

- Reviewers receive the same disclosed request and cannot see one another's same-round output.
- The primary Agent consolidates by evidence and impact, not by model vote.
- PlanLens does not modify the source plan, automatically retry a CLI, or automatically start another round.
- No monetary cost estimate is displayed.
- Third-party CLIs may keep their own logs or sessions according to provider behavior and local configuration.
- Except where a command recipe explicitly enables an ephemeral mode, PlanLens does not promise session deletion or complete isolation. Do not send sensitive material unless that boundary is acceptable.
- Kimi Code CLI must use the documented temporary no-tools custom agent. Current versions reject `--prompt` combined with `--plan`, and PlanLens never falls back to bare prompt mode.

## License

Apache License 2.0. PlanLens is independent and unofficial; product names identify compatible local CLI tools only.
