package cli

import (
	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/output"
	localsearch "github.com/coldcanuk/catfu/internal/search"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var channel, query, after, before string
	var limit, offset int
	var all bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export search results or full catalogue as JSON/CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			if all {
				limit = 1000000
			}
			if limit <= 0 {
				limit = 1000
			}
			sq := backends.SearchQuery{Query: query, Limit: limit, Offset: offset}
			if channel != "" {
				if ch, _ := st.GetChannel(ctx, channel); ch != nil {
					sq.ChannelID = ch.ID
				} else {
					sq.ChannelID = channel
				}
			}
			if after != "" {
				t, err := localsearch.ParseDate(after)
				if err != nil {
					return err
				}
				sq.After = t
			}
			if before != "" {
				t, err := localsearch.ParseDate(before)
				if err != nil {
					return err
				}
				sq.Before = t
			}
			cs := &localsearch.CatalogueSearcher{Store: st}
			results, err := cs.Search(ctx, sq)
			if err != nil {
				return err
			}
			fmtStr := formatFlag()
			if fmtStr == "table" {
				fmtStr = "json"
			}
			return output.New(fmtStr).WriteValue(results)
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "filter channel")
	cmd.Flags().StringVar(&query, "query", "", "FTS query")
	cmd.Flags().StringVar(&after, "after", "", "date after")
	cmd.Flags().StringVar(&before, "before", "", "date before")
	cmd.Flags().IntVar(&limit, "limit", 1000, "max rows")
	cmd.Flags().IntVar(&offset, "offset", 0, "offset")
	cmd.Flags().BoolVar(&all, "all", false, "export all matching (large limit)")
	return cmd
}
