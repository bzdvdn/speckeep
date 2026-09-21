# Packaging

Distribution channels for `speckeep`. These files are the **seed copies** /
reference for each channel; the live packages live outside this repo.

| Channel | Live location | Updated by |
| --- | --- | --- |
| npm | `npmjs.com/package/speckeep` | `npm-publish` job in `.github/workflows/manual-release.yml` (needs `NPM_TOKEN`) |
| Homebrew | `github.com/bzdvdn/homebrew-speckeep` → `Formula/speckeep.rb` | `publish-packages` job (needs `PKG_PUSH_TOKEN`) |
| Scoop | `github.com/bzdvdn/scoop-speckeep` → `bucket/speckeep.json` | `publish-packages` job (needs `PKG_PUSH_TOKEN`) |

On every release the `publish-packages` job computes the source-tarball
`sha256` (from the tag archive) and the Windows asset `sha256`s (from the
release `sha256sum.txt`), rewrites the formula / manifest with
`update-release.py`, and pushes them to the two package repos. The copies in
`brew/` and `scoop/` here are not auto-updated; they seeded the repos at
`v1.0.0`.

## Manual update

```bash
python3 contrib/packaging/update-release.py formula <path/to/speckeep.rb> <version-without-v> <source-sha256>
python3 contrib/packaging/update-release.py scoop   <path/to/speckeep.json> <version-without-v> <amd64-sha256> <arm64-sha256>
```

See the script's usage header for details.
