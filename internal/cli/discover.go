package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/catalogue"
	"github.com/coldcanuk/catfu/internal/discover"
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newDiscoverCmd() *cobra.Command {
	var limit int
	var country, kind string
	var doCatalogue bool
	var catLimit int
	var full bool

	cmd := &cobra.Command{
		Use:   "discover <topic>",
		Short: "Use Brave to find YouTube channels/videos to catalogue (force multiplier)",
		Long: `discover is the bridge between Brave Search and the local YouTube catalogue.

It searches Brave (video index + site:youtube.com) for public YouTube signals,
extracts channel handles / UC ids / video ids, and marks which are already in
your local SQLite catalogue.

Typical flow:
  catfu discover "golang concurrency" --json
  catfu catalogue @somechannel --limit 100
  catfu search "concurrency" --json

Optional: --catalogue auto-ingests up to 5 new suggested channels (requires yt-dlp).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if app.Config.BraveAPIKey == "" {
				return fmt.Errorf("brave API key required for discover (catfu auth set, or BRAVE_API_KEY)")
			}
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()

			svc := &discover.Service{
				Brave: &brave.Client{APIKey: app.Config.BraveAPIKey},
				Store: st,
			}
			rep, err := svc.Discover(ctx, args[0], discover.Options{
				Limit:   limit,
				Country: country,
				Kind:    kind,
			})
			if err != nil {
				return err
			}

			if doCatalogue {
				yt := &catalogue.YTDLP{Bin: app.Config.YTDLP, Logger: app.Logger}
				if _, err := yt.LookPath(); err != nil {
					return fmt.Errorf("yt-dlp required for --catalogue: %w", err)
				}
				catSvc := &catalogue.Service{Store: st, YTDLP: yt, Logger: app.Logger}
				type catResult struct {
					Target string `json:"target"`
					ID     string `json:"channel_id,omitempty"`
					Videos int    `json:"videos"`
					Error  string `json:"error,omitempty"`
				}
				var cats []catResult
				seen := map[string]bool{}
				for _, ch := range rep.Channels {
					if ch.Catalogued {
						continue
					}
					target := ch.URL
					if target == "" {
						target = ch.Handle
					}
					if target == "" || seen[target] {
						continue
					}
					seen[target] = true
					if len(cats) >= 5 {
						break
					}
					id, n, err := catSvc.CatalogueChannelWithProgress(ctx, target, backends.CatalogueOpts{
						FullMetadata: full,
						Limit:        catLimit,
						SleepRequest: app.Config.SleepReq,
						SleepMin:     app.Config.SleepMin,
						SleepMax:     app.Config.SleepMax,
					})
					cr := catResult{Target: target, ID: id, Videos: n}
					if err != nil {
						cr.Error = err.Error()
					}
					cats = append(cats, cr)
				}
				return output.New(formatFlag()).WriteValue(map[string]any{
					"discover":   rep,
					"catalogued": cats,
				})
			}

			if formatFlag() == "table" {
				fmt.Fprintf(cmd.OutOrStdout(), "discover: %s\n\n", rep.Query)
				fmt.Fprintf(cmd.OutOrStdout(), "Channels to catalogue (%d):\n", len(rep.Channels))
				for _, ch := range rep.Channels {
					status := "new"
					if ch.Catalogued {
						status = fmt.Sprintf("have %d videos", ch.LocalVideos)
					}
					label := ch.Handle
					if label == "" {
						label = ch.ChannelID
					}
					if label == "" {
						label = ch.URL
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n    %s\n", status, label, ch.URL)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nYouTube videos found (%d):\n", len(rep.Videos))
				for _, v := range rep.Videos {
					tag := "remote"
					if v.InCatalogue {
						tag = "local"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n    %s\n", tag, v.Title, v.URL)
				}
				if rep.Note != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", rep.Note)
				}
				return nil
			}
			return output.New(formatFlag()).WriteValue(rep)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 15, "Brave result budget per vertical")
	cmd.Flags().StringVar(&country, "country", "", "2-letter country code")
	cmd.Flags().StringVar(&kind, "kind", "video", "brave vertical: video|web|both")
	cmd.Flags().BoolVar(&doCatalogue, "catalogue", false, "auto-catalogue up to 5 new suggested channels")
	cmd.Flags().IntVar(&catLimit, "catalogue-limit", 50, "max videos per channel when --catalogue")
	cmd.Flags().BoolVar(&full, "full", false, "full metadata when --catalogue")
	return cmd
}
