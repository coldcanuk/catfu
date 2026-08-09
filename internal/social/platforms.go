// Package social extracts social-media profile links and handles from free text
// (typically YouTube channel/video descriptions) using cheap regular expressions.
package social

// Platform identifies a supported social network or messaging service.
type Platform string

const (
	PlatformX         Platform = "x"
	PlatformThreads   Platform = "threads"
	PlatformFacebook  Platform = "facebook"
	PlatformInstagram Platform = "instagram"
	PlatformBluesky   Platform = "bluesky"
	PlatformNostr     Platform = "nostr"
	PlatformTikTok    Platform = "tiktok"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformMastodon  Platform = "mastodon"
	PlatformDiscord   Platform = "discord"
	PlatformGitHub    Platform = "github"
	PlatformGitLab    Platform = "gitlab"
	PlatformTelegram  Platform = "telegram"
	PlatformWeChat    Platform = "wechat"
	PlatformWhatsApp  Platform = "whatsapp"
	PlatformLine      Platform = "line"
)

// AllPlatforms is the stable set of supported platforms (order used in docs/tests).
var AllPlatforms = []Platform{
	PlatformX,
	PlatformThreads,
	PlatformFacebook,
	PlatformInstagram,
	PlatformBluesky,
	PlatformNostr,
	PlatformTikTok,
	PlatformLinkedIn,
	PlatformMastodon,
	PlatformDiscord,
	PlatformGitHub,
	PlatformGitLab,
	PlatformTelegram,
	PlatformWeChat,
	PlatformWhatsApp,
	PlatformLine,
}

// DisplayName returns a human-readable label for a platform key.
func DisplayName(p Platform) string {
	switch p {
	case PlatformX:
		return "X"
	case PlatformThreads:
		return "Threads"
	case PlatformFacebook:
		return "Facebook"
	case PlatformInstagram:
		return "Instagram"
	case PlatformBluesky:
		return "Bluesky"
	case PlatformNostr:
		return "Nostr"
	case PlatformTikTok:
		return "TikTok"
	case PlatformLinkedIn:
		return "LinkedIn"
	case PlatformMastodon:
		return "Mastodon"
	case PlatformDiscord:
		return "Discord"
	case PlatformGitHub:
		return "GitHub"
	case PlatformGitLab:
		return "GitLab"
	case PlatformTelegram:
		return "Telegram"
	case PlatformWeChat:
		return "WeChat"
	case PlatformWhatsApp:
		return "WhatsApp"
	case PlatformLine:
		return "LINE"
	default:
		return string(p)
	}
}

// ValidPlatform reports whether s is a known platform key.
func ValidPlatform(s string) bool {
	switch Platform(s) {
	case PlatformX, PlatformThreads, PlatformFacebook, PlatformInstagram,
		PlatformBluesky, PlatformNostr, PlatformTikTok, PlatformLinkedIn,
		PlatformMastodon, PlatformDiscord, PlatformGitHub, PlatformGitLab,
		PlatformTelegram, PlatformWeChat, PlatformWhatsApp, PlatformLine:
		return true
	default:
		return false
	}
}
