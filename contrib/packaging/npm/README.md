# speckeep (npm launcher)

This package is a thin launcher, not a reimplementation: `speckeep` is a Go
binary, and this npm package exists so it can be run via `npx`/`npm install`
for people whose workflow already goes through the Node ecosystem.

On install, `postinstall` downloads the release archive matching this
package's version and your OS/arch from the
[speckeep GitHub Releases](https://github.com/bzdvdn/speckeep/releases),
verifies its sha256 against the release's `sha256sum.txt`, and extracts the
binary into `.bin/`. `bin/speckeep.js` then just execs that binary,
forwarding args, stdio, and exit code.

## Usage

```bash
npx speckeep init my-project --lang en --agents claude
# or
npm install -g speckeep
speckeep doctor .
```

## Env overrides

- `SPECKEEP_VERSION` — install a specific tag (e.g. `v1.0.0` or `1.0.0`)
  instead of the one matching this package's own version.
- `SPECKEEP_BINARY_PATH` — skip the download entirely and exec this path
  instead (useful for local development/testing against a locally built
  binary).
- `SPECKEEP_SKIP_DOWNLOAD` — skip the postinstall download (the launcher will
  then fail until a binary is provided via `SPECKEEP_BINARY_PATH` or a manual
  `node scripts/install.js` run).

## Supported platforms

linux/darwin/windows on amd64/arm64 — the same matrix speckeep publishes to
GitHub Releases. Anything else: install the Go binary directly, see the
[main README](https://github.com/bzdvdn/speckeep#install).

## Publishing (maintainers)

Automated: the `npm-publish` job in `.github/workflows/manual-release.yml`
runs after the `release` job creates the GitHub release for the tag. It sets
`package.json` `version` to the tag (without the `v` prefix) and runs
`npm publish --access public` — but only if the `NPM_TOKEN` repository secret
is set (an npm automation/publish token); otherwise it logs a skip notice and
does nothing. Add `NPM_TOKEN` under repo Settings → Secrets and variables →
Actions to enable it.

Manual fallback (or first publish, to reserve the name):

1. Bump `version` in `package.json` to match the git tag being released
   (without the `v` prefix), e.g. tag `v1.1.0` → package version `1.1.0`.
2. Make sure the GitHub release for that tag already exists with its assets +
   `sha256sum.txt` (installers need them at postinstall time).
3. `npm publish --access public` from this directory.
