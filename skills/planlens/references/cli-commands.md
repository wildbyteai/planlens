# Local CLI commands

Use these as argument-vector templates, not shell strings to copy blindly. The common confirmation, isolation, timeout, failure, and output rules live in `SKILL.md`.

Reviewers: [Claude](#claude-code-cli), [Codex](#openai-codex-cli), [Antigravity](#antigravity-cli), [Kimi](#kimi-code-cli), [Qoder](#qoder-cli), [Copilot](#github-copilot-cli), [OpenCode](#opencode-cli), [Cursor](#cursor-agent-cli), [ZCode](#zcode-cli), [Gemini](#gemini-cli), [Pi](#pi), [Cline](#cline-cli), [goose](#goose), [Aider](#aider), [Crush](#crush), [Kilo](#kilo-code-cli), and [Qwen](#qwen-code).

## Reviewer IDs

`Strict` means the documented non-interactive recipe can disable tools or otherwise prevent plan execution when every feature preflight passes. It does not promise zero local retention or vendor certification. `Conditional` means the CLI can be called, but the boundary depends on weaker plan, ask, configuration, permission, or sandbox controls and must be disclosed before confirmation.

| Reviewer ID | Command | Compatibility |
|---|---|---|
| `codex` | `codex` | Strict; default |
| `claude` | `claude` | Strict; default |
| `kimi` | `kimi` | Strict when the no-tools custom-agent feature is present; default |
| `gemini` | `gemini` | Conditional; preferred fourth/fallback |
| `opencode` | `opencode` | Conditional |
| `qwen` | `qwen` | Strict |
| `pi` | `pi` | Strict |
| `goose` | `goose` | Strict |
| `aider` | `aider` | Strict |
| `qoder` | `qodercli` | Strict when the no-tools, no-session, and isolated-config flags are present |
| `copilot` | `copilot` | Conditional |
| `cline` | `cline` | Conditional |
| `cursor` | `agent` | Conditional |
| `antigravity` | `agy` | Conditional |
| `zcode` | bundled `zcode.cjs` | Conditional |
| `crush` | `crush` | Conditional |
| `kilo` | `kilo` | Conditional |

## Claude Code CLI

Command: `claude`

Claude must access its configured API endpoint, network, and normal authentication or provider settings. If the host's outer sandbox blocks any of those required capabilities, require an execution mode outside that outer sandbox. In Codex, escalated process execution removes the outer sandbox from the entire `claude` child process and gives it normal user-level filesystem and network access. Disclose that full scope before confirmation. Keep `--safe-mode`, `--tools ""`, `--no-session-persistence`, and the empty temporary working directory. If the host cannot provide the required execution mode, treat Claude as unavailable before preview. If the user later declines the host's platform permission, record the reviewer as failed.

Argument template:

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

Save stdout as `claude.md`. Safe mode disables customizations including plugins, hooks, skills, and MCP servers; `--tools ""` disables built-in tools, while `--no-session-persistence` prevents session storage. The user's authentication and selected model remain available.

## OpenAI Codex CLI

Command: `codex`

Codex must access its normal `CODEX_HOME` authentication and state and initialize its in-process app-server client. If the host's outer sandbox makes that state read-only or blocks the client, require an execution mode outside that outer sandbox. For a nested Codex launch, escalated process execution removes the outer sandbox from the entire child process and gives it normal user-level filesystem and network access. Disclose that full scope before confirmation. Keep the child Codex `--sandbox read-only`, `--ephemeral`, ignored config and rules, and empty temporary working directory, but do not describe the whole process as completely isolated. Do not redirect `CODEX_HOME` to an empty directory merely to isolate state, because that also hides authentication. If the host cannot provide the required execution mode, treat Codex as unavailable before preview. If the user later declines the host's platform permission, record the reviewer as failed.

Argument template:

```text
codex --ask-for-approval never exec
  --sandbox read-only
  --ephemeral
  --ignore-user-config
  --ignore-rules
  --skip-git-repo-check
  -c project_doc_max_bytes=0
  -C <empty-temporary-directory>
  [-m <explicit-user-choice>]
  -
```

Send `request.md` through stdin. Capture the final response as `codex.md`. `--ignore-user-config` keeps authentication but skips user config; `--ignore-rules` skips user and project exec-policy rules. If either flag is unavailable, do not silently drop it. If supported, `--output-last-message` may write the final response directly to `codex.md`.

## Antigravity CLI

Command: `agy`

Before selection, require local help to expose `--mode`, `--sandbox`, `--print`, `--print-timeout`, and `--log-file`.

Antigravity starts a local language server even in print mode. It must bind random loopback ports and access its normal user authentication and app-data directory. If the host's normal process environment permits both, use it. If the host's outer sandbox blocks either capability, require an execution mode outside that outer sandbox. In Codex, escalated process execution removes the outer sandbox from the entire `agy` child process, giving it normal user-level filesystem and network access, including normal Antigravity app-data and authentication state. Disclose that full scope before confirmation. Keep the empty temporary working directory and Antigravity's own `--sandbox` enabled, but do not describe them as complete isolation because Plan Mode retains read-capable tools. Do not redirect `HOME` or Antigravity's app-data directory to an empty location merely to isolate state, because that also hides its authentication. If the host cannot provide the required execution mode, treat Antigravity as unavailable before preview. If the user later declines the host's platform permission, record the reviewer as failed.

Argument template:

```text
agy
  --mode plan
  --sandbox
  --print-timeout 20m
  --log-file <temporary-log-file>
  [--model <explicit-user-choice>]
  --print <request-text>
```

Save stdout as `antigravity.md`. `--print` is the single-prompt mode; do not use `--continue` or `--conversation`. Plan Mode retains read-capable tools, and access to normal Antigravity state weakens isolation, so keep this reviewer conditional. If the process exits non-zero or stdout is empty, inspect the temporary log and include a redacted error excerpt in `antigravity-error.md`; do not retry.

## Kimi Code CLI

Command: `kimi`

Before selection, require `kimi --help` to expose `--agent-file`; use the feature check rather than assuming a version string is sufficient. Create a temporary Markdown agent file:

```markdown
---
name: planlens-reviewer
description: Review the supplied plan without tools or delegation
tools: []
subagents: []
---

Review only the supplied PlanLens request. Do not use tools, read files, retrieve external material, delegate work, or perform the plan. Return only the final review.
```

Set this variable on the child process:

```text
KIMI_CODE_EXPERIMENTAL_FLAG=1
```

Argument template:

```text
kimi
  --prompt <request-text>
  --agent-file <temporary-agent-file>
  --skills-dir <empty-skills-directory>
  --output-format text
  [--model <explicit-user-choice>]
```

Save stdout as `kimi.md`, then remove the temporary agent file and skills directory. The explicit agent has no tools or subagents and replaces the default system prompt for this launch. Kimi may still retain local session data or run user-configured lifecycle hooks; disclose that caveat. Do not use `--plan` with `--prompt`; current versions reject that combination. Do not fall back to bare prompt mode without the no-tools agent.

## Qoder CLI

Command: `qodercli`

Before selection, require local help to expose all of these controls: `--tools`, `--no-session-persistence`, `--strict-mcp-config`, `--mcp-config`, `--config-dir`, `--cwd`, and `--output-format`. Use an empty temporary config directory and authenticate through `QODER_PERSONAL_ACCESS_TOKEN`; do not load the ordinary user configuration merely to reuse a saved login.

Argument template:

```text
qodercli
  --print
  --permission-mode dont_ask
  --tools ""
  --disallowed-tools "*"
  --no-session-persistence
  --strict-mcp-config
  --mcp-config '{"mcpServers":{}}'
  --config-dir <empty-temporary-config-directory>
  --cwd <empty-temporary-working-directory>
  --output-format text
  [--model <explicit-user-choice>]
  <request-text>
```

Save stdout as `qoder.md`. The empty tool list is the primary no-tools boundary; the wildcard deny is defense in depth. The temporary config directory also contains hooks, MCP, plugins, skills, memory, and state. If the installed CLI lacks any required control or cannot authenticate without loading ordinary configuration, record Qoder as unavailable instead of weakening the recipe.

## GitHub Copilot CLI

Command: `copilot`

Before selection, require local help to expose `--available-tools`, `--mode`, `--no-custom-instructions`, `--disable-builtin-mcps`, `--disallow-temp-dir`, `--no-auto-update`, `--no-remote`, and `--no-remote-export`. Set `COPILOT_HOME` to a fresh temporary directory, `COPILOT_PLUGIN_DIR_ONLY=1`, and route logs to a temporary directory. Authentication must come from a supported token environment variable; do not load the ordinary user home merely to reuse saved state.

Argument template:

```text
copilot
  --available-tools=
  --mode plan
  --no-custom-instructions
  --disable-builtin-mcps
  --disallow-temp-dir
  --no-auto-update
  --no-remote
  --no-remote-export
  --no-ask-user
  --log-dir <temporary-log-directory>
  --log-level none
  --output-format text
  --no-color
  --silent
  [--model <explicit-user-choice>]
  --prompt <request-text>
```

Save stdout as `copilot.md`. An empty `--available-tools` value hides all tools from the model; do not replace it with an incomplete family deny list. Never pass an allow flag, `--allow-all`, `--yolo`, `--continue`, `--resume`, or a session ID. Copilot has no strict no-session switch, so this reviewer remains conditional even with temporary state containment.

## OpenCode CLI

Set these variables on the child process without constructing a shell string:

```text
OPENCODE_CONFIG_CONTENT={"permission":"deny","share":"disabled"}
OPENCODE_DISABLE_DEFAULT_PLUGINS=1
OPENCODE_DISABLE_CLAUDE_CODE=1
OPENCODE_CONFIG_DIR=<empty-temporary-config-directory>
HOME=<temporary-home-directory>
XDG_CONFIG_HOME=<empty-temporary-xdg-config-directory>
XDG_DATA_HOME=<temporary-data-directory>
XDG_STATE_HOME=<temporary-state-directory>
XDG_CACHE_HOME=<temporary-cache-directory>
```

Command: `opencode`

Argument template:

```text
opencode
  --pure
  run
  --format default
  --dir <empty-temporary-working-directory>
  [--model <provider/model>]
  <request-text>
```

Save stdout as `opencode.md`. `--pure` disables external plugins; it does not disable configuration or sessions. OpenCode merges configuration sources and has no no-session flag, so verify the temporary roots are effective on the installed version and keep this reviewer conditional. If provider authentication exists only in the ordinary OpenCode data directory, do not load that directory without disclosing the weaker boundary.

## Cursor Agent CLI

Command: `agent`

Set `CURSOR_CONFIG_DIR` to a fresh temporary directory and authenticate with `CURSOR_API_KEY`. Write `<CURSOR_CONFIG_DIR>/cli-config.json` before launch:

```json
{
  "version": 1,
  "editor": {"vimMode": false},
  "permissions": {
    "allow": [],
    "deny": [
      "Shell(*)",
      "Read(**)",
      "Read(/**)",
      "Write(**)",
      "Write(/**)",
      "WebFetch(*)",
      "Mcp(*:*)"
    ]
  }
}
```

Argument template:

```text
agent
  --print
  --mode ask
  --sandbox enabled
  --output-format text
  --workspace <empty-temporary-working-directory>
  --trust
  [--model <explicit-user-choice>]
  <request-text>
```

Save stdout as `cursor.md`. Print mode explicitly has access to tools, and Cursor's sandbox permits workspace reads and writes; the deny rules are the actual boundary. Never use `--force`, `--yolo`, `--approve-mcps`, `--continue`, `--resume`, or a session identifier. Cursor has no strict no-session switch, so keep it conditional.

## ZCode CLI

Command: vendor-bundled `zcode.cjs`. On the verified macOS app, the path is `/Applications/ZCode.app/Contents/Resources/glm/zcode.cjs`; the vendor does not publish a standalone CLI installer or guarantee a `zcode` command on `PATH`.

Set `ZCODE_STORAGE_DIR` and `ZCODE_SESSION_DB_PATH` to temporary paths. When the user explicitly selects a model, set `ZCODE_MODEL` on the child process. Put this `zcode.json` in the empty temporary working directory:

```json
{
  "permission": {
    "mode": "plan",
    "allowedTools": [],
    "disallowedTools": ["*"],
    "autoApproveHighRisk": false
  },
  "features": {"subagent": false, "memory": false, "skill": false, "mcp": false},
  "memory": {"use": false, "write": false, "autoConsolidate": false},
  "mcp": {"servers": {}},
  "plugins": {"enabled": false, "dirs": [], "enabledPlugins": {}},
  "skills": {"enabled": false, "includeInstructions": false, "roots": []},
  "hooks": {"enabled": false, "events": {}}
}
```

Argument template:

```text
<verified-zcode.cjs-path>
  --prompt <request-text>
  --mode plan
  --cwd <empty-temporary-directory>
  --no-color
```

Save stdout as `zcode.md`. Do not use `--continue`, `--resume`, `app-server`, or `agent-server`. The inspected bundled CLI 0.15.0 supports `ZCODE_MODEL` but has no no-session flag, so temporary storage is containment rather than vendor-guaranteed ephemerality. Keep this reviewer conditional and require a local path/version feature check.

## Gemini CLI

Command: `gemini`

Set `GEMINI_CLI_HOME` and the child process's config, data, state, and cache roots to fresh temporary directories. This contains settings, hooks, extensions, MCP configuration, skills, plans, and session artifacts; it may require environment-based authentication instead of an existing OAuth login.

Create a temporary admin policy:

```toml
[[rule]]
toolName = "*"
decision = "deny"
priority = 999
```

Before selection, check the standard system admin-policy directory for the current OS. If it contains any `.toml` policy, Gemini ignores the supplemental `--admin-policy`; treat the reviewer as unavailable unless an independently verified outer sandbox supplies the boundary.

Argument template:

```text
gemini
  --approval-mode plan
  --sandbox
  --admin-policy <temporary-deny-all-policy.toml>
  --output-format text
  [-m <explicit-user-choice>]
  -p <request-text>
```

Save stdout as `gemini.md`, then remove all temporary state. Do not pass `-e none`: `none` is parsed as an extension name, not a documented empty-list sentinel. Plan Mode still exposes read, search, web, MCP, skill, and memory-related tools and writes a plan file; the deny-all policy is the actual tool boundary. The policy and temporary-home preflights are mandatory, and this reviewer remains conditional.

## Pi

Command: `pi`

Set `PI_TELEMETRY=0` on the child process.

Argument template:

```text
pi
  --no-session
  --no-tools
  --no-extensions
  --no-skills
  --no-prompt-templates
  --no-context-files
  --no-approve
  --mode text
  --print
  [--provider <provider> --model <model>]
  <request-text>
```

Save stdout as `pi.md`. Keep every no-resource flag: Pi otherwise inherits the launching process's permissions and broader local resources. `--no-session` disables session persistence. Prefer separate `--provider` and `--model` arguments for an explicit override; omit both to preserve the user's configured default.

## Cline CLI

Command: `cline`

Set this variable on the child process:

```text
CLINE_COMMAND_PERMISSIONS={"allow":[],"deny":["*"],"allowRedirects":false}
```

Argument template:

```text
cline
  --plan
  --auto-approve false
  --cwd <empty-temporary-directory>
  --hooks-dir <empty-hooks-directory>
  --config <empty-config-directory>
  --data-dir <temporary-data-directory>
  --retries 0
  --timeout 1200
  [--model <explicit-user-choice>]
  <request-text>
```

Save stdout as `cline.md`. The command rule denies shell commands, and the empty config, hooks, data, and working directories contain ordinary local state. Plan Mode still allows read/search behavior and MCP inspection; it is not a zero-tools or OS read-only boundary. Do not use an auto-approval shortcut. Cline persists task state inside the temporary data directory, so keep it conditional and remove the directory after capture.

## goose

Command: `goose`

Argument template:

```text
goose run
  --no-profile
  --no-session
  --quiet
  --output-format text
  [--provider <explicit-user-choice> --model <explicit-user-choice>]
  --text <request-text>
```

Save stdout as `goose.md`. `--no-profile` prevents the default profile's extensions from loading; do not claim that it disables every possible customization source, and do not add any `--with-*` extension. `--no-session` is the vendor no-session control. Do not use `goose review --prompt`: that command reviews code ranges and its prompt argument is a file path.

## Aider

Command: `aider`

Argument template:

```text
aider
  --chat-mode ask
  --message <request-text>
  --dry-run
  --no-git
  --no-gitignore
  --no-add-gitignore-files
  --no-auto-commits
  --no-auto-lint
  --no-auto-test
  --no-suggest-shell-commands
  --no-detect-urls
  --no-browser
  --disable-playwright
  --map-tokens 0
  --config <empty-temporary-config-file>
  --env-file <empty-temporary-env-file>
  --chat-history-file <temporary-history-file>
  --input-history-file <temporary-input-history-file>
  --llm-history-file <temporary-llm-history-file>
  --no-restore-chat-history
  --no-analytics
  --no-check-update
  --no-show-release-notes
  --no-pretty
  --no-stream
  [--model <explicit-user-choice>]
```

Save stdout as `aider.md`, then remove the temporary config, environment, and history files. Use the full `--model` option: in Aider, `-m` means `--message`, not model. Ask mode and dry-run prevent source edits; the other flags suppress repository discovery, automation, URL handling, analytics, updates, and ambient history.

## Crush

Command: `crush`

Support this recipe only for exact `crush version v0.87.0`. Abort if `/etc/crush/crush.json` exists unless its effective settings have been independently reviewed. Set `CRUSH_GLOBAL_CONFIG`, `CRUSH_GLOBAL_DATA`, `CRUSH_SKILLS_DIR`, and `XDG_CONFIG_HOME` to temporary paths. Write this exact-version config to `CRUSH_GLOBAL_CONFIG`:

```json
{
  "$schema": "https://charm.land/crush.json",
  "options": {
    "disabled_tools": [
      "agent", "bash", "crush_info", "crush_logs", "job_output",
      "job_kill", "download", "edit", "multiedit", "lsp_diagnostics",
      "lsp_references", "lsp_restart", "lsp_symbols", "lsp_definition",
      "lsp_call_hierarchy", "lsp_rename", "lsp_replace_symbol", "fetch",
      "agentic_fetch", "glob", "grep", "ls", "question", "sourcegraph",
      "todos", "view", "write", "list_mcp_resources", "read_mcp_resource"
    ],
    "disabled_skills": ["crush-config"],
    "auto_lsp": false,
    "skills_paths": []
  },
  "mcp": {},
  "hooks": {},
  "lsp": {}
}
```

Conditional argument template:

```text
crush run
  --quiet
  --cwd <empty-temporary-working-directory>
  --data-dir <temporary-data-directory>
  [--model <provider/model>]
  <request-text>
```

Run only when the user explicitly accepts the weaker boundary. Version 0.87.0 has no wildcard deny for built-in tools, so a different version may add a tool that escapes this list; fail the reviewer instead of reusing the list. `crush run` always creates a session inside the temporary data directory. Do not use global `--yolo`, `--continue`, or `--session`.

## Kilo Code CLI

Command: `kilo`

Set these variables on the child process:

```text
KILO_DB=:memory:
KILO_CONFIG_CONTENT={"permission":"deny","plugin":[],"mcp":{},"instructions":[]}
KILO_DISABLE_DEFAULT_PLUGINS=true
KILO_DISABLE_PROJECT_CONFIG=true
KILO_DISABLE_CLAUDE_CODE_PROMPT=true
KILO_DISABLE_EXTERNAL_SKILLS=true
XDG_CONFIG_HOME=<empty-temporary-config-directory>
XDG_DATA_HOME=<temporary-data-directory>
XDG_STATE_HOME=<temporary-state-directory>
XDG_CACHE_HOME=<temporary-cache-directory>
```

Before any model call, run `kilo debug config` with the same environment and abort unless the effective configuration still has global permission deny, empty MCP/plugins/instructions, and all isolation switches honored.

Conditional argument template:

```text
kilo
  --pure
  run
  --agent plan
  --format default
  --dir <empty-temporary-working-directory>
  [--model <provider/model>]
  <request-text>
```

Run only when the user explicitly accepts the weaker boundary. The built-in Plan agent alone is not the boundary; the effective global deny and isolated roots are. Never pass `--continue`, `--session`, `--fork`, `--share`, `--auto`, or `--dangerously-skip-permissions`. Kilo may still initialize local state, so keep it conditional.

## Qwen Code

Command: `qwen`

Argument template:

```text
qwen
  --safe-mode
  --approval-mode plan
  --sandbox
  --max-tool-calls 0
  --max-wall-time 20m
  --output-format text
  [-m <explicit-user-choice>]
  -p <request-text>
```

Save stdout as `qwen.md`. Safe mode disables project context files, hooks, extensions, skills, MCP servers, custom subagents, and configuration-based permission overrides. Plan Mode is read-only, while `--max-tool-calls 0` is the fail-closed no-tool control. The latter may be hidden from abbreviated help, so feature-check the installed parser and fail if it is rejected. Do not use `--yolo`.
