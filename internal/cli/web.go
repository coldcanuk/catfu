package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/output"
	localsearch "github.com/coldcanuk/catfu/internal/search"
	"github.com/spf13/cobra"
)

func newWebCmd() *cobra.Command {
	var limit, offset int
	var after, before string

	cmd := &cobra.Command{
		Use:     "web <query>",
		Aliases: []string{"web-search"},
		Short:   "Search the web via Brave Search API (BYOK)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := &brave.Client{APIKey: app.Config.BraveAPIKey}
			sq := backends.SearchQuery{Query: args[0], Limit: limit, Offset: offset}
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
			results, err := client.Search(ctx, sq)
			if err != nil {
				return err
			}
			if formatFlag() == "table" {
				headers := []string{"title", "url", "description"}
				rows := make([][]string, 0, len(results))
				for _, r := range results {
					rows = append(rows, []string{r.Title, r.URL, r.Description})
				}
				return output.New("table").WriteRows(headers, rows)
			}
			return output.New(formatFlag()).WriteValue(results)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "max results (max 20)")
	cmd.Flags().IntVar(&offset, "offset", 0, "result offset (max 9)")
	cmd.Flags().StringVar(&after, "after", "", "freshness start date")
	cmd.Flags().StringVar(&before, "before", "", "freshness end date")
	return cmd
}
