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

### 📊 Premium Analytics & Export
Generate professional reports in **JSON**, **Markdown**, or **Premium HTML**. Use the specialized HTML template for management reviews with responsive design and high-impact visuals.

### 🛡️ Proactive PR Intelligence
`gws pr create` now performs an AI-driven **Risk Assessment**. It calculates impact levels, identifies sensitive code areas (Security, API, DB), and automatically applies categorized GitHub labels.

### 🏢 Enterprise Ecosystem Sharing
Connect your terminal work to the platforms your team uses. Official support for **Notion**, **Microsoft Teams**, and **Email (SMTP)** with secure, keychain-backed credentials.

### ⚙️ Automation & Sovereignty
Total control over your reporting engine. Use **Custom AI Templates** (`~/.gws/templates/`) to override prompts, or use the **Webhook Listener** to trigger summaries from external CI/CD events.

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
| `gws share setup` | Configure Slack, Discord, Teams, Notion, or Email. |
| `gws webhook start`| Start an HTTP listener for event-driven summaries. |

---

## 🛡 Security

Your privacy is paramount. **GWS never stores secrets in plain text.**
- **GitHub Tokens & API Keys**: Stored securely in the native system keychain (macOS Keychain, Linux Secret Service).
- **Local AI**: Use **Ollama** to keep all your code analysis entirely on-device.

---

## 📄 License

MIT © [RDX463](https://github.com/RDX463)
