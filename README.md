# github-work-summary

A terminal CLI that authenticates with GitHub, lets you select repositories, fetches your commits from the last 24 hours, and prints a categorized summary.

## Features

- OAuth 2.0 Device Flow login (`login`)
- Logout support (`logout`) to clear stored credentials
- Secure token storage in OS keychain/credential manager
- Interactive multi-repo checkbox selection
- Commit retrieval from the last 24 hours
- Categorization into:
  - Features/Implementations
  - Bug Fixes
  - Other
- Colorized terminal output

## Requirements

- Go 1.21+ (if building from source)
- A GitHub OAuth App with **Device Flow enabled**

## Installation

### Linux (one command)

Install latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | GWS_VERSION=v0.1.0 bash
```

The installer adds:

- `github-work-summary`
- `gws` (shortcut)

### macOS and Linux (Homebrew)

If you published a tap:

```bash
brew tap RDX463/tap
brew install RDX463/tap/github-work-summary
```

Then run:

```bash
github-work-summary --help
```

Optional short alias:

```bash
echo 'alias gws="github-work-summary"' >> ~/.zshrc
source ~/.zshrc
gws --help
```

### Windows / macOS / Linux (from GitHub Releases)

1. Download the correct binary asset from Releases.
2. Extract it.
3. Put the binary in your `PATH`.

Examples:

- Windows: place `github-work-summary.exe` in a folder on `PATH`.
- macOS/Linux: place `github-work-summary` in `/usr/local/bin` or `~/.local/bin`.

### Install with Go

```bash
go install github.com/RDX463/github-work-summary@latest
```

### Build from source (all platforms)

```bash
git clone https://github.com/RDX463/github-work-summary.git
cd github-work-summary
go build -o github-work-summary .
```

For Windows, output with `.exe`:

```bash
go build -o github-work-summary.exe .
```

## GitHub OAuth App Setup

Create or configure an OAuth App in GitHub:

1. Open: `https://github.com/settings/developers`
2. Go to **OAuth Apps** -> your app
3. Enable **Device Flow**
4. Use your app credentials with the CLI

You can provide credentials through flags or env vars:

- `--client-id` or `GITHUB_CLIENT_ID`
- `--client-secret` or `GITHUB_CLIENT_SECRET` (optional)

## Quick Start

1. Authenticate:

```bash
github-work-summary login --client-id <YOUR_CLIENT_ID> --client-secret <YOUR_CLIENT_SECRET>
```

2. Generate summary:

```bash
github-work-summary summary
```

3. Optional: only list/select repositories:

```bash
github-work-summary repos
```

## Commands

### `login`

Authenticates with GitHub using OAuth Device Flow and stores the token in your OS keychain.

```bash
github-work-summary login [--client-id ...] [--client-secret ...]
```

### `repos`

Fetches all accessible repositories and opens an interactive multi-select list.

```bash
github-work-summary repos
```

### `logout`

Removes the stored GitHub token from your OS keychain.

```bash
github-work-summary logout
```

### `summary`

Lets you select repositories, fetches your commits from the last 24 hours, categorizes them, and prints a colorized report.

```bash
github-work-summary summary
```

## Interactive Selection Controls

Inside the repo selector:

- `1 3 5` toggle specific rows
- `2-6` toggle a range
- `a` select all
- `n` clear all
- `d` done
- `q` cancel

## Output Categories

Commit messages are grouped into:

- **Features/Implementations**
- **Bug Fixes**
- **Other**

The tool uses lightweight keyword matching on commit subjects.

## Security and Token Storage

Access tokens are stored using native OS credential storage via keyring:

- macOS: Keychain
- Linux: Secret Service (for example `gnome-keyring`)
- Windows: Credential Manager

## Troubleshooting

### `device_flow_disabled`

Your GitHub OAuth App has Device Flow disabled.

Fix:

1. Open GitHub Developer Settings
2. OAuth Apps -> your app
3. Enable Device Flow

### `stored token is invalid or expired`

Re-authenticate:

```bash
github-work-summary login
```

### `secret not found in keyring`

No saved token found. Run:

```bash
github-work-summary login
```

### No commits shown

If no commits exist in the last 24 hours, the summary prints an explicit no-commits message.

## Development

```bash
go fmt ./...
go build ./...
go test ./...
```
