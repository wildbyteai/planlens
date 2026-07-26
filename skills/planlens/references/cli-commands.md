# Local CLI invocation recipes

Use these as argument-vector templates, not as shell strings to copy blindly. Pass the complete `request.md` text through the host's process tool or standard input without interpolating it into an unquoted command.

Run each CLI from a new temporary working directory that contains no project files. Write the final response to the review output directory. Remove the temporary directory after the process exits.

## Common behavior

- Check whether the command exists on `PATH`; report it as unavailable when missing.
- Use one fresh non-interactive invocation per reviewer.
- Apply a reasonable timeout; default to 20 minutes when the host requires a value.
- Keep the user's existing authentication and default model.
- Add a model flag only when the user explicitly selected a model.
- Capture stdout and stderr separately when the host supports it; otherwise preserve combined output and label it as combined.
- Treat non-zero exit, timeout, cancellation, or empty final output as failure.
- Do not automatically retry, continue a session, or switch to another reviewer.

## Claude Code CLI

Command: `claude`

Argument template:

```text
claude
  --print
  --no-session-persistence
  --no-chrome
  --disable-slash-commands
  --permission-mode dontAsk
  --tools ""
  --output-format text
  [--model <explicit-user-choice>]
  <request-text>
```

Save stdout as `claude.md`. This preserves the user's configured provider and default model, including a deliberate local third-party model configuration.

## OpenAI Codex CLI

Command: `codex`

Argument template:

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

Send `request.md` through stdin. Capture the final response as `codex.md`. If the installed Codex supports `--output-last-message`, it may be used to write the final response directly to that file.

## Antigravity CLI

Command: `agy`

Argument template:

```text
agy
  --mode plan
  --sandbox
  --print-timeout 20m
  [--model <explicit-user-choice>]
  --print <request-text>
```

Save stdout as `antigravity.md`. `--print` is the single-prompt mode; do not use `--continue` or `--conversation`.

## Kimi Code CLI

Command: `kimi`

Create an empty skills directory inside the temporary working directory, then use:

```text
kimi
  --plan
  --skills-dir <empty-skills-directory>
  [--model <explicit-user-choice>]
  --output-format text
  --prompt <request-text>
```

Save stdout as `kimi.md`. Do not use `--continue` or `--session`.

Kimi Code CLI may persist session or diagnostic data according to its current implementation and the user's local configuration. PlanLens does not promise ephemeral execution for Kimi and must not send sensitive material to it unless the user accepts that boundary.

## Output handling

For a successful invocation, keep only the final reviewer response and any short warning needed to interpret it. Do not merge raw outputs before all selected reviewers have finished.

For a failed invocation, write `<reviewer>-error.md`:

```markdown
# Reviewer failure

- Reviewer: <name>
- Command: <command name>
- Status: unavailable | failed | timed out | cancelled | empty output
- Exit code: <number or unavailable>
- Error: <concise stderr excerpt without credentials>
```

Redact tokens, cookies, account identifiers, private URLs, and other credentials from all saved errors.
