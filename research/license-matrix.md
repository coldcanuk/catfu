# License compatibility matrix (catfu → GPLv3)

| Component | License | Compatible with GPLv3? | Notes |
|-----------|---------|------------------------|-------|
| catfu (this project) | GPLv3 | n/a | primary license |
| yt-dlp (runtime, external) | Unlicense | **Yes** | Not linked; users install binary separately. Document in README/NOTICE. |
| modernc.org/sqlite | BSD-3-Clause | **Yes** | pure-Go SQLite |
| github.com/spf13/cobra | Apache-2.0 | **Yes** | CLI |
| github.com/spf13/viper | MIT | **Yes** | config |
| github.com/modelcontextprotocol/go-sdk | MIT (confirm at pin) | **Yes** | official MCP SDK |
| golang.org/x/* | BSD-3-Clause | **Yes** | transitive |
| stdlib | Go license (BSD-style) | **Yes** | |

## Summary

GPLv3 is **safe** for the catfu foundation as designed:

1. No GPL-incompatible copyleft (e.g. no proprietary SDK linkage).
2. yt-dlp is an **external process**, Unlicense public domain dedication — no distribution of yt-dlp source required.
3. Apache-2.0 / MIT / BSD deps are GPLv3-compatible when distributed as combined GPLv3 work (provide notices).

## NOTICE obligations

- Ship `LICENSE` (GPLv3 full text).
- Ship `NOTICE` or README section listing third-party licenses.
- Do not re-license third-party code; preserve their notices in `go.mod` / vendor if vendored (we use modules, not vendor by default).
