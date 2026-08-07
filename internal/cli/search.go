package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/output"
	localsearch "github.com/coldcanuk/catfu/internal/search"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var channel string
	var after, before string
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search the local catalogue (FTS + date/channel filters)",
		Args:  cobra.MaximumNArgs(1),
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
			sq := backends.SearchQuery{Query: q, ChannelID: channel, Limit: limit, Offset: offset}
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
			// If channel looks like handle, resolve to id
			if channel != "" {
				if ch, _ := st.GetChannel(ctx, channel); ch != nil {
					sq.ChannelID = ch.ID
				}
			}
			cs := &localsearch.CatalogueSearcher{Store: st}
			results, err := cs.Search(ctx, sq)
			if err != nil {
				return err
			}
			if formatFlag() == "table" {
				headers := []string{"id", "title", "channel", "upload_date", "duration", "url"}
				rows := make([][]string, 0, len(results))
				for _, r := range results {
					rows = append(rows, []string{r.ID, r.Title, r.Channel, r.UploadDate, fmt.Sprintf("%d", r.Duration), r.URL})
				}
				return output.New("table").WriteRows(headers, rows)
			}
			return output.New(formatFlag()).WriteValue(results)
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "filter by channel id or @handle")
	cmd.Flags().StringVar(&after, "after", "", "only videos on/after date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&before, "before", "", "only videos on/before date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "result offset")
	return cmd
}
