# Contributing to github-work-summary

Thanks for your interest in contributing.

## Ways to contribute

- Report bugs
- Suggest features
- Improve documentation
- Submit code fixes or enhancements

## Before you start

- Search existing issues and pull requests first
- Keep changes focused and small when possible
- Open an issue first for large feature work

## Development setup

1. Fork the repository and clone your fork.
2. Create a feature branch from `main`.
3. Install Go `1.21` or newer.
4. Install dependencies and verify build.

```bash
go mod tidy
go build ./...
go test ./...
```

Run locally:

```bash
go run . --help
go run . login
go run . summary
```

## Code guidelines

- Follow idiomatic Go and keep code modular
- Use clear names and straightforward control flow
- Keep terminal output readable in both interactive and non-interactive modes
- Preserve cross-platform behavior (macOS, Linux, Windows)
- Do not commit secrets, tokens, or local credential files

## Commit and PR guidelines

- Use descriptive commit messages
- Include context in PR description:
  - what changed
  - why it changed
  - how it was tested
- Add screenshots or terminal output for UI/CLI behavior changes
- Ensure `go build ./...` and `go test ./...` pass before requesting review

## Testing checklist

- Command help output still works (`--help`, `--version`)
- Login flow behavior remains clear and safe
- Repo and branch selection flows are usable
- Summary output renders correctly with and without color
- No regressions in no-commit and fallback paths

## Release notes for contributors

Releases are currently published by maintainers and include:

- GitHub release assets for macOS/Linux/Windows
- `checksums.txt`
- Homebrew tap formula update

If your PR affects install, versioning, or release assets, note it clearly in the PR.

## Questions

If anything is unclear, open an issue with a short reproduction or proposal.
