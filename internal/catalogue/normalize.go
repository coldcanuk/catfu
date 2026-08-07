package catalogue

import (
	"net/url"
	"strings"
)

// NormalizeChannelURL turns handles and partial paths into a YouTube URL.
// Prefers the /videos tab for listing when a channel root is given.
func NormalizeChannelURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	// Bare handle
	if strings.HasPrefix(raw, "@") {
		return "https://www.youtube.com/" + raw + "/videos"
	}
	// UC channel id
	if strings.HasPrefix(raw, "UC") && !strings.Contains(raw, "/") && len(raw) >= 20 {
		return "https://www.youtube.com/channel/" + raw + "/videos"
	}
	// Missing scheme
	if !strings.Contains(raw, "://") {
		if strings.HasPrefix(raw, "youtube.com") || strings.HasPrefix(raw, "www.youtube.com") || strings.HasPrefix(raw, "m.youtube.com") {
			raw = "https://" + raw
		} else if strings.Contains(raw, "youtube.com/") {
			raw = "https://" + strings.TrimPrefix(raw, "//")
		} else {
			// treat as handle without @
			return "https://www.youtube.com/@" + strings.TrimPrefix(raw, "@") + "/videos"
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	// Add /videos if path is channel root
	path := strings.TrimSuffix(u.Path, "/")
	parts := strings.Split(path, "/")
	// /@handle, /channel/UC, /c/name, /user/name without further segment
	if len(parts) == 2 && parts[1] != "" && !strings.EqualFold(parts[1], "watch") {
		// /@handle
		if strings.HasPrefix(parts[1], "@") {
			u.Path = path + "/videos"
			return u.String()
		}
	}
	if len(parts) == 3 {
		if parts[1] == "channel" || parts[1] == "c" || parts[1] == "user" {
			u.Path = path + "/videos"
			return u.String()
		}
	}
	return u.String()
}
