# CLI recipe validation — 2026-07-26

This note records the evidence behind PlanLens's third-party CLI recipes. It is not loaded by the Skill at runtime.

## Method

- Scope: macOS ARM64 local validation plus first-party documentation/source review for the supported macOS and Windows x64 install paths.
- Local installations were isolated under a temporary directory. Existing `codex`, `claude`, `kimi`, `agy`, and the ZCode desktop app were inspected in place.
- Validation was limited to package provenance, installer behavior, version output, help output, argument parsing, configuration/source inspection, and filesystem side effects.
- No account login was started, no review prompt was sent, and no model was called.
- A recipe is `Strict` only when its documented preflight can remove tools or otherwise prevent execution of the reviewed plan. This is an internal PlanLens category, not vendor certification. `Conditional` means the CLI remains usable, but a weaker configuration, plan, ask, permission, or sandbox boundary must be disclosed.

## Validated matrix

| Reviewer | Version/build checked | Evidence | PlanLens status | Main boundary |
|---|---:|---|---|---|
| OpenAI Codex CLI | `0.146.0-alpha.3.1` | Existing local CLI; version/help and parser | Strict | Read-only sandbox, ephemeral session, ignored user config/rules, empty cwd |
| Claude Code | `2.1.218` | Existing local CLI; version/help | Strict | Safe Mode, empty built-in tool list, no session persistence |
| Kimi Code CLI | `0.29.1` | Existing local CLI; version/help and installed source | Strict when feature-gated | Temporary no-tools/no-subagents agent; hooks/session caveat remains |
| Gemini CLI | `0.52.0` | Official npm package in temporary prefix; version/help and source | Conditional | Deny-all admin policy plus isolated temporary CLI home |
| OpenCode | `1.18.5` | Official npm package/platform binary in temporary prefix | Conditional | Inline global deny, pure mode, isolated config/data roots; no no-session flag |
| Qwen Code | `0.21.0` | Official npm package in temporary prefix; parser and source | Strict | Safe Mode, Plan Mode, sandbox, zero tool-call limit |
| Pi | `0.82.1` | Official npm package in temporary prefix; version/help | Strict | Explicit no-tools, no-resources, no-session flags |
| goose | `1.44.0` | Official installer into temporary bin directory; version/help | Strict | No profile extensions and vendor no-session mode |
| Aider | `0.86.2` | Official PyPI package in temporary Python 3.12 venv; version/help | Strict | Ask Mode, dry-run, no repository automation, isolated config/history |
| Qoder CLI | `1.1.5` | Official checksum-verifying installer in temporary home; version/help/parser | Strict when feature-gated | Empty tool set, wildcard deny, no session persistence, isolated config |
| GitHub Copilot CLI | `1.0.75` | Official npm package/platform binary in temporary prefix; help/parser | Conditional | Empty available-tool set and temporary Copilot home; no no-session flag |
| Cline CLI | `3.0.46` | Official npm package/platform binary in temporary prefix; version/help | Conditional | Plan Mode plus isolated config/data/hooks; read/search/MCP capability remains |
| Cursor Agent CLI | `2026.07.23-e383d2b` | Official installer in temporary home; version/help | Conditional | Ask Mode plus deny-all CLI permissions; sandbox itself is read/write |
| Antigravity CLI | `1.1.7` | Existing local CLI; version/help | Conditional | Plan Mode and sandbox; read-capable tools remain |
| ZCode CLI | `0.15.0` in app `3.2.2` | Existing official desktop app bundle; version/help and bundled source | Conditional | Restrictive `zcode.json`, temporary storage/database; no standalone installer |
| Crush | `0.87.0` | Official macOS ARM64 release; SHA-256 matched release checksums; version/help/source | Conditional | Exact-version 29-tool deny list and temporary data; sessions are always created |
| Kilo Code CLI | `7.4.16` | Official npm package/platform binary in temporary prefix; version/help/source | Conditional | Global deny, Plan agent, isolated config/state, in-memory DB, effective-config preflight |

## Material command corrections

### Codex

Current builds expose `--ignore-user-config` and `--ignore-rules`. The canonical recipe uses both in addition to `--sandbox read-only`, `--ephemeral`, an empty working directory, and stdin input. Read-only is not described as no-tools.

### Kimi

Use feature detection for `--agent-file`; do not rely on a guessed minimum version. Current prompt mode rejects `--prompt` combined with `--plan`. The safe recipe supplies a temporary agent with `tools: []` and `subagents: []`, plus an empty skills directory. User lifecycle hooks and local retention remain possible and must be disclosed.

### Gemini

`-e none` is not an official disable-all sentinel; it is parsed as an extension name. Plan Mode still exposes read/search/web/MCP/skill-related tools and writes a plan file. The recipe therefore uses a deny-all admin policy, checks for a higher-precedence system policy, isolates the CLI home, and remains conditional.

### OpenCode

`--pure` disables external plugins; it does not disable configuration or sessions. During validation, redirecting only `OPENCODE_CONFIG_DIR` still caused a write attempt under the ordinary XDG configuration path. The recipe therefore isolates `HOME`, `XDG_CONFIG_HOME`, config, data, state, and cache roots together. OpenCode has no no-session flag and remains conditional.

### Qwen

The checked package accepts the hidden `--approval-mode plan`, `--max-tool-calls 0`, and `--max-wall-time 20m` controls. `--max-tool-calls 0` is the fail-closed no-tool boundary; every installed version must parse it before selection.

### Pi

The canonical model override is expressed as separate `--provider` and `--model` arguments. The CLI also accepts provider-prefixed model IDs, but the separate form is less ambiguous. Keep every no-resource flag and `--no-session`.

### Qoder

Version `1.1.5` materially improves the recipe: local help exposes `--tools ""`, `--no-session-persistence`, `--strict-mcp-config`, `--mcp-config`, `--config-dir`, and `--cwd`. A temporary config root plus token-based authentication prevents loading ordinary hooks, plugins, skills, MCP, memory, and session state. If any control is absent, PlanLens fails the reviewer rather than weakening the command.

### Copilot

`--available-tools=` hides every tool outside the empty list and is stronger than an incomplete family deny string. The checked vector also disables custom instructions, built-in MCP, temp-directory access, updates, and remote export/control. Copilot still lacks strict no-session persistence and remains conditional.

### Cline, Cursor, and Antigravity

Their Plan/Ask modes are not equivalent to no-tools. Cline retains read/search/MCP behavior; Cursor print mode advertises access to all tools and its sandbox permits workspace reads/writes; Antigravity Plan Mode retains read-capable tools. Their recipes rely on additional containment and remain conditional.

### Aider

`-m` means `--message`, not model. Use `--model` for an override. Ask Mode and dry-run are combined with empty config/env files, temporary input/chat/LLM histories, no Git automation, no URL detection, no browser/Playwright, no analytics, and no update checks.

### ZCode

The vendor does not publish a standalone CLI installer. The checked macOS app bundles `zcode.cjs`; local source confirms `ZCODE_MODEL`, `ZCODE_STORAGE_DIR`, and `ZCODE_SESSION_DB_PATH`. A restrictive `zcode.json` can disable tools, MCP, plugins, skills, hooks, memory, and subagents, but sessions are only redirected, not disabled.

### Crush

Version `0.87.0` has 29 built-in tool names and no wildcard deny for `options.disabled_tools`. The recipe is exact-version only, checks `/etc/crush/crush.json`, disables every known built-in plus MCP/hooks/skills/LSP, and redirects the unavoidable session database.

### Kilo

The Plan agent alone is insufficient. The recipe adds global permission deny, empty MCP/plugins/instructions, project/Claude/external-skill isolation, `KILO_DB=:memory:`, empty cwd, and a non-model `kilo debug config` preflight. Even `--help` initialized local database state without isolation during validation, confirming that temporary state handling is required.

## First-party sources

- Codex: [CLI reference](https://developers.openai.com/codex/cli/reference/)
- Claude Code: [CLI reference](https://code.claude.com/docs/en/cli-reference)
- Kimi Code: [official repository](https://github.com/MoonshotAI/kimi-code)
- Gemini CLI: [official repository](https://github.com/google-gemini/gemini-cli), [Plan Mode](https://geminicli.com/docs/cli/plan-mode/), [policy engine](https://geminicli.com/docs/reference/policy-engine/)
- OpenCode: [CLI](https://opencode.ai/docs/cli/), [configuration](https://opencode.ai/docs/config/), [permissions](https://opencode.ai/docs/permissions/)
- Qwen Code: [CLI configuration](https://qwenlm.github.io/qwen-code-docs/en/users/configuration/cli/), [Safe Mode](https://qwenlm.github.io/qwen-code-docs/en/users/features/safe-mode/)
- Pi: [official repository](https://github.com/earendil-works/pi)
- goose: [official repository and installation](https://github.com/aaif-goose/goose)
- Aider: [installation](https://aider.chat/docs/install.html), [options](https://aider.chat/docs/config/options.html)
- Qoder CLI: [quick start](https://docs.qoder.com/cli/quick-start), [permissions](https://docs.qoder.com/cli/permissions), [hooks](https://docs.qoder.com/cli/hooks)
- GitHub Copilot CLI: [installation](https://docs.github.com/en/copilot/how-tos/copilot-cli/install-copilot-cli), [CLI reference](https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-programmatic-reference)
- Cline CLI: [CLI reference](https://docs.cline.bot/cli/cli-reference), [Plan and Act](https://docs.cline.bot/features/plan-and-act)
- Cursor Agent CLI: [installation](https://cursor.com/docs/cli/installation), [parameters](https://cursor.com/docs/cli/reference/parameters), [permissions](https://cursor.com/docs/cli/reference/permissions)
- Antigravity CLI: [official installation](https://antigravity.google/docs/cli/install)
- ZCode: [official documentation](https://zcode.z.ai/en/docs)
- Crush: [release `v0.87.0`](https://github.com/charmbracelet/crush/releases/tag/v0.87.0)
- Kilo Code CLI: [official repository](https://github.com/Kilo-Org/kilocode), [CLI documentation](https://kilo.ai/docs/code-with-ai/platforms/cli)
