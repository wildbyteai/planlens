# PlanLens

PlanLens is an installable Agent Skill that asks one or more local AI CLIs to review the same plan, then lets the primary Agent summarize their findings and disagreements.

It is intentionally small: the workflow lives in `SKILL.md`. There is no PlanLens runtime, binary, Node command, hosted service, daemon, port, database, or state machine.

## Install

Install the complete [`skills/planlens`](skills/planlens) directory, including its `references` and `agents` subdirectories.

### Codex

Ask Codex:

```text
Use $skill-installer to install PlanLens from https://github.com/wildbyteai/planlens/tree/main/skills/planlens.
```

Invoke it on a later turn with `$planlens`.

### Claude Code

Copy `skills/planlens` to `~/.claude/skills/planlens` for a personal installation, or to `.claude/skills/planlens` inside one project. Invoke it with `/planlens`.

### Antigravity

Copy `skills/planlens` to `~/.gemini/config/skills/planlens` for a personal installation, or to `.agents/skills/planlens` inside one project. Restart the CLI if it was already open, then invoke `/planlens`.

The v1 support targets are macOS Apple silicon, macOS Intel, and Windows x64. Linux and Windows ARM64 are not v1 support targets.

## Use

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

PlanLens can call local installations of:

- Claude Code CLI: `claude`
- OpenAI Codex CLI: `codex`
- Antigravity CLI: `agy`
- Kimi Code CLI: `kimi`

PlanLens does not install, authenticate, update, bundle, or replace these CLIs. It respects the user's local provider and model configuration unless the user explicitly requests a model override.

## Profiles

Four small Markdown profiles are included:

- General plan
- Software design
- Implementation plan
- Security

Profiles guide the review; they are not schemas, executable plugins, or automatic approval rules.

## Output

When the project is writable, one round uses:

```text
.planlens/reviews/<timestamp>/
├── request.md
├── claude.md
├── codex.md
├── antigravity.md
├── kimi.md
└── summary.md
```

Only files for selected reviewers are created. A failed reviewer receives a matching `-error.md` file instead of a fabricated result.

## Boundaries

- Reviewers receive the same disclosed request and cannot see one another's same-round output.
- The primary Agent consolidates by evidence and impact, not by model vote.
- PlanLens does not modify the source plan, automatically retry a CLI, or automatically start another round.
- No monetary cost estimate is displayed.
- Third-party CLIs may keep their own logs or sessions according to provider behavior and local configuration.
- Kimi Code CLI does not currently provide a PlanLens-verified ephemeral guarantee; do not send sensitive material to Kimi unless that boundary is acceptable.

## License

Apache License 2.0. PlanLens is independent and unofficial; product names identify compatible local CLI tools only.
