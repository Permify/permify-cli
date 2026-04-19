# AI and Cursor agent guide (this repository)

This document is for **human contributors** and for **coding agents** (Cursor Composer, Kimi, cloud agents, etc.) working on **Permify CLI**. It summarizes **vendor-neutral** practices and how to use **MCP** and **GitHub CLI** effectively without leaking credentials into issues or PRs.

Upstream reference: [Cursor — best practices for coding with agents](https://cursor.com/blog/agent-best-practices). Official rules docs: [Cursor Rules](https://cursor.com/docs/context/rules).

## Project facts (anchor for models)

- **Language**: Go (see `go.mod` for the supported toolchain).
- **CLI**: Cobra-based; entry in `cmd/permctl`.
- **Config**: YAML profiles; paths are user-controlled via flags (see `core/config` and `core/cli`).
- **Tests**: `go test ./...` (see `CONTRIBUTING.md`).

Keep instructions in **English** in commits, PR descriptions, and code comments unless maintainers require otherwise.

## Harness: instructions, tools, model

Agents behave as **instructions + tools + model**. Different models need different explicitness (e.g. when to run tests, when to stop and ask). You improve outcomes by:

1. **Planning first** for non-trivial work (Cursor **Plan Mode** / explicit plan in the thread). Plans can be saved under `.cursor/plans/` in a developer workspace for resumability.
2. **Letting the agent retrieve context** instead of pasting whole files: rely on search, `@path`, and repo structure; avoid dumping unrelated files into context.
3. **Starting a new chat** when the task changes or the model loops on the same mistake; use past-chat references instead of infinite scroll.
4. **Giving verifiable goals**: tests, `gofmt`, `go test`, smallest compile check.

## Rules and workspace layout

Cursor loads **Project Rules** from `.cursor/rules/` in the **opened workspace** (often the monorepo or parent folder that contains this clone). This upstream repo may be opened **alone** or **inside a larger workspace**; if rules live only in the parent repo, symlink or copy the relevant `.cursor/rules/` into the workspace root you open in Cursor.

Use **small, focused rules** (commands to run, style pointers, links to canonical files). Prefer **lint config** over pasting entire style guides.

## MCP (Model Context Protocol)

### Configuration

- MCP servers are declared in **Cursor user** or **project** `mcp.json`. **Never commit** API keys, bearer tokens, or `.mcp.json` secrets into this public repository.
- Cursor typically exposes each configured server to agents with an id of the form **`user-` + `<key>`**, where `<key>` is the key under `mcpServers` in `mcp.json`. Some hosts normalize slashes or punctuation in `<key>`; **confirm the exact id** in Cursor Settings → MCP, or from the local tool descriptor cache under the Cursor project directory.

### Before calling a tool

1. Open the server’s **tool descriptor** (JSON schema) and use the **exact** `toolName` and argument shape. Guessing names causes hard failures.
2. Prefer **one well-scoped call** over a chain of speculative calls.
3. On **HTTP MCP** timeouts or auth errors: check the latest `MCP user-*.log` under the Cursor logs directory, confirm network/VPN, then **switch strategy** (e.g. `gh`, `WebFetch`, local terminal) instead of retry-spamming the same call.

### Suggested routing (adjust to your own server list)

| Need | Typical approach |
|------|------------------|
| Up-to-date library docs / API surface | Context-style doc MCP when configured |
| Open web UI already logged in (e.g. dashboards, hosts with session) | Playwright-style browser MCP when configured |
| Long-form site crawl / search | Dedicated crawl MCP when configured; respect quotas |
| Session memory / code index | Memory or codebase MCP when configured |

**Playwright (or equivalent browser MCP)**: treat the remote browser as **already authenticated** where the operator set it up; use it for flows that `gh` cannot do (visual checks, authenticated web apps). Still avoid storing secrets in chat logs.

### CRAG / memory MCP

If a memory MCP is available, use it for **session handoff** on long threads (save state after meaningful steps, search at thread start). If it is unavailable, continue without blocking.

## GitHub and shipping

- Prefer **`gh`** for `issue view`, `pr create`, `pr diff`, CI log URLs, and comments (`/attempt`, `/claim` per program rules).
- After the maintainer or user approves shipping: **push branch, open PR, post required issue comments** via `gh`; do not stop at a paste-only template unless `gh` is broken (then report the error once with evidence).

## Security and review

- Do not echo **tokens** or session cookies in issues, PRs, or agent transcripts.
- AI-generated diffs still need **human review**; run **Agent Review** / CI and read the diff yourself.

## Maintainer note

This file is **documentation only**; it does not change CLI behavior. Update it when team-wide agent or MCP conventions change.
