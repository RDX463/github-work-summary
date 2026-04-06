<div align="center">
  <h1>github-work-summary (gws)</h1>
  <p><em>The AI-powered work intelligence platform for high-velocity developers.</em></p>
</div>

<p align="center">
  <a href="https://github.com/RDX463/github-work-summary/stargazers"><img src="https://img.shields.io/github/stars/RDX463/github-work-summary?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/RDX463/github-work-summary/releases"><img src="https://img.shields.io/github/v/tag/RDX463/github-work-summary?label=version&style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
</p>

`gws` is a sophisticated CLI tool designed to bridge the gap between technical code changes and business impact. It leverages cutting-edge AI to automate your daily standups, pull request descriptions, and project management updates.

## 🚀 Key Features

### 🤖 Multi-LLM Intelligence
Choose the AI that fits your workflow. Support for **Google Gemini Pro**, **Anthropic Claude 3**, and **Local Ollama** (Llama 3/Mistral) for privacy-conscious environments.

### 🔀 AI-Powered PR Automation
Generate professional Pull Requests with a single command. `gws pr create` analyzes your local changes, identifies the business impact, and opens a perfectly formatted PR on GitHub.

### 🎫 Contextual Insights (Jira & Linear)
No more manual status updates. `gws` automatically extracts ticket IDs from your commits and fetches issue titles and statuses from **Jira** or **Linear** to enrich your reports.

### 🕰️ Zero-Touch Scheduling
Wake up to a completed work summary. Use `gws schedule` to register native macOS background jobs that automatically post your AI-summarized work to **Slack** or **Discord**.

### 📊 Interactive TUI Dashboard
A premium terminal interface built with **Bubble Tea**. Browse, edit, and curate your work history with a keyboard-driven workflow.

---

## 🛠 Installation

### via Homebrew (Recommended)
```bash
brew tap RDX463/tap
brew install RDX463/tap/github-work-summary
```

### via Binary
Download the latest release from the [Releases](https://github.com/RDX463/github-work-summary/releases) page and move it to your `/usr/local/bin`.

---

## 📖 Quick Start

1.  **GitHub Login**: `gws login`
2.  **AI Provider**: `gws ai-login --provider gemini`
3.  **Project Context (Optional)**: `gws tickets-login jira`
4.  **Run Summary**: `gws summary -i --ai`

---

## 📟 Command Reference

| Command | Description |
|---------|-------------|
| `gws summary` | Generate work reports. Use `-i` for TUI, `--ai` for insights. |
| `gws pr create` | Create an AI-powered Pull Request from your current branch. |
| `gws schedule` | Configure automated background reports (macOS). |
| `gws watch` | Monitor your schedule in a persistent foreground process. |
| `gws tickets-login`| Securely store Jira or Linear API credentials. |
| `gws ai-login` | Configure Gemini or Anthropic API keys. |
| `gws profiles` | Manage different work/client/OSS configurations. |
| `gws share setup` | Configure Slack/Discord webhooks. |

---

## 🛡 Security

Your privacy is paramount. **GWS never stores secrets in plain text.**
- **GitHub Tokens & API Keys**: Stored securely in the native system keychain (macOS Keychain, Linux Secret Service).
- **Local AI**: Use **Ollama** to keep all your code analysis entirely on-device.

---

## 📄 License

MIT © [RDX463](https://github.com/RDX463)
