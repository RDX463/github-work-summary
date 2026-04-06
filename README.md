<div align="center">
  <h1>github-work-summary</h1>
  <p><em>Professional work reporting for developers. AI-powered summaries, interactive dashboards, and team sharing.</em></p>
</div>

<p align="center">
  <a href="https://github.com/RDX463/github-work-summary/stargazers"><img src="https://img.shields.io/github/stars/RDX463/github-work-summary?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/RDX463/github-work-summary/releases"><img src="https://img.shields.io/github/v/tag/RDX463/github-work-summary?label=version&style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
  <a href="https://github.com/RDX463/github-work-summary/commits/main"><img src="https://img.shields.io/github/commit-activity/m/RDX463/github-work-summary?style=flat-square" alt="Commits"></a>
</p>

`github-work-summary` (or `gws`) is a CLI tool designed to help developers track and communicate their impact. It goes beyond simple commit listing by providing AI-powered narratives and interactive review workflows.

## ✨ Features

- **Professional AI Summaries**: Uses **Google Gemini** to transform technical commits into high-impact narrative summaries for your daily reports.
- **Interactive TUI Dashboard**: A full-screen review screen built with **Bubble Tea** and **Lip Gloss** for browsing, editing, and curating your summaries.
- **Team Integrations**: Share your reports directly to **Slack** or **Discord** using rich, platform-native formatting (Block Kit & Embeds).
- **Multi-Profile Support**: Manage sets of repositories and context (Work, Personal, OSS) with seamless profile switching.
- **Secure Keychain Storage**: Sensitive credentials (GitHub tokens, AI keys, Webhook URLs) are stored in your OS keychain—never in plaintext.
- **Auto-Update Engine**: Built-in semantic version checking to ensure you always have the latest features.

## 🚀 Quick Start

### Install via Homebrew (macOS/Linux)

```bash
brew tap RDX463/tap
brew install RDX463/tap/github-work-summary
```

### Install via script (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/RDX463/github-work-summary/main/install.sh | bash
```

### Fast Setup

1. **Login**: `gws login` (Uses GitHub Device Flow)
2. **AI Setup**: `gws ai-login` (Get a free key from [AI Studio](https://aistudio.google.com/))
3. **Run**: `gws summary -i --ai`

## 🛠 Commands

| Command | Description |
|---------|-------------|
| `gws` | Open the interactive startup dashboard |
| `gws summary` | Generate work summary. Use `-i` for TUI and `--ai` for AI insights. |
| `gws profiles` | Manage configuration profiles (Switch/Add/Remove) |
| `gws share setup` | Configure Slack or Discord webhooks securely |
| `gws ai-login` | Store your Google AI API key in the OS keychain |
| `gws update` | Manually check for project updates |
| `gws logout` | Securely wipe all credentials from your system |

## 🧩 Advanced usage

### Interactive TUI Controls
Inside the Dashboard (`gws summary -i`):
- `j` / `k`: Scroll the report
- `e`: **Edit Mode** - tweak the AI narrative in-place
- `s`: Share the current summary to **Slack**
- `d`: Share the current summary to **Discord**
- `Enter`: Finalize and exit

### Direct Sharing from CLI
Post your daily summary to a channel with one command:
```bash
gws summary --ai --share slack
```

## 🏗 Installation from Source

```bash
git clone https://github.com/RDX463/github-work-summary.git
cd github-work-summary
go build -o gws main.go
```

## 🛡 Security

`github-work-summary` prioritizes security by leveraging native OS credential managers:
- **macOS**: Keychain Access
- **Linux**: Secret Service (`gnome-keyring` / `ksecrets`)
- **Windows**: Credential Manager

No secrets are ever written to `~/.gws.yaml` or disk.

## 📄 License

MIT © [RDX463](https://github.com/RDX463)
