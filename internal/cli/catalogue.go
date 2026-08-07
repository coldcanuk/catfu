package cli

import (
	"fmt"
	"os"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/catalogue"
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newCatalogueCmd() *cobra.Command {
	var full bool
	var limit int

	cmd := &cobra.Command{
		Use:   "catalogue <channel>",
		Short: "Catalogue a YouTube channel's video metadata via yt-dlp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()

			yt := &catalogue.YTDLP{Bin: app.Config.YTDLP, Logger: app.Logger}
			if _, err := yt.LookPath(); err != nil {
				return fmt.Errorf("yt-dlp not found on PATH (%s): install yt-dlp or set --ytdlp", app.Config.YTDLP)
			}
			svc := &catalogue.Service{Store: st, YTDLP: yt, Logger: app.Logger}
			opts := backends.CatalogueOpts{
				FullMetadata: full,
				Limit:        limit,
				SleepRequest: app.Config.SleepReq,
				SleepMin:     app.Config.SleepMin,
				SleepMax:     app.Config.SleepMax,
			}
			if !app.Config.Quiet && formatFlag() != "json" {
				opts.Progress = func(n int, id, title string) {
					fmt.Fprintf(os.Stderr, "\r[%d] %s %s", n, id, truncate(title, 60))
				}
			}
			id, count, err := svc.CatalogueChannelWithProgress(ctx, args[0], opts)
			if opts.Progress != nil {
				fmt.Fprintln(os.Stderr)
			}
			if err != nil {
				return err
			}
			return output.New(formatFlag()).WriteValue(map[string]any{
				"channel_id": id,
				"videos":     count,
				"status":     "ok",
			})
		},
	}
	cmd.Flags().BoolVar(&full, "full", false, "full metadata: views/likes/comments, languages, captions/transcript flags, descriptions (slower)")
	cmd.Flags().IntVar(&limit, "limit", 0, "max videos to catalogue (0 = all)")
	return cmd
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
