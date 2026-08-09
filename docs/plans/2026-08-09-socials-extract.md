# Plan: Social media link extraction (`catfu socials`)

**Date:** 2026-08-09  
**Branch / worktree:** `feat/socials-extract` @ `/workspace/repos/catfu-socials`  
**Repo:** https://github.com/coldcanuk/catfu  
**Base:** `main` @ `cd923fe`

---

## 0. Methodology

### Selected: **Research-Driven Incremental Delivery (RDID)**

RDID combines:

1. **Evidence-first research** (Phase 1) before code — required by this engagement and justified by catfu’s existing schema/CLI contracts.
2. **Incremental vertical slices** with small, verifiable tasks (unit tests → package → store → CLI → docs).
3. **Milestone-gated commits** so each step is reviewable and reversible.
4. **Definition of done** = `go test ./...`, `go vet ./...`, `make build`, CLI smoke with `--json`.

**Why not pure Waterfall / pure Shape Up alone**

| Method | Fit | Gap |
|--------|-----|-----|
| Waterfall | Good docs | Too slow to learn regex edge cases |
| Shape Up (6-week cycle) | Appetite framing | Overkill for a pure package + one CLI command |
| TDD-only | Great for extract.go | Misses store/CLI integration research |
| **RDID** | Research gates + TDD slices + catfu CLI conventions | Chosen |

**Evidence used to pick RDID**

- Schema: `videos.description` / `channels.description` already stored; often empty without `--full` (`docs/schema.md`).
- CLI: global `--json` / `--format` via `formatFlag()` (`internal/cli/root.go`).
- Output: `output.Encoder` JSON/table/csv (`internal/output/output.go`).
- Store: `GetChannel`, `GetVideo`, `SearchVideos` with empty query + `channel_id` lists videos.
- No existing social package — greenfield under `internal/social`.

---

## 1. Goal

Build a **regex-based** parser that scans catalogued channel and video description text for social profiles/links (X, Threads, Facebook, Instagram, Bluesky, Nostr, TikTok, LinkedIn, Mastodon, Discord, GitHub, GitLab, Telegram, WeChat, WhatsApp, Line) and expose results via:

```bash
catfu socials [channel]
catfu socials --video <id>
catfu socials @handle --json
```

**Non-goals (v1):** HTTP expansion of link-in-bio shorteners; persisting links in a new table; downloading media; OCR.

---

## 2. Architecture (target)

```text
internal/social/
  platforms.go   # Platform constants + display names
  extract.go     # Extract(text), Merge, Dedup, canonicalize
  extract_test.go
internal/cli/socials.go
internal/store/  # ListVideosByChannel helper (thin wrapper)
docs/socials.md  # user-facing notes
README.md        # command row + feature bullet
```

**Data flow**

```text
CLI resolve channel/video
  → load description text(s) from SQLite
  → social.Extract(text) per blob
  → attach source metadata
  → Dedup (platform+handle/url)
  → output.Encoder (table | json | csv)
```

**JSON result shape**

```json
{
  "channel_id": "UC…",
  "custom_url": "@example",
  "scanned_videos": 12,
  "scanned_sources": 13,
  "links": [
    {
      "platform": "x",
      "handle": "example",
      "url": "https://x.com/example",
      "raw": "https://twitter.com/example",
      "confidence": "high",
      "source": "video",
      "video_id": "abc",
      "channel_id": "UC…"
    }
  ]
}
```

---

## Phase 1 — Research only

### Milestone 1.1 — Schema & CLI contract research

- [x] Task 1.1.1 — Confirm description columns and full-vs-flat catalogue  
  Commands:
  ```bash
  sed -n '1,120p' docs/schema.md
  rg -n "Description|description" internal/store/models.go internal/catalogue/
  ```
- [x] Task 1.1.2 — Confirm channel resolution and video listing APIs  
  Commands:
  ```bash
  rg -n "func \(s \*Store\)" internal/store/store.go
  rg -n "GetChannel|SearchVideos|formatFlag" internal/cli/
  ```
- [x] Task 1.1.3 — Record findings in this plan §Research notes (1.1)

**Commit:** `docs(plan): research schema and CLI contracts for socials`

### Milestone 1.2 — Platform URL / handle pattern research

- [x] Task 1.2.1 — Enumerate canonical domains per platform (table in plan)
- [x] Task 1.2.2 — Note false-positive traps (emails, github.com/features, bare @YouTube)
- [x] Task 1.2.3 — Inspect sample yt-dlp research dumps for description richness  
  Commands:
  ```bash
  python3 -c "import json; ..."  # sample scan
  ```

**Commit:** `docs(plan): research social platform URL/handle patterns`

### Milestone 1.3 — Plan synthesis (update plan with all new knowledge)

- [x] Task 1.3.1 — Write finalized regex strategy, CLI flags, file list, test matrix
- [x] Task 1.3.2 — **Final task of Phase 1:** rewrite §Research notes + §Finalized design in this file

**Commit:** `docs(plan): finalize socials extract design after research`

---

## Phase 2 — Core `internal/social` package

### Milestone 2.1 — Types + platform registry

- [x] Task 2.1.1 — Create `platforms.go` with Platform constants and `AllPlatforms`
- [x] Task 2.1.2 — Create `Link` / `Confidence` types in `extract.go` (stubs ok)

**Commit:** `feat(social): add platform constants and link types`

### Milestone 2.2 — URL + labeled-handle extractors

- [x] Task 2.2.1 — Implement domain URL regexes for all 16 platforms
- [x] Task 2.2.2 — Implement keyword-labeled `@handle` patterns
- [x] Task 2.2.3 — Implement Nostr npub + Mastodon user@host (email denylist)
- [x] Task 2.2.4 — Normalize (strip @, trailing punct, query/hash), canonical URLs, Dedup

**Commit:** `feat(social): implement regex extractors and normalization`

### Milestone 2.3 — Unit tests

- [x] Task 2.3.1 — Table-driven tests: one high-confidence URL per platform
- [x] Task 2.3.2 — Labeled handles, trailing punctuation, false positives
- [x] Task 2.3.3 — Run: `go test ./internal/social/ -count=1 -v`

**Commit:** `test(social): table-driven extract coverage for all platforms`

---

## Phase 3 — Store + CLI + docs

### Milestone 3.1 — Store helper

- [x] Task 3.1.1 — Add `ListVideosByChannel(ctx, channelID, limit)`  
  Snippet: query videos ordered by upload_date DESC with COALESCE(description,'')

**Commit:** `feat(store): list videos by channel for socials scan`

### Milestone 3.2 — CLI command

- [x] Task 3.2.1 — `internal/cli/socials.go` with flags: `--video`, `--source`, `--platform`, `--unique`, `--limit`
- [x] Task 3.2.2 — Register in `root.go` `AddCommand`
- [x] Task 3.2.3 — Table + JSON output via `output` package

**Commit:** `feat(cli): add catfu socials command with --json`

### Milestone 3.3 — Documentation

- [x] Task 3.3.1 — `docs/socials.md` usage + `--full` requirement
- [x] Task 3.3.2 — README feature bullet + command table row

**Commit:** `docs: document catfu socials extraction`

---

## Phase 4 — Verify, merge, hygiene

### Milestone 4.1 — Full verification

- [x] Task 4.1.1 — `go test ./... && go vet ./... && make build`
- [x] Task 4.1.2 — Smoke: seed temp SQLite with synthetic descriptions, run `./bin/catfu socials --json`

**Commit:** `test: socials CLI smoke fixtures (if any test helpers)` or skip if pure package tests suffice

### Milestone 4.2 — Ship

- [x] Task 4.2.1 — Push branch, open PR, merge to main
- [x] Task 4.2.2 — Delete feature worktree and local branch
- [x] Task 4.2.3 — Delete merged remote litter branches (`wt/*`)
- [x] Task 4.2.4 — Confirm only `main` remains locally and on origin

**Final commit on branch before merge:** as needed  
**Post-merge:** no extra commit unless conflict resolution

---

## Research notes (filled during Phase 1)

### 1.1 Schema & CLI

**Sources of text**

| Source | Column | How populated |
|--------|--------|---------------|
| Channel about | `channels.description` | yt-dlp channel metadata in catalogue |
| Video description | `videos.description` | often **empty** in flat mode; reliable with `--full` |

**Store APIs available**

- `GetChannel(ctx, idOrHandle)` — resolves `UC…`, `@handle`, custom_url, URL substring
- `GetVideo(ctx, id)` — single video with description
- `SearchVideos(ctx, SearchParams{ChannelID, Limit, …})` — empty query lists by channel
- **Gap:** no dedicated `ListVideosByChannel`; add thin helper for clarity + default high limit

**CLI contracts to mirror**

- Global `--json` / `--format table|json|csv` via `formatFlag()`
- Commands use `openStore` + `defer st.Close()` + `output.New(formatFlag()).WriteValue(...)`
- Table mode for lists uses `WriteRows(headers, rows)` (see `list.go`, `search.go`)
- Register via `root.AddCommand(newSocialsCmd())`

**Sample research dumps**

- `research/yt-dlp-sample.ndjson`: 5 flat entries, **0** non-empty video descriptions → unit tests must use synthetic fixtures
- Channel JSON description is org blurb only (no social URLs in sample)

**Implication:** v1 is **read-time extraction** over stored text; document `--full` requirement; no schema migration.

### 1.2 Platform patterns

| Platform | Key | High-confidence URL shapes | Labeled keywords | Notes |
|----------|-----|----------------------------|------------------|-------|
| X | `x` | `x.com/`, `twitter.com/` | twitter, x | Handle max 15; map twitter→x |
| Threads | `threads` | `threads.net/@` | threads | |
| Facebook | `facebook` | `facebook.com/`, `fb.com/`, `fb.me/` | facebook, fb | Skip `/share`, `/watch`, `/reel` noise carefully |
| Instagram | `instagram` | `instagram.com/`, `instagr.am/` | instagram, ig, insta | Skip `/p/`, `/reel/`, `/stories/` for profile-only? Keep path as handle only for profile URLs |
| Bluesky | `bluesky` | `bsky.app/profile/` | bluesky, bsky | Handle may be `user.bsky.social` |
| Nostr | `nostr` | `npub1…` bech32 | nostr | Also `nprofile1` optional |
| TikTok | `tiktok` | `tiktok.com/@` | tiktok, tt | |
| LinkedIn | `linkedin` | `linkedin.com/in/`, `…/company/` | linkedin | Preserve in/company prefix in handle |
| Mastodon | `mastodon` | `https://host/@user`, `user@host` | mastodon | Email domain denylist |
| Discord | `discord` | `discord.gg/`, `discord.com/invite/` | discord | Invite code as handle |
| GitHub | `github` | `github.com/{user}` | github, gh | Denylist: features, settings, pull, issues, org, marketplace, topics, collections, explore, login, signup, about, pricing, enterprise, customer-stories, security, readme, sponsors, codespaces |
| GitLab | `gitlab` | `gitlab.com/{user}` | gitlab, gl | Similar path denylist |
| Telegram | `telegram` | `t.me/`, `telegram.me/` | telegram, tg | Skip `t.me/c/` private |
| WeChat | `wechat` | rare public URL | wechat, weixin | Labeled ID medium confidence |
| WhatsApp | `whatsapp` | `wa.me/`, `api.whatsapp.com/`, `chat.whatsapp.com/` | whatsapp, wa | Phone or invite |
| Line | `line` | `line.me/`, `lin.ee/` | line | |

**False positives to reject**

- Emails (`user@gmail.com`) as Mastodon
- Bare `@YouTubeHandle` without platform keyword
- `github.com/features`, `linkedin.com/feed`
- `t.co/` short links (optional: ignore — no handle)
- Trailing punctuation: `x.com/foo.` → strip

### Finalized design (end of Phase 1)

1. **Package** `internal/social` — pure, no I/O:
   - `type Platform string`, `type Link struct`, `Extract(text string) []Link`
   - `Dedup(links []Link) []Link`, `FilterPlatforms(links, set)`
   - Precompiled `regexp.Regexp` via `sync.Once` or package-level `MustCompile`
2. **Two-pass extraction:** URL domains (high) → labeled handles (medium) → Nostr/Mastodon specials
3. **CLI** `catfu socials [channel|@handle]`:
   - `--video ID` scan one video
   - `--source all|channel|videos` (default all)
   - `--platform` repeatable/csv filter
   - `--unique` default true (dedupe platform+handle)
   - `--limit` max videos (default 10000)
   - Honors global `--json`
4. **Store** add `ListVideosByChannel(ctx, channelID string, limit int) ([]Video, error)`
5. **No DB migration** in v1
6. **Tests** table-driven synthetic multi-line bio covering all platforms + false positives
7. **Docs** `docs/socials.md` + README row; note catalogue with `--full`
8. **MCP** deferred (non-goal v1) unless trivial; skip for scope control

**File touch list**

```
internal/social/platforms.go
internal/social/extract.go
internal/social/extract_test.go
internal/store/store.go          # ListVideosByChannel
internal/store/store_test.go     # optional
internal/cli/socials.go
internal/cli/root.go
docs/socials.md
docs/plans/2026-08-09-socials-extract.md
README.md
```

**CLI examples (acceptance)**

```bash
catfu socials @SomeChannel
catfu socials @SomeChannel --json
catfu socials --video dQw4w9WgXcQ --json
catfu socials UC… --source channel --platform x,instagram --json
```

---

## Risk register

| Risk | Mitigation |
|------|------------|
| Empty descriptions without `--full` | Document; CLI reports `scanned_*` and empty links |
| False positive GitHub/GitLab paths | Path denylist |
| Mastodon vs email | Common email domain denylist |
| Rate of regex maintenance | Single platform table in one file |


---

## Plan audit (pre-execution)

| Issue found | Fix applied |
|-------------|-------------|
| Sample dumps lack social URLs | Rely on synthetic unit tests, not live yt-dlp |
| SearchVideos already lists by channel | Still add ListVideosByChannel for clearer API + description-friendly query |
| MCP in earlier chat design | Explicitly deferred in finalized design |
| Milestone commits must be real | Each milestone ends with `git add . && git commit` |
| Hygiene requires only main | Delete merged `origin/wt/*` after ship |

**Definition of Done**

1. `go test ./...` pass  
2. `go vet ./...` pass  
3. `make build` produces `bin/catfu`  
4. `bin/catfu socials --help` works  
5. Temp DB smoke returns expected platforms in JSON  
6. PR merged to `main`; worktrees/branches cleaned  

