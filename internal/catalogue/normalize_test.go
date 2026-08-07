package catalogue

import "testing"

func TestNormalizeChannelURL(t *testing.T) {
	cases := map[string]string{
		"@CTVNews": "https://www.youtube.com/@CTVNews/videos",
		"CTVNews":  "https://www.youtube.com/@CTVNews/videos",
		"https://www.youtube.com/@CTVNews": "https://www.youtube.com/@CTVNews/videos",
		"UCi7Zk9baY1tvdlgxIML8MXg": "https://www.youtube.com/channel/UCi7Zk9baY1tvdlgxIML8MXg/videos",
	}
	for in, want := range cases {
		got := NormalizeChannelURL(in)
		if got != want {
			t.Errorf("NormalizeChannelURL(%q)=%q want %q", in, got, want)
		}
	}
}
