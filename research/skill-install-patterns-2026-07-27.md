# Open-source Agent Skill installation patterns

Date: 2026-07-27

## Question

How do established open-source Agent Skill projects structure and distribute their skills, and which installation pattern should PlanLens adopt?

## Conclusion

Large projects increasingly separate two layers:

1. A portable Skill uses the Agent Skills directory convention, normally `skills/<skill-name>/SKILL.md` plus optional references, scripts, and assets.
2. A generic installer discovers that directory and installs it into the selected host. The repository URL identifies the source; users normally do not copy a long GitHub `tree/<tag>/...` URL.

For PlanLens, the best primary entry is the same generic Skills CLI pattern used by `mattpocock/skills`:

```bash
npx skills@latest add wildbyteai/planlens
```

This preserves PlanLens as a small, runtime-free Skill and gives users a short cross-Agent command. Codex's built-in `$skill-installer` is the no-Node alternative, while GitHub CLI remains the advanced path for exact release pinning and source-tracked updates.

## Representative projects

| Project | Main pattern | Version pinning | Fit for PlanLens |
|---|---|---|---|
| `mattpocock/skills` | `npx skills@latest add owner/repo` | Primary quickstart follows the current repository version | Strong model for PlanLens's short primary command |
| `github/awesome-copilot` | Standard `skills/` tree; `gh skill install owner/repo skill` | `--pin <tag-or-sha>` | Strong model for PlanLens's advanced pinned command |
| `K-Dense-AI/scientific-agent-skills` | Standard `skills/` tree; both `gh skill install` and `npx skills add` | Release tags or commit SHA | Strong model for releases and reproducible installation |
| `vercel-labs/skills` | Cross-agent `npx skills add owner/repo` installer | No equally clear tag/SHA example in its main README | Strong primary installer when Node.js is available |
| `anthropics/skills` | Claude Code Plugin Marketplace | Marketplace/plugin versioning | Appropriate for Claude-specific bundles, not PlanLens's portable core |
| `obra/superpowers` | Separate native installation path for each host/plugin system | Host-dependent | Appropriate for a full workflow package, heavier than PlanLens needs |

### GitHub Awesome Copilot

The repository keeps reusable skills under `skills/` and documents GitHub CLI as the preferred cross-agent installation path:

```bash
gh skill install github/awesome-copilot <skill-name>
```

It also supports `npx skills add`, manual copying, and host-specific plugin packages. This is a useful distinction: portable skills use the standard Skill installer, while richer Copilot integrations use the Copilot plugin layer.

Sources:

- <https://github.com/github/awesome-copilot>
- <https://github.com/github/awesome-copilot/blob/main/docs/README.skills.md>
- <https://cli.github.com/manual/gh_skill_install>

### K-Dense Scientific Agent Skills

This project uses the standard `skills/` directory, publishes tagged releases, and shows both universal installers:

```bash
npx skills add K-Dense-AI/scientific-agent-skills
gh skill install K-Dense-AI/scientific-agent-skills
```

It explicitly demonstrates reproducible installation with a release tag or commit SHA. This is the closest large-project model for PlanLens's desired combination of portability, simple installation, and version locking.

Sources:

- <https://github.com/K-Dense-AI/scientific-agent-skills>
- <https://github.com/K-Dense-AI/scientific-agent-skills/releases>

### Vercel Skills CLI

Vercel's installer provides a concise cross-agent command and auto-detects supported hosts:

```bash
npx skills add <owner/repo>
```

It is broadly compatible and convenient, but requires Node/npm and does not make immutable release pinning as prominent as GitHub CLI. PlanLens can use it as the short primary path while documenting a pinned alternative.

Source:

- <https://github.com/vercel-labs/skills>

### Matt Pocock Skills

The project's README makes this its primary quickstart:

```bash
npx skills@latest add mattpocock/skills
```

The installer then asks which skills and supported Agents to target. The README also offers `--all`, but does not make a pinned Git tag part of the primary quickstart. It prioritizes a short current-version install over an immutable command.

PlanLens's matching commands are:

```bash
npx skills@latest add wildbyteai/planlens
npx skills@latest add wildbyteai/planlens --skill planlens --agent codex --global --yes
```

Sources:

- <https://github.com/mattpocock/skills>
- <https://skills.sh/mattpocock/skills>

### Anthropic Skills

Anthropic distributes its larger collection through the Claude Code Plugin Marketplace:

```text
/plugin marketplace add anthropics/skills
/plugin install example-skills@anthropic-agent-skills
```

This works well when a package needs Claude-specific plugin behavior. It is not the best primary path for a Skill intended to work across Codex, Claude Code, Antigravity, Kimi, Qoder, and other hosts.

Source:

- <https://github.com/anthropics/skills>

### Superpowers

Superpowers provides native installation instructions for several host ecosystems, including Claude Code, Cursor, Codex, OpenCode, and Gemini CLI. This is appropriate because Superpowers is a broader workflow/plugin system rather than one small portable Skill.

PlanLens should not copy this complexity unless it later needs host-specific hooks, commands, MCP servers, or lifecycle behavior that cannot live in a standard Skill.

Source:

- <https://github.com/obra/superpowers>

## Relevant standards and GitHub CLI behavior

The Agent Skills specification defines a Skill as a directory with a required `SKILL.md`. Optional folders such as `scripts/`, `references/`, and `assets/` are loaded as needed. PlanLens already follows the important part of this convention:

```text
skills/planlens/
├── SKILL.md
├── agents/
└── references/
```

GitHub CLI discovers `skills/*/SKILL.md`, supports host-specific user or project installation, can pin a tag or commit, injects source metadata for later updates, and can validate/publish Agent Skills. The command remains marked as preview as of this review.

Sources:

- <https://agentskills.io/specification>
- <https://cli.github.com/manual/gh_skill>
- <https://cli.github.com/manual/gh_skill_install>
- <https://cli.github.com/manual/gh_skill_publish>

## Recommended PlanLens installation hierarchy

### 1. Primary: Skills CLI

Interactive cross-Agent installation:

```bash
npx skills@latest add wildbyteai/planlens
```

Direct Codex user installation:

```bash
npx skills@latest add wildbyteai/planlens --skill planlens --agent codex --global --yes
```

This is the shortest cross-Agent path, but it requires Node.js and follows the current default branch when no immutable source ref is supplied.

### 2. Codex without Node.js

Ask Codex:

```text
Use $skill-installer to install skills/planlens from the wildbyteai/planlens repository at tag v1.1.2.
```

This uses the system Skill already included with Codex and does not require Node.js or GitHub CLI.

### 3. Version-pinned: GitHub CLI

```bash
gh skill install wildbyteai/planlens planlens@v1.1.2 --agent codex --scope user
```

Use this when a recent GitHub CLI is already installed and the exact release must be pinned. Change `--agent` or `--scope` for another supported target.

### 4. Final fallback: manual copy

Keep manual copying for older or unsupported hosts, but do not make it the first instruction.

## Publishing recommendation

For each PlanLens release:

1. Keep the portable Skill under `skills/planlens/`.
2. Validate locally with `gh skill publish --dry-run`.
3. Publish a semantic Git tag and GitHub Release.
4. Keep the short current-version command first and an explicit tag in the reproducible alternatives.
5. Add the repository topic `agent-skills`.
6. Keep host-specific adapters optional; do not introduce a PlanLens runtime merely for installation.

## Local verification performed

The following historical release checks were verified on this machine with GitHub CLI 2.95.0:

- `gh skill preview wildbyteai/planlens planlens@v1.1.1` succeeded.
- An isolated `gh skill install ... --dir <temporary-directory>` succeeded.
- The installer selected tag `v1.1.1`, installed `SKILL.md`, `agents/`, and `references/`, and injected repository/ref metadata.
- `gh skill publish --dry-run` passed; its recommendation to add a `license` field was addressed during the `v1.1.2` release preparation.

The `v1.1.2` release-candidate checks also verified:

- `npx skills@latest add wildbyteai/planlens --skill planlens --agent codex --yes --copy` completed in an isolated project and installed `SKILL.md`, `agents/`, and all `references/` files.
- Codex's built-in installer helper completed an isolated install of the published `v1.1.1` tree with the same complete file set; the exact `v1.1.2` tag is rechecked after publication.
- Complete PlanLens review rounds were exercised manually in Codex and produced `request.md`, reviewer outputs, and `summary.md` artifacts.
- `gh skill publish --dry-run` passed without warnings after the `license: Apache-2.0` frontmatter was added.

## Decision for PlanLens

Replace the long GitHub subtree URL as the main README installation path. Use `npx skills@latest add wildbyteai/planlens` first, Codex's `$skill-installer` when Node.js is unavailable, `gh skill install` for exact release pinning, and manual copying only as the final fallback. No custom installer, PlanLens runtime, service, daemon, or plugin marketplace is justified for the current scope.
