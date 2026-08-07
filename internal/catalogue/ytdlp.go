package catalogue

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// YTDLP wraps the yt-dlp binary for metadata extraction.
type YTDLP struct {
	Bin    string
	Logger *slog.Logger
}

// ChannelMeta is resolved channel metadata from yt-dlp.
type ChannelMeta struct {
	ID           string
	Title        string
	Description  string
	CustomURL    string
	URL          string
	ThumbnailURL string
}

// RawEntry is a loosely typed yt-dlp JSON object.
type RawEntry map[string]any

// ResolveChannel fetches channel metadata (single JSON).
func (y *YTDLP) ResolveChannel(ctx context.Context, channelURL string) (*ChannelMeta, error) {
	args := []string{
		"--dump-single-json",
		"--skip-download",
		"--playlist-end", "1",
		"--no-warnings",
		channelURL,
	}
	cmd := exec.CommandContext(ctx, y.bin(), args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp resolve: %w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("yt-dlp resolve: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, fmt.Errorf("parse channel json: %w", err)
	}
	meta := &ChannelMeta{
		ID:          firstString(m, "channel_id", "id"),
		Title:       firstString(m, "channel", "title", "uploader"),
		Description: firstString(m, "description"),
		CustomURL:   firstString(m, "uploader_id", "channel"),
		URL:         firstString(m, "channel_url", "webpage_url", "original_url"),
	}
	// id might be @handle when channel_id present
	if cid := firstString(m, "channel_id"); cid != "" {
		meta.ID = cid
	}
	if cu := firstString(m, "uploader_id"); strings.HasPrefix(cu, "@") {
		meta.CustomURL = cu
	} else if id := firstString(m, "id"); strings.HasPrefix(id, "@") {
		meta.CustomURL = id
	}
	meta.ThumbnailURL = pickThumbnail(m)
	if meta.ID == "" {
		return nil, fmt.Errorf("could not resolve channel id from %s", channelURL)
	}
	if meta.Title == "" {
		meta.Title = meta.ID
	}
	return meta, nil
}

// DumpOpts controls playlist dumping.
type DumpOpts struct {
	FullMetadata bool
	Limit        int
	SleepRequest float64
	SleepMin     float64
	SleepMax     float64
	DateAfter    string // YYYYMMDD for non-flat
}

// StreamEntries runs yt-dlp --dump-json and invokes fn for each video entry.
func (y *YTDLP) StreamEntries(ctx context.Context, channelURL string, opts DumpOpts, fn func(RawEntry) error) (int, error) {
	args := []string{
		"--dump-json",
		"--skip-download",
		"--no-warnings",
	}
	if !opts.FullMetadata {
		args = append(args, "--flat-playlist")
	}
	if opts.Limit > 0 {
		args = append(args, "--playlist-end", strconv.Itoa(opts.Limit))
	}
	if opts.SleepRequest > 0 {
		args = append(args, "--sleep-requests", fmt.Sprintf("%g", opts.SleepRequest))
	}
	if opts.SleepMin > 0 {
		args = append(args, "--sleep-interval", fmt.Sprintf("%g", opts.SleepMin))
	}
	if opts.SleepMax > 0 {
		args = append(args, "--max-sleep-interval", fmt.Sprintf("%g", opts.SleepMax))
	}
	if opts.DateAfter != "" && opts.FullMetadata {
		args = append(args, "--dateafter", opts.DateAfter)
	}
	args = append(args, channelURL)

	cmd := exec.CommandContext(ctx, y.bin(), args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start yt-dlp: %w", err)
	}

	// drain stderr
	errCh := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(stderr)
		errCh <- string(b)
	}()

	sc := bufio.NewScanner(stdout)
	// large JSON lines
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 10*1024*1024)

	n := 0
	var scanErr error
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var entry RawEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			if y.Logger != nil {
				y.Logger.Warn("skip malformed yt-dlp line", "err", err)
			}
			continue
		}
		// skip non-video playlist wrappers if any
		if t, _ := entry["_type"].(string); t == "playlist" {
			continue
		}
		if err := fn(entry); err != nil {
			scanErr = err
			break
		}
		n++
	}
	if scanErr == nil {
		scanErr = sc.Err()
	}
	waitErr := cmd.Wait()
	stderrOut := <-errCh
	if scanErr != nil {
		return n, scanErr
	}
	if waitErr != nil && n == 0 {
		return n, fmt.Errorf("yt-dlp: %w: %s", waitErr, strings.TrimSpace(stderrOut))
	}
	if waitErr != nil && y.Logger != nil {
		y.Logger.Warn("yt-dlp exited non-zero after partial stream", "err", waitErr, "ingested", n)
	}
	return n, nil
}

// Version returns yt-dlp --version output.
func (y *YTDLP) Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, y.bin(), "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// LookPath checks whether the binary exists.
func (y *YTDLP) LookPath() (string, error) {
	return exec.LookPath(y.bin())
}

func (y *YTDLP) bin() string {
	if y != nil && y.Bin != "" {
		return y.Bin
	}
	return "yt-dlp"
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			case float64:
				return strconv.FormatInt(int64(t), 10)
			case json.Number:
				return t.String()
			}
		}
	}
	return ""
}

func pickThumbnail(m map[string]any) string {
	if t, ok := m["thumbnail"].(string); ok && t != "" {
		return t
	}
	thumbs, ok := m["thumbnails"].([]any)
	if !ok || len(thumbs) == 0 {
		return ""
	}
	// last is usually largest
	for i := len(thumbs) - 1; i >= 0; i-- {
		if tm, ok := thumbs[i].(map[string]any); ok {
			if u, ok := tm["url"].(string); ok && u != "" {
				return u
			}
		}
	}
	return ""
}

// MappedVideo is structured metadata extracted from one yt-dlp JSON entry.
type MappedVideo struct {
	ID              string
	Title           string
	Description     string
	UploadDate      string
	ThumbnailURL    string
	WebpageURL      string
	ChannelID       string
	ChannelTitle    string
	Duration        int
	ViewCount       *int64
	LikeCount       *int64
	CommentCount    *int64
	Language        string
	Languages       []string // caption/subtitle language codes
	HasSubtitles    bool
	HasAutoCaptions bool
	HasTranscript   bool
}

// MapEntry converts a yt-dlp entry into structured video metadata.
func MapEntry(entry RawEntry, fallbackChannelID string) MappedVideo {
	var m MappedVideo
	m.ID = firstString(entry, "id")
	m.Title = firstString(entry, "title")
	m.Description = firstString(entry, "description")
	m.UploadDate = normalizeUploadDate(firstString(entry, "upload_date", "release_date"))
	m.ThumbnailURL = pickThumbnail(entry)
	m.WebpageURL = firstString(entry, "webpage_url", "url")
	if m.WebpageURL == "" && m.ID != "" {
		m.WebpageURL = "https://www.youtube.com/watch?v=" + m.ID
	}
	m.ChannelID = firstString(entry, "channel_id", "playlist_channel_id")
	if m.ChannelID == "" {
		m.ChannelID = fallbackChannelID
	}
	m.ChannelTitle = firstString(entry, "channel", "playlist_channel", "playlist_title", "uploader")
	if d, ok := asInt(entry["duration"]); ok {
		m.Duration = d
	}
	if v, ok := asInt64(entry["view_count"]); ok {
		m.ViewCount = &v
	}
	if v, ok := asInt64(entry["like_count"]); ok {
		m.LikeCount = &v
	}
	if v, ok := asInt64(entry["comment_count"]); ok {
		m.CommentCount = &v
	}
	m.Language = firstString(entry, "language")
	// Manual subtitles (human) vs automatic captions
	manLangs := mapKeys(entry["subtitles"])
	autoLangs := mapKeys(entry["automatic_captions"])
	m.HasSubtitles = len(manLangs) > 0
	m.HasAutoCaptions = len(autoLangs) > 0
	m.HasTranscript = m.HasSubtitles || m.HasAutoCaptions
	// Union of language codes (normalize en-orig → en)
	seen := map[string]bool{}
	var langs []string
	for _, code := range append(append([]string{}, manLangs...), autoLangs...) {
		code = normalizeLangCode(code)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		langs = append(langs, code)
	}
	if m.Language != "" && !seen[m.Language] {
		langs = append(langs, m.Language)
	}
	sort.Strings(langs)
	m.Languages = langs
	return m
}

func mapKeys(v any) []string {
	m, ok := v.(map[string]any)
	if !ok || len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		if k != "" {
			out = append(out, k)
		}
	}
	return out
}

func normalizeLangCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	// en-orig, en-US, en-j3Py… → base language when possible
	if i := strings.IndexByte(code, '-'); i > 0 {
		base := code[:i]
		// keep region-style en-US as en-US if second part is 2 letters; drop garbage suffixes
		rest := code[i+1:]
		if len(rest) == 2 {
			return strings.ToLower(base) + "-" + strings.ToUpper(rest)
		}
		return strings.ToLower(base)
	}
	return strings.ToLower(code)
}

func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	case json.Number:
		i, err := t.Int64()
		return int(i), err == nil
	case string:
		i, err := strconv.Atoi(t)
		return i, err == nil
	default:
		return 0, false
	}
}

func asInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case json.Number:
		i, err := t.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}

func normalizeUploadDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 8 && !strings.Contains(s, "-") {
		return s[0:4] + "-" + s[4:6] + "-" + s[6:8]
	}
	return s
}
