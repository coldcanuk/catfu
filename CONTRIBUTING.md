# Contributing to catfu

Thanks for helping. Small, focused changes are welcome.

## Before you start

Use the [issue templates](https://github.com/coldcanuk/catfu/issues/new/choose) when reporting bugs or requesting features.

1. Read the [README](README.md) and [Code of Conduct](CODE_OF_CONDUCT.md).
2. Search existing issues and PRs so you don’t duplicate work.
3. For larger changes, open an issue first and sketch the approach.

## Development setup

```bash
# Requires Go 1.25+ and yt-dlp on PATH
git clone https://github.com/coldcanuk/catfu.git
cd catfu
make build
make test
make vet
./bin/catfu doctor
```

## Branch & PR workflow

`main` is **protected**. You cannot push to it directly.

1. Create a branch from up-to-date `main`:
   ```bash
   git checkout main && git pull
   git checkout -b feature/short-description
   ```
2. Make focused commits (conventional commits preferred, e.g. `fix:`, `feat:`, `docs:`).
3. Push your branch and open a **pull request** into `main`.
4. Keep the PR small when you can. Explain *why*, not only *what*.
5. Ensure `go test ./...` and `go vet ./...` pass before requesting merge.

## What makes a good contribution

- Fixes a real bug or adds a clearly useful feature.
- Matches existing style and package layout.
- Includes tests for non-trivial logic.
- Does not commit secrets, local `*.db` files, or vendored yt-dlp.
- Updates docs when user-facing behavior changes.
- Leaves logging free of secrets (API keys, tokens).

## What we will reject or close

- Drive-by noise: empty PRs, “fix typo” spam, or churn with no benefit.
- Mass automated PRs that only bump deps or reformat without discussion.
- Changes that break the agent-friendly CLI contract without a strong reason
  (stable `--json` output, meaningful exit codes, stderr for progress/logs).
- Bundling or vendoring yt-dlp; it stays an external runtime dependency.

## Security

Do not open public issues for sensitive security reports if that would increase risk.
Prefer contacting the maintainers privately when possible.

## License

By contributing, you agree your contributions are licensed under the project’s
[GNU GPLv3](LICENSE) license.

---

## Guidelines for AI assistants and automated agents

catfu is often improved with help from coding agents. That is welcome **when the
agent behaves like a careful collaborator**, not a spam bot.

### Do

- Work from a real need stated by a human (issue, PR description, or clear request).
- Prefer **one focused PR** over many tiny drive-by PRs.
- Run tests and fix failures before opening or updating a PR.
- Summarize what changed and how it was verified.
- Respect branch protection: open a PR; do not fight `main`.
- Keep diffs minimal. No drive-by refactors, renames, or “cleanup” unless asked.
- Never invent API keys, commit secrets, or log credentials.
- Quote uncertainty honestly. Do not claim tests passed if they did not run.

### Do not

- Open bulk PRs, dependency-noise PRs, or cosmetic-only churn without a human ask.
- Spam issues/PRs with repetitive comments, status fluff, or self-promotion.
- Rewrite large parts of the codebase “for style” or “best practices” unprompted.
- Add frameworks, agents, or telemetry that the project did not request.
- Force-push over others’ work, or reopen closed spam.
- Generate fake reviews, fake benchmarks, or padded commit history.
- Ignore the Code of Conduct. Automated does not mean excused.

### Rule of thumb

If a careful human maintainer would call it noisy, low-value, or inconsiderate,
**don’t open it**. Be cool, be kind, be helpful — and ship something that earns
its place in the tree.
