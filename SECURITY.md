# Security Policy

## The short version

**Don’t be douchebags.**

That covers researchers, users, maintainers, and anyone filing reports:

- Report vulnerabilities **privately** when you can — don’t dump exploit details in public issues for clout.
- Give us a fair chance to fix things before full disclosure.
- Don’t steal data, harm users, or break systems “for research” when a responsible path exists.
- We won’t be jerks to people who report in good faith. Report in good faith.

## Supported versions

Security fixes are applied to the **latest `main`** branch (and releases cut from it, when tags exist).

Older commits / forks are best-effort only.

## How to report a vulnerability

**Preferred:** use GitHub’s private reporting for this repo:

→ [Report a vulnerability](https://github.com/coldcanuk/catfu/security/advisories/new)

If private reporting isn’t available for some reason, contact the maintainer privately (GitHub user [@coldcanuk](https://github.com/coldcanuk)) instead of opening a public issue with exploit details.

### What to include

- Description of the issue and impact
- Steps to reproduce (PoC if you have one)
- Affected version / commit if known
- Whether you have already disclosed it elsewhere

**Do not** put API keys, production credentials, or personal data in the report if you can avoid it. Redact when possible.

## What to expect

1. We’ll acknowledge the report when we can.
2. We’ll assess severity and work on a fix for in-scope issues.
3. We’ll coordinate disclosure timing with you when that helps everyone.
4. Credit is available if you want it (and you weren’t a douchebag about it).

We may decline reports that are:

- Social-engineering or physical attacks
- Denial of service against third-party services (YouTube, Brave, etc.)
- Issues only in **dependencies you run yourself** (e.g. outdated yt-dlp on your machine) with no catfu-side fix
- Pure theoretical risks with no practical impact
- Spam, drive-by noise, or automated scanner dumps with no validation

## Scope notes for catfu

catfu is a local CLI / MCP tool:

- It talks to **yt-dlp** (external) and optional **Brave Search** (BYOK API key).
- Secrets should never be logged; please report if they are.
- Path / config handling bugs that could leak or overwrite user data are in scope.
- “YouTube rate-limited me” is not a security vulnerability.

## Safe harbor (good faith)

If you research and report in good faith, without privacy harm, service destruction, or extortion, we will not pursue legal action for that research. **Don’t be douchebags** remains the rule either way.

## Thanks

Responsible reports make open source safer for everyone. We appreciate the help.
