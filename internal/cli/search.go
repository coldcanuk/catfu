package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/output"
	localsearch "github.com/coldcanuk/catfu/internal/search"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var channel string
	var after, before string
	var limit, offset int
	var withWeb bool

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search the local catalogue (FTS + date/channel filters)",
		Long: `Search catalogued YouTube metadata in the local SQLite FTS index.

Use --web to force-multiply: after local hits, query Brave video search for
YouTube results not yet in your catalogue (requires Brave Search plan key).

Channel filter accepts UC… id or @handle (resolved via the channels table).
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			q := ""
			if len(args) > 0 {
				q = args[0]
			}
			sq := backends.SearchQuery{Query: q, Limit: limit, Offset: offset}
			if after != "" {
				t, err := localsearch.ParseDate(after)
				if err != nil {
					return fmt.Errorf("invalid --after: %w", err)
				}
				sq.After = t
			}
			if before != "" {
				t, err := localsearch.ParseDate(before)
				if err != nil {
					return fmt.Errorf("invalid --before: %w", err)
				}
				sq.Before = t
			}
			if channel != "" {
				ch, err := st.GetChannel(ctx, channel)
				if err != nil {
					return err
				}
				if ch == nil {
					return fmt.Errorf("channel not in catalogue: %q (try catfu list; use UC id or @handle after catalogue)", channel)
				}
				sq.ChannelID = ch.ID
			}

			cs := &localsearch.CatalogueSearcher{Store: st}
			var results []backends.Result
			if withWeb {
				if app.Config.BraveAPIKey == "" {
					return fmt.Errorf("--web requires Brave API key (catfu auth set)")
				}
				h := &localsearch.Hybrid{
					Local: cs,
					Brave: &brave.Client{APIKey: app.Config.BraveAPIKey},
				}
				results, err = h.Search(ctx, sq)
			} else {
				results, err = cs.Search(ctx, sq)
			}
			if err != nil {
				return err
			}
			if formatFlag() == "table" {
				headers := []string{"source", "id", "title", "channel", "upload_date", "duration", "url"}
				rows := make([][]string, 0, len(results))
				for _, r := range results {
					rows = append(rows, []string{r.Source, r.ID, r.Title, r.Channel, r.UploadDate, fmt.Sprintf("%d", r.Duration), r.URL})
				}
				return output.New("table").WriteRows(headers, rows)
			}
			return output.New(formatFlag()).WriteValue(results)
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "filter by channel id or @handle (must already be catalogued)")
	cmd.Flags().StringVar(&after, "after", "", "only videos on/after date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&before, "before", "", "only videos on/before date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "result offset")
	cmd.Flags().BoolVar(&withWeb, "web", false, "also query Brave video for YouTube hits not in catalogue")
	return cmd
}
