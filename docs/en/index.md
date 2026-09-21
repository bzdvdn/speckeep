# speckeep Docs

`speckeep` is a lightweight project context kit for development agents and humans.

SpecKeep is the successor to DraftSpec (archived). If you are migrating an existing DraftSpec workspace, start with `speckeep migrate`.

This documentation is organized into a few practical guides:

- [Overview](overview.md)
- [CLI Reference](cli.md)
- [Workflow Model](workflow.md)
- [Architecture](architecture.md)
- [Agents](agents.md)
- [Language and Configuration](language-and-config.md)
- [Self-Hosting and Development](self-hosting.md)
- [Examples & Recipes](examples.md) — full lifecycle walkthrough plus a recipe for every secondary command (propose, converge, challenge, scope, glossary, hotfix, handoff, recap, rollback, repo-map, import, guard, self-update, express lane)
- [FAQ](faq.md)
- [Glossary](glossary.md)
- [Roadmap](roadmap.md)

Built with **Go 1.26+**.

## Quick Start

```bash
speckeep init my-project --lang en --agents claude --agents codex
```

This creates:

- `.speckeep/` workspace files
- project-local agent command files when `--agents` is used
- `AGENTS.md` guidance linked to SpecKeep structure, workflow, and templates

For a concise product summary, see the root [README](https://github.com/bzdvdn/speckeep/blob/master/README.md) and [MVP](https://github.com/bzdvdn/speckeep/blob/master/MVP.md).
