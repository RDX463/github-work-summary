# Security Policy

## Supported versions

Security fixes are applied to the latest release line.

| Version | Supported |
| --- | --- |
| Latest (`v0.1.x`) | Yes |
| Older releases | No |

## Reporting a vulnerability

Please do not open public GitHub issues for security vulnerabilities.

Use GitHub private vulnerability reporting:

1. Go to the repository **Security** tab.
2. Click **Report a vulnerability**.
3. Provide details:
   - affected version
   - environment (OS, shell, install method)
   - reproduction steps
   - expected vs actual behavior
   - impact assessment

If private reporting is unavailable, open an issue with minimal details and request a secure contact channel.

## Response targets

- Initial triage acknowledgement: within 72 hours
- Status update after triage: within 7 days
- Fix timeline depends on severity and exploitability

## Disclosure policy

- Please allow time for triage and a patch before public disclosure
- Once fixed, release notes will include a security acknowledgement when appropriate

## Security design notes

- OAuth tokens are stored in OS-native credential stores:
  - macOS Keychain
  - Linux Secret Service
  - Windows Credential Manager
- The project avoids writing tokens to plain text config files
- Network operations are limited to expected GitHub API and release/update flows

## User hardening recommendations

- Use least-privilege scopes for GitHub OAuth apps
- Revoke and reissue tokens if compromise is suspected
- Keep the CLI updated to the latest release
- Review scripts before piping from the internet in sensitive environments
