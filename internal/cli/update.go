package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/catalogue"
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var full bool
	var limit int

	cmd := &cobra.Command{
		Use:   "update <channel>",
		Short: "Incrementally refresh a catalogued channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			ch, err := st.GetChannel(ctx, args[0])
			if err != nil {
				return err
			}
			target := args[0]
			dateAfter := ""
			if ch != nil {
				if ch.URL != "" {
					target = ch.URL
				} else if ch.CustomURL != "" {
					target = ch.CustomURL
				} else {
					target = ch.ID
				}
				if full {
					if d, err := st.NewestUploadDate(ctx, ch.ID); err == nil && d != "" {
						dateAfter = strings.ReplaceAll(d, "-", "")
					}
				}
			}
			yt := &catalogue.YTDLP{Bin: app.Config.YTDLP, Logger: app.Logger}
			svc := &catalogue.Service{Store: st, YTDLP: yt, Logger: app.Logger}
			opts := backends.CatalogueOpts{
				FullMetadata: full,
				Limit:        limit,
				SleepRequest: app.Config.SleepReq,
				SleepMin:     app.Config.SleepMin,
				SleepMax:     app.Config.SleepMax,
				DateAfter:    dateAfter,
			}
			if !app.Config.Quiet && formatFlag() != "json" {
				opts.Progress = func(n int, id, title string) {
					fmt.Fprintf(os.Stderr, "\r[%d] %s", n, id)
				}
			}
			id, count, err := svc.CatalogueChannelWithProgress(ctx, target, opts)
			if opts.Progress != nil {
				fmt.Fprintln(os.Stderr)
			}
			if err != nil {
				return err
			}
			return output.New(formatFlag()).WriteValue(map[string]any{
				"channel_id": id,
				"videos":     count,
				"status":     "updated",
				"dateafter":  dateAfter,
			})
		},
	}
	cmd.Flags().BoolVar(&full, "full", false, "full metadata refresh with optional dateafter")
	cmd.Flags().IntVar(&limit, "limit", 0, "max videos (0 = all)")
	return cmd
}
