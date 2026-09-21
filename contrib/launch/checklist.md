# speckeep launch checklist

Working doc for shipping **v1.0.1** and introducing speckeep to the community.
Tick items as they land; keep this file in `dev` until the launch is done.

## 1. Ship

- [x] `sh scripts/release-check.sh` green on `dev` (25/25)
- [ ] Merge `dev` → `master`
- [ ] Tag `v1.0.1` and run `.github/workflows/manual-release.yml`
      (it auto-generates release notes; optionally replace the release body with
      the curated `contrib/launch/v1.0.1.md` afterward)
- [ ] `docs` workflow deployed → <https://bzdvdn.github.io/speckeep/> returns 200
- [ ] `speckeep self check` reports the new version
- [ ] Set the GitHub **social preview** image (Settings → General → Social preview;
      use a frame from `demo/speckeep-demo.gif` — the API cannot set this)

## 2. Publish channels

- [x] **npm**: `NPM_TOKEN` secret present; the `npm-publish` job in
      `manual-release.yml` publishes `contrib/packaging/npm`. Verify after release
      with `npm view speckeep version`.
- [x] **Homebrew tap**: <https://github.com/bzdvdn/homebrew-speckeep> created
      with `Formula/speckeep.rb` (seeded at v1.0.0; auto-bumped on each release).
- [x] **Scoop bucket**: <https://github.com/bzdvdn/scoop-speckeep> created
      with `bucket/speckeep.json` (seeded at v1.0.0; auto-bumped on each release).
- [x] **`PKG_PUSH_TOKEN` secret** added: fine-grained PAT with **Contents: Read
      and write** on both package repos. The `publish-packages` job in
      `manual-release.yml` bumps the formula and manifest automatically on every
      release (source-tarball sha256 from the tag archive, Windows asset sha256
      from the release `sha256sum.txt`); it skips with a notice when absent.
      Manual fallback: `python3 contrib/packaging/update-release.py formula|scoop …`.
- [ ] **GitHub Action Marketplace**: from the `v1.0.1` release page, "Publish this
      Action" → pick `.github/actions/speckeep` (it already declares `branding`).

## 3. Announce

Order: docs live → release notes posted → social. One channel per day beats a
single blast, so replies can be answered.

- [ ] Hacker News (`Show HN`)
- [ ] Reddit: r/ClaudeAI, r/ChatGPTCoding, r/LocalLLaMA, r/ExperiencedDevs
- [ ] X/Twitter thread
- [ ] dev.to / Medium article (expand the "Why" section below)
- [ ] LinkedIn post
- [ ] Agent Discords/Slacks (Claude, opencode, Continue, Cursor communities) — link, don't spam

### Show HN draft

> **Show HN: speckeep – spec-driven development your coding agent can actually follow**
>
> Agent coding sessions lose context between runs. speckeep keeps specs, plans,
> tasks and proof-based traceability in plain files, turns them into
> `/spk-<phase>` skills for 19 agents (Claude Code, Codex, Cursor, Copilot,
> OpenCode, …), and adds a CLI gate (`speckeep check` / `converge` / `guard`)
> so "done" is machine-verifiable. Single Go binary, MIT.
> Demo: https://bzdvdn.github.io/speckeep/ · Repo: https://github.com/bzdvdn/speckeep
>
> Happy to answer anything about the workflow design or the per-agent quirks
> (skills vs slash commands differ a lot between tools).

### X/Twitter thread draft

1. Shipping speckeep v1.0.1 — strict, lightweight spec-driven development for coding agents.
2. The problem: agents lose context, drift from requirements, and mark things done without proof.
3. The fix: a tiny file-based chain — constitution → spec → plan → tasks → implement → archive — with stable `AC-*`/`RQ-*` IDs and `Proof:` lines.
4. One binary. 19 agent adapters. Skills-first, so it works the way each tool actually invokes things.
5. CI gate: `speckeep guard` fails the PR when a touched feature isn't closeable.
6. Docs (EN/RU) + 30-second demo: https://bzdvdn.github.io/speckeep/ — MIT, feedback welcome.

### "Why" article outline

- Pain: why chat-history-driven agent work doesn't scale
- Design: discipline per token, narrow context, proof-based done
- Demo: `speckeep demo` → lifecycle walkthrough
- Integrations: 19 adapters + the skills-vs-commands reality per tool
- CI: `guard` in a PR
- Migration: OpenSpec / Spec Kit import
- What's next / roadmap

## 4. Awesome lists & directories

- [ ] `awesome-claude-code`
- [ ] `awesome-ai-coding` / `awesome-ai-agents`
- [ ] `awesome-spec-driven-development` (if present)
- [ ] `awesome-go` → "Project Management" / "Other Software"
- [ ] `awesome-context-engineering`
- [ ] PR text: one-line description, repo link, what makes it distinct
      (strict phase chain + narrow context + proof-based traceability + brownfield)

## 5. Post-launch

- [ ] Pin a `good first issue` and label a few `help wanted`
- [ ] Enable a Discussions "Show your setup" / "Q&A" welcome post
- [ ] Add `CODEOWNERS` for `src/internal/templates/assets/**`
- [ ] Publish a public versioning/deprecation policy (agent-target tiers already exist)
- [ ] Track installs (`self check` telemetry is opt-in/none — use GitHub release download counts)
