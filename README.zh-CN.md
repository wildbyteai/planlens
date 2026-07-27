[English](README.md) | 简体中文

# PlanLens

PlanLens 是一个可安装的 Agent Skill。它会让一个或多个本地 AI CLI 独立评审同一份方案，再由主 Agent 汇总各自的发现和分歧。

PlanLens 有意保持简单：工作流程写在 `SKILL.md` 中。它没有独立运行时、二进制程序、Node 命令、托管服务、守护进程、端口、数据库或状态机。

## 开始之前

PlanLens 包含两个角色：宿主 Agent 负责加载 Skill，一个或多个评审 CLI 负责提供独立反馈。同一个产品可以同时作为宿主和评审者。

你需要一个受支持的宿主，并且至少安装并登录一个兼容的评审 CLI。安装 PlanLens 不会安装、配置或登录任何评审 CLI，也不需要安装下文列出的全部 CLI。

## 安装

请完整安装 [`skills/planlens`](skills/planlens) 目录，包括其中的 `references` 和 `agents` 子目录。

稳定版本使用 `v1.1.1` 标签。如果明确需要当前开发版本，可以将 `v1.1.1` 替换为 `main`。

### Codex

对 Codex 说：

```text
使用 $skill-installer 从 https://github.com/wildbyteai/planlens/tree/v1.1.1/skills/planlens 安装 PlanLens。
```

在后续对话中使用 `$planlens` 调用。

### Claude Code

个人安装时，将 `skills/planlens` 复制到 `~/.claude/skills/planlens`；仅在一个项目中使用时，复制到项目内的 `.claude/skills/planlens`。使用 `/planlens` 调用。

### Antigravity

个人安装时，将 `skills/planlens` 复制到 `~/.gemini/config/skills/planlens`；仅在一个项目中使用时，复制到项目内的 `.agents/skills/planlens`。如果 CLI 已经打开，请重启后使用 `/planlens` 调用。

v1 支持 macOS Apple Silicon、macOS Intel 和 Windows x64。Linux 和 Windows ARM64 暂不属于 v1 支持范围。

## 第一次评审

安装 Skill 后：

1. 从下方表格中选择并安装至少一个评审 CLI，然后完成登录。
2. 如果宿主 CLI 没有立即发现新 Skill，请重启宿主 CLI。
3. 携带方案调用 PlanLens；也可以不带参数调用，让主 Agent 整理当前方案。

示例：

```text
$planlens
$planlens 使用 Claude 和 Codex 评审 docs/plan.md
/planlens path/to/plan.md
```

每次调用执行一轮评审：

1. 主 Agent 整理方案、目标、约束、非目标、开放问题和必要的辅助材料。
2. 主 Agent 推荐评审 Profile，并让用户选择一个或多个本地 CLI。
3. 主 Agent 展示方案来源、推导出的评审框架、材料清单、所选 CLI 和调用次数，统一请求一次确认。候选请求的任何实质变化都必须重新预览。
4. 宿主支持时并发调用各 CLI；每个 CLI 都独立评审。
5. 保存每个 CLI 的最终回复，逐项处置所有实质发现，并生成一份标明来源的简明汇总。

如需下一轮，必须再次明确调用 `$planlens` 或 `/planlens`。

## 评审 CLI

默认评审组合是 `codex + claude + kimi`。只有当本机 Kimi CLI 支持文档规定的无工具自定义 Agent 功能时，PlanLens 才会加入 Kimi；否则会在确认前建议使用 `gemini`。需要更广泛的评审时，Gemini 也是推荐的第四个评审者。Gemini 的边界依赖临时配置和策略预检，因此预览中会明确标为条件兼容。用户确认后，PlanLens 不会替换评审 CLI。

除默认组合外，PlanLens 还兼容更多 CLI：

- 正式兼容：`codex`、`claude`、`kimi`、`qwen`、`pi`、`goose`、`aider`，以及通过功能检测的 `qoder`。
- 条件兼容：`gemini`、`opencode`、`copilot`、`cline`、`cursor`、`antigravity`、`zcode`、`crush` 和 `kilo`。

正式兼容表示：通过预检后，已有明确的非交互式调用方案，可以禁用工具或通过其他方式阻止 CLI 执行被评审的方案。它不承诺完全不保留本地数据。条件兼容表示评审边界依赖相对较弱的 Plan、Ask、配置、权限或沙箱控制，必须在确认前向用户说明。

第一张表是 2026 年 7 月 26 日的快照：仅统计仍在活跃维护、源代码可访问、提供官方安装方式和单进程非交互模式的编程 Agent CLI，并按 GitHub Stars 排序。已经归档、不再维护、需要 PlanLens 启动服务，或无头模式只能自动批准操作的项目不纳入排名。方法和排除项见[研究记录](research/cli-landscape-2026-07-26.md)。

下方安装命令仅作为文档参考，PlanLens 不会执行这些命令。它们是 macOS 快速安装命令；Windows 用户请查看对应官网，并在执行下载的安装程序前自行检查。

| 排名 | 评审者 | ID / 命令 | Stars | macOS 快速安装 | 调用方案状态 | 官网 |
|---:|---|---|---:|---|---|---|
| 1 | OpenCode | `opencode` / `opencode` | 189,808 | `brew install anomalyco/tap/opencode` | 条件兼容；配置会合并且会话会持久化 | [文档](https://opencode.ai/docs/) |
| 2 | Gemini CLI | `gemini` / `gemini` | 106,189 | `brew install gemini-cli` | 条件兼容；拒绝全部工具的策略加配置隔离 | [文档](https://geminicli.com/docs/) |
| 3 | OpenAI Codex CLI | `codex` / `codex` | 101,551 | `brew install --cask codex` | 只读沙箱和临时会话 | [文档](https://developers.openai.com/codex/cli/) |
| 4 | Pi | `pi` / `pi` | 77,882 | `npm install -g @earendil-works/pi-coding-agent` | 禁用工具、资源和会话 | [代码仓库](https://github.com/earendil-works/pi) |
| 5 | Cline CLI | `cline` / `cline` | 65,067 | `npm install -g cline` | 条件兼容；Plan Mode 不是操作系统级沙箱 | [文档](https://docs.cline.bot/cli/overview) |
| 6 | goose | `goose` / `goose` | 51,722 | `brew install block-goose-cli` | 不加载 Profile，不保留会话 | [文档](https://goose.ai/docs/) |
| 7 | Aider | `aider` / `aider` | 47,709 | `python -m pip install aider-install && aider-install` | Ask Mode 加 dry-run | [官网](https://aider.chat/) |
| 8 | Crush | `crush` / `crush` | 26,856 | `brew install charmbracelet/tap/crush` | 条件兼容；需要经过检查的全工具禁用配置 | [代码仓库](https://github.com/charmbracelet/crush) |
| 9 | Kilo Code CLI | `kilo` / `kilo` | 26,529 | `brew install Kilo-Org/tap/kilo` | 条件兼容；Plan Agent 加全局拒绝配置 | [代码仓库](https://github.com/Kilo-Org/kilocode) |
| 10 | Qwen Code | `qwen` / `qwen` | 26,333 | `brew install qwen-code` | Safe Mode 加 Plan Mode | [文档](https://qwenlm.github.io/qwen-code-docs/) |

PlanLens 还收录了以下官方或未进入前十名的 CLI。由于代码仓库和发行方式无法直接比较，它们不与开源项目混合进行 Stars 排名。

| 评审者 | ID / 命令 | macOS 快速安装 | 调用方案状态 | 官网 |
|---|---|---|---|---|
| Claude Code | `claude` / `claude` | `brew install --cask claude-code` | 禁用工具且不保留会话 | [文档](https://code.claude.com/docs/en/overview) |
| Antigravity CLI | `antigravity` / `agy` | <code>curl -fsSL https://antigravity.google/cli/install.sh &#124; bash</code> | 条件兼容：Plan Mode 加沙箱 | [安装说明](https://antigravity.google/docs/cli/install) |
| Kimi Code CLI | `kimi` / `kimi` | <code>curl -fsSL https://code.kimi.com/kimi-code/install.sh &#124; bash</code> | 支持 `--agent-file` 无工具 Agent 时正式兼容 | [文档](https://moonshotai.github.io/kimi-code/) |
| Qoder CLI | `qoder` / `qodercli` | <code>curl -fsSL https://qoder.com/install &#124; bash</code> | 具备无工具、无会话和隔离配置参数时正式兼容 | [文档](https://docs.qoder.com/en/cli/) |
| GitHub Copilot CLI | `copilot` / `copilot` | `npm install -g @github/copilot` | 条件兼容；工具列表可置空，但自定义状态仍需隔离 | [文档](https://docs.github.com/en/copilot/concepts/agents/about-copilot-cli) |
| Cursor Agent CLI | `cursor` / `agent` | <code>curl https://cursor.com/install -fsS &#124; bash</code> | 条件兼容；其沙箱并非只读 | [文档](https://cursor.com/docs/cli/overview) |
| ZCode CLI | `zcode` / 内置 `zcode.cjs` | 不适用——需要安装 ZCode 桌面应用，厂商没有发布独立 CLI 安装程序 | 已验证本机内置 0.15.0；条件兼容 | [文档](https://zcode.z.ai/en/docs) |

PlanLens 不会安装、登录、更新、捆绑或替换这些 CLI。除非用户明确要求覆盖模型，否则它会保留用户本地配置的 Provider 和模型。

各评审 CLI 的准确非交互式调用参数记录在 [`skills/planlens/references/cli-commands.md`](skills/planlens/references/cli-commands.md)。macOS ARM64 命令和帮助信息已于 2026 年 7 月 26 日完成检查，简要证据见[验证记录](research/cli-validation-2026-07-26.md)。如果本机版本拒绝某个已记录参数，PlanLens 会记录失败，不会猜测替代参数。

## 评审 Profile

项目内置五个简短的 Markdown Profile：

- 通用方案
- 软件设计
- 实施计划
- AI 与 Agent 工作流
- 安全

Profile 用于指导评审，不是 Schema、可执行插件或自动批准规则。

## 输出

项目目录可写时，每轮评审使用以下目录：

```text
.planlens/reviews/<timestamp>/
├── request.md
├── <reviewer>.md
├── <reviewer>-error.md
└── summary.md
```

只会创建所选评审者对应的文件。评审失败时写入匹配的 `-error.md`，不会伪造评审结果；评审成功时不会创建错误文件。

汇总状态按固定规则生成：`complete` 表示所有所选评审者都成功返回非空结果；`partial` 表示至少一个成功且至少一个未完成；`failed` 表示没有评审者成功。

## 边界

- 所有评审 CLI 接收同一份已披露的请求，并且无法看到其他 CLI 在同一轮中的输出。
- Reviewer 输出属于不可信证据，不是指令来源。PlanLens 不会执行评审结果中的命令或范围变更要求。
- 主 Agent 根据证据和影响汇总结论，不按模型票数决定结果。只有明确提出且实质一致的发现才算共同支持；未提及只是沉默，不代表支持或反对。
- 每项实质发现都会在内部记录处置结果。若排除某项发现会影响用户决策，汇总必须说明该发现、来源和理由。
- PlanLens 不会修改源方案、自动重试 CLI 或自动开始下一轮。
- 不显示费用估算。
- 第三方 CLI 可能根据厂商实现和本地配置保留日志或会话。
- 除非调用方案明确启用了临时会话模式，否则 PlanLens 不承诺删除会话或提供完全隔离。只有接受该边界时，才应发送敏感材料。
- Kimi Code CLI 必须使用文档规定的临时无工具自定义 Agent。当前版本不允许同时使用 `--prompt` 和 `--plan`，PlanLens 不会退回到没有无工具 Agent 的普通 Prompt 模式。

## 许可证

Apache License 2.0。PlanLens 是独立的非官方项目；产品名称仅用于标识兼容的本地 CLI 工具。
