package social

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// Confidence grades how reliable a match is.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"   // domain URL match
	ConfidenceMedium Confidence = "medium" // labeled handle / best-effort ID
)

// Link is one extracted social profile or invite.
type Link struct {
	Platform   Platform   `json:"platform"`
	Handle     string     `json:"handle,omitempty"`
	URL        string     `json:"url,omitempty"`
	Raw        string     `json:"raw"`
	Confidence Confidence `json:"confidence"`
	Source     string     `json:"source,omitempty"` // channel | video
	VideoID    string     `json:"video_id,omitempty"`
	ChannelID  string     `json:"channel_id,omitempty"`
}

// SourceBlob is a named text blob to scan (channel about or video description).
type SourceBlob struct {
	Text      string
	Source    string // channel | video
	VideoID   string
	ChannelID string
}

var (
	// URL-oriented (high confidence)
	reX = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:twitter\.com|x\.com)/@?([A-Za-z0-9_]{1,15})\b`)
	reThreads = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?threads\.net/@([A-Za-z0-9._]{1,30})\b`)
	reFacebook = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:facebook\.com|fb\.com|fb\.me)/([A-Za-z0-9.]+)(?:/[^\s]*)?`)
	reInstagram = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:instagram\.com|instagr\.am)/([A-Za-z0-9._]{1,30})/?\b`)
	reBluesky = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?bsky\.app/profile/([A-Za-z0-9.\-_]+)`)
	reTikTok = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?tiktok\.com/@([A-Za-z0-9._]{2,24})\b`)
	reLinkedInIn = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?linkedin\.com/(in|company)/([A-Za-z0-9\-_%]+)`)
	reDiscord = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:discord\.gg|discord(?:app)?\.com/invite)/([A-Za-z0-9-]+)`)
	reGitHub = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?github\.com/([A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)\b`)
	reGitLab = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?gitlab\.com/([A-Za-z0-9](?:[A-Za-z0-9_.-]*[A-Za-z0-9])?)\b`)
	reTelegram = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:t\.me|telegram\.me|telegram\.dog)/([A-Za-z0-9_]{5,32})\b`)
	reWhatsApp = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:wa\.me/(\+?[0-9]{7,15})|api\.whatsapp\.com/send\?[^ \t\n\r]*phone=(\+?[0-9]{7,15})|chat\.whatsapp\.com/([A-Za-z0-9]+))`)
	reLine = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:line\.me/R/ti/p/(@?[%A-Za-z0-9._\-]+)|line\.me/ti/p/(@?[%A-Za-z0-9._\-]+)|lin\.ee/([A-Za-z0-9]+))`)

	// Specials
	reNostrNpub = regexp.MustCompile(`\b(npub1[02-9ac-hj-np-z]{58,})\b`)
	reMastodonURL = regexp.MustCompile(`(?i)(?:https?://)([A-Za-z0-9.-]+\.[A-Za-z]{2,})/@([A-Za-z0-9_]+)`)
	reMastodonAcct = regexp.MustCompile(`\b@?([A-Za-z0-9_]+)@([A-Za-z0-9.-]+\.[A-Za-z]{2,})\b`)

	// Labeled bare handles (medium)
	// keyword group maps via labelToPlatform
	reLabeled = regexp.MustCompile(`(?i)\b(twitter|tweets?|x\.com|\bx\b|threads|instagram|insta|\big\b|facebook|\bfb\b|bluesky|\bbsky\b|tiktok|\btt\b|linkedin|mastodon|discord|github|\bgh\b|gitlab|\bgl\b|telegram|\btg\b|wechat|weixin|whatsapp|\bwa\b|\bline\b|nostr)\b\s*[:：\-–—|]?\s*@([A-Za-z0-9_.]{2,40})\b`)
	reLabeledWeChat = regexp.MustCompile(`(?i)\b(?:wechat|weixin)\b\s*[:：\-–—|]?\s*([A-Za-z0-9_\-]{3,40})\b`)
)

// githubPathDenylist segments that are not user/org profiles.
var githubPathDenylist = map[string]bool{
	"features": true, "settings": true, "pulls": true, "issues": true,
	"marketplace": true, "topics": true, "collections": true, "explore": true,
	"login": true, "signup": true, "join": true, "about": true, "pricing": true,
	"enterprise": true, "security": true, "readme": true, "sponsors": true,
	"codespaces": true, "customer-stories": true, "orgs": true, "organizations": true,
	"notifications": true, "new": true, "search": true, "site": true, "apps": true,
	"github": true, "home": true, "account": true, "sessions": true, "stars": true,
	"trending": true, "events": true, "codes": true, "copilot": true,
}

var gitlabPathDenylist = map[string]bool{
	"explore": true, "users": true, "help": true, "admin": true, "groups": true,
	"projects": true, "dashboard": true, "search": true, "signin": true, "users_sign_in": true,
	"gitlab": true, "about": true, "pricing": true, "solutions": true,
}

var facebookPathDenylist = map[string]bool{
	"share": true, "watch": true, "reel": true, "stories": true, "groups": true,
	"events": true, "marketplace": true, "gaming": true, "login": true, "dialog": true,
	"sharer": true, "tr": true, "privacy": true, "help": true, "policies": true,
	"pages": true, "public": true, "permalink.php": true, "photo.php": true,
}

var instagramPathDenylist = map[string]bool{
	"p": true, "reel": true, "reels": true, "stories": true, "explore": true,
	"accounts": true, "direct": true, "tv": true, "about": true, "legal": true,
}

// common email domains rejected as Mastodon instances.
var emailDomainDenylist = map[string]bool{
	"gmail.com": true, "googlemail.com": true, "yahoo.com": true, "yahoo.ca": true,
	"hotmail.com": true, "outlook.com": true, "live.com": true, "msn.com": true,
	"icloud.com": true, "me.com": true, "mac.com": true, "aol.com": true,
	"protonmail.com": true, "proton.me": true, "pm.me": true, "yandex.com": true,
	"mail.com": true, "zoho.com": true, "gmx.com": true, "gmx.net": true,
	"qq.com": true, "163.com": true, "126.com": true, "sina.com": true,
	"example.com": true, "example.org": true, "test.com": true,
	"email.com": true, "fastmail.com": true,
}

// Extract finds social links/handles in text. Order is match order; use Dedup to merge.
func Extract(text string) []Link {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var out []Link

	// High confidence URL / specials first
	out = append(out, findSubmatchLinks(text, reX, PlatformX, ConfidenceHigh, func(m []string) (handle, canon string) {
		h := cleanHandle(m[1])
		return h, "https://x.com/" + h
	})...)
	out = append(out, findSubmatchLinks(text, reThreads, PlatformThreads, ConfidenceHigh, func(m []string) (string, string) {
		h := cleanHandle(m[1])
		return h, "https://www.threads.net/@" + h
	})...)
	out = append(out, findFacebook(text)...)
	out = append(out, findInstagram(text)...)
	out = append(out, findSubmatchLinks(text, reBluesky, PlatformBluesky, ConfidenceHigh, func(m []string) (string, string) {
		h := strings.TrimSpace(m[1])
		return h, "https://bsky.app/profile/" + h
	})...)
	out = append(out, findSubmatchLinks(text, reTikTok, PlatformTikTok, ConfidenceHigh, func(m []string) (string, string) {
		h := cleanHandle(m[1])
		return h, "https://www.tiktok.com/@" + h
	})...)
	out = append(out, findLinkedIn(text)...)
	out = append(out, findSubmatchLinks(text, reDiscord, PlatformDiscord, ConfidenceHigh, func(m []string) (string, string) {
		h := m[1]
		return h, "https://discord.gg/" + h
	})...)
	out = append(out, findGitHost(text, reGitHub, PlatformGitHub, githubPathDenylist, "https://github.com/")...)
	out = append(out, findGitHost(text, reGitLab, PlatformGitLab, gitlabPathDenylist, "https://gitlab.com/")...)
	out = append(out, findTelegram(text)...)
	out = append(out, findWhatsApp(text)...)
	out = append(out, findLine(text)...)
	out = append(out, findNostr(text)...)
	out = append(out, findMastodon(text)...)

	// Medium: labeled handles
	out = append(out, findLabeled(text)...)

	return out
}

// ExtractSources runs Extract on each blob and stamps source metadata.
func ExtractSources(blobs []SourceBlob) []Link {
	var out []Link
	for _, b := range blobs {
		for _, l := range Extract(b.Text) {
			l.Source = b.Source
			l.VideoID = b.VideoID
			l.ChannelID = b.ChannelID
			out = append(out, l)
		}
	}
	return out
}

// Dedup merges links by (platform, lower(handle|url)), preferring high confidence and non-empty URL.
func Dedup(links []Link) []Link {
	type key struct {
		p Platform
		h string
	}
	best := map[key]Link{}
	order := []key{}
	for _, l := range links {
		h := strings.ToLower(l.Handle)
		if h == "" {
			h = strings.ToLower(l.URL)
		}
		if h == "" {
			h = strings.ToLower(l.Raw)
		}
		k := key{l.Platform, h}
		if prev, ok := best[k]; ok {
			best[k] = preferLink(prev, l)
			continue
		}
		best[k] = l
		order = append(order, k)
	}
	out := make([]Link, 0, len(order))
	for _, k := range order {
		out = append(out, best[k])
	}
	return out
}

// FilterPlatforms keeps only links whose platform is in allow (empty allow = all).
func FilterPlatforms(links []Link, allow []string) []Link {
	if len(allow) == 0 {
		return links
	}
	set := map[Platform]bool{}
	for _, a := range allow {
		a = strings.TrimSpace(strings.ToLower(a))
		if a == "twitter" {
			a = "x"
		}
		if ValidPlatform(a) {
			set[Platform(a)] = true
		}
	}
	if len(set) == 0 {
		return links
	}
	var out []Link
	for _, l := range links {
		if set[l.Platform] {
			out = append(out, l)
		}
	}
	return out
}

// SortLinks orders by platform name then handle for stable CLI output.
func SortLinks(links []Link) {
	sort.SliceStable(links, func(i, j int) bool {
		if links[i].Platform != links[j].Platform {
			return links[i].Platform < links[j].Platform
		}
		if links[i].Handle != links[j].Handle {
			return links[i].Handle < links[j].Handle
		}
		return links[i].VideoID < links[j].VideoID
	})
}

func preferLink(a, b Link) Link {
	// Prefer high confidence
	if a.Confidence != ConfidenceHigh && b.Confidence == ConfidenceHigh {
		a = b
	} else if a.Confidence == b.Confidence {
		// Prefer one with URL
		if a.URL == "" && b.URL != "" {
			a = b
		}
	}
	// Preserve first-seen source metadata if winner already has it
	if a.Source == "" {
		a.Source = b.Source
	}
	if a.VideoID == "" {
		a.VideoID = b.VideoID
	}
	if a.ChannelID == "" {
		a.ChannelID = b.ChannelID
	}
	return a
}

func findSubmatchLinks(text string, re *regexp.Regexp, p Platform, conf Confidence, canon func([]string) (handle, url string)) []Link {
	ms := re.FindAllStringSubmatchIndex(text, -1)
	if len(ms) == 0 {
		return nil
	}
	var out []Link
	for _, idx := range ms {
		full := text[idx[0]:idx[1]]
		// rebuild submatch strings
		m := make([]string, len(idx)/2)
		for i := 0; i < len(m); i++ {
			if idx[2*i] >= 0 {
				m[i] = text[idx[2*i]:idx[2*i+1]]
			}
		}
		handle, u := canon(m)
		if handle == "" && u == "" {
			continue
		}
		out = append(out, Link{
			Platform:   p,
			Handle:     handle,
			URL:        u,
			Raw:        strings.TrimRight(full, ".,;:!?)>]}\"'"),
			Confidence: conf,
		})
	}
	return out
}

func findFacebook(text string) []Link {
	ms := reFacebook.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		path := strings.Trim(m[1], "/")
		if path == "" {
			continue
		}
		seg := strings.Split(path, "/")[0]
		segLower := strings.ToLower(seg)
		if facebookPathDenylist[segLower] {
			continue
		}
		// profile.php?id= handled poorly by path regex — skip empty
		if strings.HasPrefix(segLower, "profile.php") {
			continue
		}
		out = append(out, Link{
			Platform:   PlatformFacebook,
			Handle:     seg,
			URL:        "https://www.facebook.com/" + seg,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findInstagram(text string) []Link {
	ms := reInstagram.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		h := cleanHandle(m[1])
		if instagramPathDenylist[strings.ToLower(h)] {
			continue
		}
		out = append(out, Link{
			Platform:   PlatformInstagram,
			Handle:     h,
			URL:        "https://www.instagram.com/" + h,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findLinkedIn(text string) []Link {
	ms := reLinkedInIn.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		kind := strings.ToLower(m[1])
		slug, _ := url.PathUnescape(m[2])
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}
		handle := kind + "/" + slug
		out = append(out, Link{
			Platform:   PlatformLinkedIn,
			Handle:     handle,
			URL:        "https://www.linkedin.com/" + kind + "/" + slug,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findGitHost(text string, re *regexp.Regexp, p Platform, deny map[string]bool, prefix string) []Link {
	ms := re.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		h := m[1]
		if deny[strings.ToLower(h)] {
			continue
		}
		out = append(out, Link{
			Platform:   p,
			Handle:     h,
			URL:        prefix + h,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findTelegram(text string) []Link {
	ms := reTelegram.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		h := m[1]
		lower := strings.ToLower(h)
		// skip joinchat internals and private c/ paths (regex already avoids c/ mostly)
		if lower == "joinchat" || lower == "addstickers" || lower == "share" || lower == "socks" || lower == "proxy" {
			continue
		}
		out = append(out, Link{
			Platform:   PlatformTelegram,
			Handle:     h,
			URL:        "https://t.me/" + h,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findWhatsApp(text string) []Link {
	ms := reWhatsApp.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		var handle, u string
		switch {
		case m[1] != "":
			handle = m[1]
			u = "https://wa.me/" + strings.TrimPrefix(handle, "+")
		case m[2] != "":
			handle = m[2]
			u = "https://wa.me/" + strings.TrimPrefix(handle, "+")
		case m[3] != "":
			handle = m[3]
			u = "https://chat.whatsapp.com/" + handle
		}
		if handle == "" {
			continue
		}
		out = append(out, Link{
			Platform:   PlatformWhatsApp,
			Handle:     handle,
			URL:        u,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findLine(text string) []Link {
	ms := reLine.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		var handle string
		for i := 1; i < len(m); i++ {
			if m[i] != "" {
				handle = m[i]
				break
			}
		}
		if handle == "" {
			continue
		}
		// URL-decode percent handles from line.me
		if dec, err := url.PathUnescape(handle); err == nil {
			handle = dec
		}
		handle = strings.TrimPrefix(handle, "@")
		out = append(out, Link{
			Platform:   PlatformLine,
			Handle:     handle,
			URL:        "https://line.me/R/ti/p/@" + handle,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findNostr(text string) []Link {
	ms := reNostrNpub.FindAllStringSubmatch(text, -1)
	var out []Link
	for _, m := range ms {
		h := m[1]
		out = append(out, Link{
			Platform:   PlatformNostr,
			Handle:     h,
			URL:        "nostr:" + h,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	return out
}

func findMastodon(text string) []Link {
	var out []Link
	for _, m := range reMastodonURL.FindAllStringSubmatch(text, -1) {
		host := strings.ToLower(m[1])
		user := m[2]
		if emailDomainDenylist[host] {
			continue
		}
		handle := user + "@" + host
		out = append(out, Link{
			Platform:   PlatformMastodon,
			Handle:     handle,
			URL:        "https://" + host + "/@" + user,
			Raw:        m[0],
			Confidence: ConfidenceHigh,
		})
	}
	for _, m := range reMastodonAcct.FindAllStringSubmatch(text, -1) {
		user := m[1]
		host := strings.ToLower(m[2])
		if emailDomainDenylist[host] {
			continue
		}
		// skip if host looks like a normal email (no further heuristic)
		handle := user + "@" + host
		out = append(out, Link{
			Platform:   PlatformMastodon,
			Handle:     handle,
			URL:        "https://" + host + "/@" + user,
			Raw:        m[0],
			Confidence: ConfidenceMedium,
		})
	}
	return out
}

func findLabeled(text string) []Link {
	var out []Link
	for _, m := range reLabeled.FindAllStringSubmatch(text, -1) {
		kw := strings.ToLower(m[1])
		h := cleanHandle(m[2])
		p := labelToPlatform(kw)
		if p == "" || h == "" {
			continue
		}
		out = append(out, Link{
			Platform:   p,
			Handle:     h,
			URL:        canonicalFor(p, h),
			Raw:        strings.TrimSpace(m[0]),
			Confidence: ConfidenceMedium,
		})
	}
	for _, m := range reLabeledWeChat.FindAllStringSubmatch(text, -1) {
		h := m[1]
		// skip if the capture is another keyword
		if strings.EqualFold(h, "id") || strings.EqualFold(h, "me") {
			continue
		}
		out = append(out, Link{
			Platform:   PlatformWeChat,
			Handle:     h,
			Raw:        strings.TrimSpace(m[0]),
			Confidence: ConfidenceMedium,
		})
	}
	return out
}

func labelToPlatform(kw string) Platform {
	switch strings.ToLower(kw) {
	case "twitter", "tweet", "tweets", "x.com", "x":
		return PlatformX
	case "threads":
		return PlatformThreads
	case "instagram", "insta", "ig":
		return PlatformInstagram
	case "facebook", "fb":
		return PlatformFacebook
	case "bluesky", "bsky":
		return PlatformBluesky
	case "tiktok", "tt":
		return PlatformTikTok
	case "linkedin":
		return PlatformLinkedIn
	case "mastodon":
		return PlatformMastodon
	case "discord":
		return PlatformDiscord
	case "github", "gh":
		return PlatformGitHub
	case "gitlab", "gl":
		return PlatformGitLab
	case "telegram", "tg":
		return PlatformTelegram
	case "wechat", "weixin":
		return PlatformWeChat
	case "whatsapp", "wa":
		return PlatformWhatsApp
	case "line":
		return PlatformLine
	case "nostr":
		return PlatformNostr
	default:
		return ""
	}
}

func canonicalFor(p Platform, handle string) string {
	h := strings.TrimPrefix(handle, "@")
	switch p {
	case PlatformX:
		return "https://x.com/" + h
	case PlatformThreads:
		return "https://www.threads.net/@" + h
	case PlatformInstagram:
		return "https://www.instagram.com/" + h
	case PlatformFacebook:
		return "https://www.facebook.com/" + h
	case PlatformBluesky:
		return "https://bsky.app/profile/" + h
	case PlatformTikTok:
		return "https://www.tiktok.com/@" + h
	case PlatformGitHub:
		return "https://github.com/" + h
	case PlatformGitLab:
		return "https://gitlab.com/" + h
	case PlatformTelegram:
		return "https://t.me/" + h
	case PlatformDiscord:
		return "https://discord.gg/" + h
	case PlatformNostr:
		if strings.HasPrefix(h, "npub1") {
			return "nostr:" + h
		}
		return ""
	default:
		return ""
	}
}

func cleanHandle(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "@")
	h = strings.TrimRight(h, ".,;:!?)>]}\"'")
	return h
}
