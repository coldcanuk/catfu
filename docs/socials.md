# Social link extraction (`catfu socials`)

`catfu socials` scans **already catalogued** channel about text and video
descriptions for social media profiles using **regular expressions** only
(no network, no NLP).

## Supported platforms

| Key | Examples matched |
|-----|------------------|
| `x` | `x.com/`, `twitter.com/`, `Twitter: @user` |
| `threads` | `threads.net/@user` |
| `facebook` | `facebook.com/`, `fb.com/` |
| `instagram` | `instagram.com/`, `IG: @user` |
| `bluesky` | `bsky.app/profile/` |
| `nostr` | `npub1…` |
| `tiktok` | `tiktok.com/@`, `TikTok: @user` |
| `linkedin` | `linkedin.com/in/`, `linkedin.com/company/` |
| `mastodon` | `https://host/@user`, `user@host` (email + non-fediverse hosts excluded; YouTube/TikTok/Threads `/@` paths are not Mastodon) |
| `discord` | `discord.gg/`, `discord.com/invite/` |
| `github` | `github.com/user` (non-profile paths denylisted) |
| `gitlab` | `gitlab.com/user` |
| `telegram` | `t.me/`, `Telegram: @user` |
| `wechat` | `WeChat: id` (medium confidence) |
| `whatsapp` | `wa.me/`, `chat.whatsapp.com/` |
| `line` | `line.me/`, `lin.ee/` |

## Requirements

Video descriptions are **often empty** after a fast/flat catalogue. Prefer:

```bash
catfu catalogue @SomeChannel --full --limit 100
```

Channel about text is still scanned even without `--full`.

## Usage

```bash
catfu socials @SomeChannel
catfu socials @SomeChannel --json
catfu socials @SomeChannel --platform x,instagram,telegram --json
catfu socials @SomeChannel --source channel
catfu socials --video VIDEO_ID --json
```

### Flags

| Flag | Default | Meaning |
|------|---------|---------|
| `--video` | | Scan one video id |
| `--source` | `all` | `all` \| `channel` \| `videos` |
| `--platform` | all | Filter (repeatable / comma-separated) |
| `--unique` | `true` | Dedupe by platform+handle |
| `--limit` | `10000` | Max videos to load |

Global `--json` / `--format` apply as usual.

## JSON shape

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

## Accuracy notes

- High confidence = domain URL match.
- Medium confidence = keyword-labeled bare handle (e.g. `IG: @foo`).
- Link-in-bio shorteners (`linktr.ee`, etc.) are **not** expanded.
- Mentions of third-party accounts are extracted; frequency ranking is not applied in v1.
