<div align="center">
  <h1>github-work-summary</h1>
  <p><em>Generate a clean 24-hour GitHub work summary from your terminal.</em></p>
</div>

<p align="center">
  <a href="https://github.com/RDX463/github-work-summary/stargazers"><img src="https://img.shields.io/github/stars/RDX463/github-work-summary?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/RDX463/github-work-summary/releases"><img src="https://img.shields.io/github/v/tag/RDX463/github-work-summary?label=version&style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
  <a href="https://github.com/RDX463/github-work-summary/commits/main"><img src="https://img.shields.io/github/commit-activity/m/RDX463/github-work-summary?style=flat-square" alt="Commits"></a>
  <a href="https://github.com/RDX463/github-work-summary/issues"><img src="https://img.shields.io/github/issues/RDX463/github-work-summary?style=flat-square" alt="Issues"></a>
</p>

## Features

- OAuth 2.0 Device Flow login with secure OS keychain storage
- Interactive startup dashboard menu with keyboard navigation
- Interactive multi-select repository picker in terminal
- Commit fetch for selected repos (last 24 hours)
- Categorized output: Features/Implementations, Bug Fixes, Other
- Colorized summary output with per-repo grouping
- `logout` command to remove saved token safely
- Built-in update notifier with changelog highlights when a newer release exists

## Quick Start

### Install via script (Linux, one command)

```bash
curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | bash
```

### Install via Homebrew (macOS/Linux)

```bash
brew tap RDX463/tap
brew install RDX463/tap/github-work-summary
```

### Run

```bash
gws                              # Interactive dashboard menu
gws login
gws summary
gws logout
```

## Installation

### Linux script options

```bash
# Latest release
curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | bash

# Specific version
curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | GWS_VERSION=v0.1.1 bash
```

Installer outputs:

- `github-work-summary`
- `gws` (shortcut)

### From GitHub Releases (Windows/macOS/Linux)

1. Download matching archive from [Releases](https://github.com/RDX463/github-work-summary/releases).
2. Extract binary.
3. Move binary into a folder on `PATH`.

### Go install

```bash
go install github.com/RDX463/github-work-summary@latest
```

### Build from source

```bash
git clone https://github.com/RDX463/github-work-summary.git
cd github-work-summary
go build -o github-work-summary .
```

## Commands

```bash
gws                              # Interactive startup menu (arrow keys + enter)
gws login                        # Authenticate with GitHub
gws repos                        # Pick repositories interactively
gws summary                      # Generate work summary (last 24h)
gws logout                       # Remove stored token from keychain
gws --help                       # Show all commands
gws --version                    # Show installed version
```

### Interactive selection controls

Inside `repos` and `summary` repo picker:

- `1 3 5` toggle specific items
- `2-6` toggle a range
- `a` select all
- `n` clear all
- `d` done
- `q` cancel

## Auto Update Notice

If your installed version is older than the latest GitHub release, the CLI shows:

- current version -> latest version
- changelog highlights (top bullet points from latest release notes)
- direct update command

Disable update checks (for CI/non-network environments):

```bash
export GWS_NO_UPDATE_CHECK=1
```

## Authentication Setup (GitHub OAuth App)

1. Open `https://github.com/settings/developers`
2. Go to **OAuth Apps** -> your app
3. Enable **Device Flow**

Optional credential env vars:

- `GITHUB_CLIENT_ID`
- `GITHUB_CLIENT_SECRET`

## Security

Tokens are stored using native OS credential stores:

- macOS: Keychain
- Linux: Secret Service (`gnome-keyring` compatible)
- Windows: Credential Manager

## Troubleshooting

### `device_flow_disabled`

Enable Device Flow in your GitHub OAuth App settings.

### `stored token is invalid or expired`

```bash
gws login
```

### `secret not found in keyring`

```bash
gws login
```

### No commits in output

If there are no commits in the last 24 hours, summary prints a clean no-activity message.

## Development

```bash
go fmt ./...
go build ./...
go test ./...
```

## License

MIT.
