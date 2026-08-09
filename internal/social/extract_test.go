package social

import (
	"strings"
	"testing"
)

func TestExtract_AllPlatformURLs(t *testing.T) {
	bio := strings.Join([]string{
		"Follow us everywhere:",
		"https://twitter.com/AcmeNews",
		"https://x.com/AcmeNews2",
		"https://www.threads.net/@acme_threads",
		"https://www.facebook.com/AcmePage",
		"https://instagram.com/acme.ig",
		"https://bsky.app/profile/acme.bsky.social",
		"https://www.tiktok.com/@acme_tok",
		"https://www.linkedin.com/in/jane-doe",
		"https://www.linkedin.com/company/acme-corp",
		"https://mastodon.social/@acme",
		"https://discord.gg/acmeinvite",
		"https://github.com/acme-org",
		"https://gitlab.com/acme-gl",
		"https://t.me/acme_channel",
		"https://wa.me/14165551234",
		"https://chat.whatsapp.com/AbCdEfGhIjK",
		"https://line.me/R/ti/p/@acmelite",
		"npub1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
	}, "\n")

	links := Dedup(Extract(bio))
	got := map[Platform][]string{}
	for _, l := range links {
		got[l.Platform] = append(got[l.Platform], l.Handle)
	}

	wantPlatforms := []Platform{
		PlatformX, PlatformThreads, PlatformFacebook, PlatformInstagram,
		PlatformBluesky, PlatformTikTok, PlatformLinkedIn, PlatformMastodon,
		PlatformDiscord, PlatformGitHub, PlatformGitLab, PlatformTelegram,
		PlatformWhatsApp, PlatformLine, PlatformNostr,
	}
	for _, p := range wantPlatforms {
		if len(got[p]) == 0 {
			t.Errorf("expected at least one %s link, got none (all=%v)", p, platformsPresent(links))
		}
	}
	// X should have acmenews and acmenews2 (dedup case)
	if !containsFold(got[PlatformX], "AcmeNews") && !containsFold(got[PlatformX], "acmenews") {
		// handles stored as cleaned without forcing case lower for display; cleanHandle keeps case
		found := false
		for _, h := range got[PlatformX] {
			if strings.EqualFold(h, "AcmeNews") || strings.EqualFold(h, "AcmeNews2") {
				found = true
			}
		}
		if !found {
			t.Errorf("x handles = %v", got[PlatformX])
		}
	}
}

func TestExtract_LabeledHandles(t *testing.T) {
	text := `
Twitter: @labeled_x
IG: @labeled_ig
TikTok - @labeled_tt
GitHub: @labeled_gh
WeChat: mywechatid99
Telegram: @labeled_tg
`
	links := Dedup(Extract(text))
	by := indexByPlatform(links)
	assertHandle(t, by, PlatformX, "labeled_x")
	assertHandle(t, by, PlatformInstagram, "labeled_ig")
	assertHandle(t, by, PlatformTikTok, "labeled_tt")
	assertHandle(t, by, PlatformGitHub, "labeled_gh")
	assertHandle(t, by, PlatformWeChat, "mywechatid99")
	assertHandle(t, by, PlatformTelegram, "labeled_tg")
	for _, l := range links {
		if l.Confidence != ConfidenceMedium && l.Platform != PlatformWeChat {
			// wechat labeled is medium; URLs would be high — all labeled should be medium
		}
		if l.Platform != PlatformWeChat && l.Confidence != ConfidenceMedium {
			// allow only medium for pure labeled (no URL in text)
			if l.URL != "" && l.Confidence == ConfidenceHigh {
				continue
			}
			if l.Confidence != ConfidenceMedium {
				t.Errorf("%s confidence = %s want medium", l.Platform, l.Confidence)
			}
		}
	}
}

func TestExtract_TrailingPunctuation(t *testing.T) {
	text := "see https://x.com/punct_user."
	links := Extract(text)
	if len(links) == 0 {
		t.Fatal("no links")
	}
	if links[0].Handle != "punct_user" {
		t.Fatalf("handle = %q", links[0].Handle)
	}
}

func TestExtract_FalsePositives(t *testing.T) {
	text := strings.Join([]string{
		"Email me at editor@gmail.com for tips",
		"https://github.com/features/copilot",
		"https://github.com/settings/profile",
		"Bare mention @OnlyYouTube with no platform keyword",
		"https://instagram.com/p/AbCdPhotoID",
		"https://www.facebook.com/share/xyz",
	}, "\n")
	links := Extract(text)
	for _, l := range links {
		switch l.Platform {
		case PlatformMastodon:
			if strings.Contains(strings.ToLower(l.Handle), "gmail.com") {
				t.Errorf("email treated as mastodon: %+v", l)
			}
		case PlatformGitHub:
			if strings.EqualFold(l.Handle, "features") || strings.EqualFold(l.Handle, "settings") {
				t.Errorf("github denylist failed: %+v", l)
			}
		case PlatformInstagram:
			if strings.EqualFold(l.Handle, "p") {
				t.Errorf("instagram /p/ treated as profile: %+v", l)
			}
		case PlatformFacebook:
			if strings.EqualFold(l.Handle, "share") {
				t.Errorf("facebook /share treated as profile: %+v", l)
			}
		}
	}
	// bare @OnlyYouTube alone should not invent a platform
	for _, l := range links {
		if strings.EqualFold(l.Handle, "OnlyYouTube") {
			t.Errorf("bare @handle extracted without keyword: %+v", l)
		}
	}
}

func TestDedup_PrefersHighConfidence(t *testing.T) {
	text := "https://x.com/SameUser\nTwitter: @SameUser"
	links := Dedup(Extract(text))
	var x []Link
	for _, l := range links {
		if l.Platform == PlatformX && strings.EqualFold(l.Handle, "SameUser") {
			x = append(x, l)
		}
	}
	if len(x) != 1 {
		t.Fatalf("expected 1 deduped x link, got %d: %+v", len(x), x)
	}
	if x[0].Confidence != ConfidenceHigh {
		t.Errorf("confidence = %s", x[0].Confidence)
	}
	if x[0].URL == "" {
		t.Error("expected canonical URL")
	}
}

func TestExtractSources_StampsMetadata(t *testing.T) {
	links := ExtractSources([]SourceBlob{
		{Text: "https://t.me/fromchannel", Source: "channel", ChannelID: "UCabc"},
		{Text: "https://instagram.com/fromvideo", Source: "video", VideoID: "vid1", ChannelID: "UCabc"},
	})
	if len(links) < 2 {
		t.Fatalf("links = %d", len(links))
	}
	var sawCh, sawVid bool
	for _, l := range links {
		if l.Source == "channel" && l.ChannelID == "UCabc" {
			sawCh = true
		}
		if l.Source == "video" && l.VideoID == "vid1" {
			sawVid = true
		}
	}
	if !sawCh || !sawVid {
		t.Fatalf("metadata missing: %+v", links)
	}
}

func TestFilterPlatforms(t *testing.T) {
	links := Extract("https://x.com/a https://instagram.com/b https://t.me/ccccc")
	filtered := FilterPlatforms(links, []string{"x", "twitter", "telegram"})
	for _, l := range filtered {
		if l.Platform != PlatformX && l.Platform != PlatformTelegram {
			t.Errorf("unexpected platform %s", l.Platform)
		}
	}
}

func TestValidPlatform(t *testing.T) {
	if !ValidPlatform("x") || ValidPlatform("myspace") {
		t.Fatal("ValidPlatform broken")
	}
}

func platformsPresent(links []Link) []Platform {
	var out []Platform
	seen := map[Platform]bool{}
	for _, l := range links {
		if !seen[l.Platform] {
			seen[l.Platform] = true
			out = append(out, l.Platform)
		}
	}
	return out
}

func indexByPlatform(links []Link) map[Platform][]Link {
	m := map[Platform][]Link{}
	for _, l := range links {
		m[l.Platform] = append(m[l.Platform], l)
	}
	return m
}

func assertHandle(t *testing.T, by map[Platform][]Link, p Platform, want string) {
	t.Helper()
	list := by[p]
	for _, l := range list {
		if strings.EqualFold(l.Handle, want) {
			return
		}
	}
	t.Errorf("platform %s: want handle %q in %v", p, want, handlesOf(list))
}

func handlesOf(links []Link) []string {
	var s []string
	for _, l := range links {
		s = append(s, l.Handle)
	}
	return s
}

func containsFold(list []string, want string) bool {
	for _, s := range list {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

// Platforms that use https://host/@user must not be classified as Mastodon.
// Regression: channel bios often list youtube.com/@me and tiktok.com/@me; the
// greedy mastodon URL regex used to emit false "mastodon" rows alongside the
// correct TikTok/Threads extracts.
func TestExtract_MastodonRejectsOtherAtPathHosts(t *testing.T) {
	text := strings.Join([]string{
		"https://www.youtube.com/@ChannelHandle",
		"https://youtube.com/@ChannelHandle",
		"https://www.tiktok.com/@real_tiktok",
		"https://www.threads.net/@thread_user",
		"https://instagram.com/@not_a_profile",
		"https://mastodon.social/@real_masto",
		"https://fosstodon.org/@foss_dev",
		"https://github.com/acme-org",
		"https://t.me/acme_channel",
	}, "\n")
	links := Dedup(Extract(text))
	var mastodon []Link
	for _, l := range links {
		if l.Platform == PlatformMastodon {
			mastodon = append(mastodon, l)
			h := strings.ToLower(l.Handle)
			u := strings.ToLower(l.URL)
			for _, bad := range []string{
				"youtube.com", "tiktok.com", "threads.net", "instagram.com",
				"github.com", "t.me",
			} {
				if strings.Contains(h, bad) || strings.Contains(u, bad) {
					t.Errorf("mastodon false positive for %s: %+v", bad, l)
				}
			}
		}
	}
	var sawSocial, sawFoss bool
	for _, l := range mastodon {
		if strings.Contains(strings.ToLower(l.Handle), "mastodon.social") {
			sawSocial = true
		}
		if strings.Contains(strings.ToLower(l.Handle), "fosstodon.org") {
			sawFoss = true
		}
	}
	if !sawSocial || !sawFoss {
		t.Fatalf("expected real mastodon instances; got %+v", mastodon)
	}
	var sawTT bool
	for _, l := range links {
		if l.Platform == PlatformTikTok && strings.EqualFold(l.Handle, "real_tiktok") {
			sawTT = true
		}
	}
	if !sawTT {
		t.Fatalf("expected tiktok handle real_tiktok in %+v", links)
	}
}

func TestNormalizeMastodonHost(t *testing.T) {
	if normalizeMastodonHost("www.youtube.com") != "youtube.com" {
		t.Fatal(normalizeMastodonHost("www.youtube.com"))
	}
	if normalizeMastodonHost("m.youtube.com") != "youtube.com" {
		t.Fatal(normalizeMastodonHost("m.youtube.com"))
	}
	if !rejectMastodonHost("www.tiktok.com") || !rejectMastodonHost("music.youtube.com") {
		t.Fatal("expected platform hosts rejected")
	}
	if rejectMastodonHost("mastodon.social") || rejectMastodonHost("fosstodon.org") {
		t.Fatal("expected real instances allowed")
	}
}
