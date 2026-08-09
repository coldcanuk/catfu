package cli

import (
	"fmt"
	"strings"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/coldcanuk/catfu/internal/social"
	"github.com/coldcanuk/catfu/internal/store"
	"github.com/spf13/cobra"
)

// SocialsResult is the CLI JSON envelope for catfu socials.
type SocialsResult struct {
	ChannelID      string        `json:"channel_id,omitempty"`
	CustomURL      string        `json:"custom_url,omitempty"`
	Title          string        `json:"title,omitempty"`
	VideoID        string        `json:"video_id,omitempty"`
	ScannedVideos  int           `json:"scanned_videos"`
	ScannedSources int           `json:"scanned_sources"`
	Links          []social.Link `json:"links"`
}

func newSocialsCmd() *cobra.Command {
	var (
		videoID  string
		source   string
		platform []string
		unique   bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "socials [channel]",
		Short: "Extract social media links/handles from channel and video descriptions",
		Long: `Scan catalogued channel about text and video descriptions for social profiles
using regular expressions (X, Threads, Facebook, Instagram, Bluesky, Nostr,
TikTok, LinkedIn, Mastodon, Discord, GitHub, GitLab, Telegram, WeChat,
WhatsApp, LINE).

Descriptions are often empty unless the channel was catalogued with --full:

  catfu catalogue @SomeChannel --full --limit 100
  catfu socials @SomeChannel --json

Examples:
  catfu socials @SomeChannel
  catfu socials @SomeChannel --json
  catfu socials --video dQw4w9WgXcQ --json
  catfu socials UC… --source channel --platform x,instagram
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()

			platforms := expandPlatformFlags(platform)
			src := strings.ToLower(strings.TrimSpace(source))
			if src == "" {
				src = "all"
			}
			switch src {
			case "all", "channel", "videos":
			default:
				return fmt.Errorf("invalid --source %q (want all|channel|videos)", source)
			}

			var (
				blobs   []social.SourceBlob
				result  SocialsResult
				channel *store.Channel
			)

			if videoID != "" {
				v, err := st.GetVideo(ctx, videoID)
				if err != nil {
					return err
				}
				if v == nil {
					return fmt.Errorf("video not found: %s", videoID)
				}
				result.VideoID = v.ID
				result.ChannelID = v.ChannelID
				if ch, _ := st.GetChannel(ctx, v.ChannelID); ch != nil {
					channel = ch
					result.CustomURL = ch.CustomURL
					result.Title = ch.Title
				}
				if src == "all" || src == "videos" {
					blobs = append(blobs, social.SourceBlob{
						Text: v.Description, Source: "video", VideoID: v.ID, ChannelID: v.ChannelID,
					})
					result.ScannedVideos = 1
				}
			} else {
				if len(args) == 0 {
					return fmt.Errorf("channel argument required (or pass --video)")
				}
				ch, err := st.GetChannel(ctx, args[0])
				if err != nil {
					return err
				}
				if ch == nil {
					return fmt.Errorf("channel not found: %s", args[0])
				}
				channel = ch
				result.ChannelID = ch.ID
				result.CustomURL = ch.CustomURL
				result.Title = ch.Title

				if src == "all" || src == "channel" {
					blobs = append(blobs, social.SourceBlob{
						Text: ch.Description, Source: "channel", ChannelID: ch.ID,
					})
				}
				if src == "all" || src == "videos" {
					vids, err := st.ListVideosByChannel(ctx, ch.ID, limit)
					if err != nil {
						return err
					}
					result.ScannedVideos = len(vids)
					for _, v := range vids {
						blobs = append(blobs, social.SourceBlob{
							Text: v.Description, Source: "video", VideoID: v.ID, ChannelID: v.ChannelID,
						})
					}
				}
			}

			_ = channel
			result.ScannedSources = len(blobs)
			links := social.ExtractSources(blobs)
			links = social.FilterPlatforms(links, platforms)
			if unique {
				links = social.Dedup(links)
			}
			social.SortLinks(links)
			if links == nil {
				links = []social.Link{}
			}
			result.Links = links

			if formatFlag() == "table" {
				headers := []string{"platform", "handle", "url", "confidence", "source", "video_id"}
				rows := make([][]string, 0, len(links))
				for _, l := range links {
					rows = append(rows, []string{
						string(l.Platform),
						l.Handle,
						l.URL,
						string(l.Confidence),
						l.Source,
						l.VideoID,
					})
				}
				if len(rows) == 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "no social links found (scanned_sources=%d scanned_videos=%d; catalogue with --full for descriptions)\n",
						result.ScannedSources, result.ScannedVideos)
				}
				return output.New("table").WriteRows(headers, rows)
			}
			return output.New(formatFlag()).WriteValue(result)
		},
	}

	cmd.Flags().StringVar(&videoID, "video", "", "scan a single video id instead of a channel")
	cmd.Flags().StringVar(&source, "source", "all", "text source: all|channel|videos")
	cmd.Flags().StringSliceVar(&platform, "platform", nil, "filter platforms (repeatable or comma-separated): x,instagram,...")
	cmd.Flags().BoolVar(&unique, "unique", true, "dedupe by platform+handle across sources")
	cmd.Flags().IntVar(&limit, "limit", 10000, "max videos to scan for a channel")
	return cmd
}

func expandPlatformFlags(flags []string) []string {
	var out []string
	for _, f := range flags {
		for _, p := range strings.Split(f, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}
