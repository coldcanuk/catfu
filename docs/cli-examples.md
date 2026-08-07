# catfu CLI examples

```bash
# Health check
catfu doctor --json

# Catalogue a channel (metadata only; requires yt-dlp)
catfu catalogue @CTVNews --limit 50
catfu catalogue https://www.youtube.com/@veritasium/videos --full

# Search local catalogue
catfu search "climate" --after 2024-01-01 --limit 20 --json
catfu search --channel @CTVNews --limit 10

# Web search (Brave BYOK)
export BRAVE_API_KEY=…
catfu web "youtube channel analytics best practices" --json

# Management
catfu list
catfu status @CTVNews
catfu update @CTVNews
catfu info dQw4w9WgXcQ
catfu delete @CTVNews --force
catfu export --channel @CTVNews --format csv > out.csv
catfu stats --json

# MCP
catfu mcp
```
