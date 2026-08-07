package cli

import (
	"fmt"
	"strings"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/output"
	localsearch "github.com/coldcanuk/catfu/internal/search"
	"github.com/spf13/cobra"
)

func newWebCmd() *cobra.Command {
	var limit, offset int
	var after, before, kind, country, searchLang, safeSearch, freshness string

	cmd := &cobra.Command{
		Use:     "web <query>",
		Aliases: []string{"web-search"},
		Short:   "Search via Brave Search API (Search plan BYOK: web, news, video)",
		Long: `Search the public web (or news/video indexes) using your Brave Search
subscription token (BRAVE_API_KEY / --brave-api-key).

Requires a Brave **Search** plan key (not Answers). Free monthly credits apply
on that plan. See README for signup notes.

Kinds:
  web    Web Search API (default; count max 20)
  news   News Search API (count max 50)
  video  Video Search API (count max 50)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client := &brave.Client{APIKey: app.Config.BraveAPIKey}
			k := backends.SearchKind(strings.ToLower(strings.TrimSpace(kind)))
			if k == "" {
				k = backends.SearchKindWeb
			}
			sq := backends.SearchQuery{
				Query:      args[0],
				Limit:      limit,
				Offset:     offset,
				Kind:       k,
				Country:    country,
				SearchLang: searchLang,
				SafeSearch: safeSearch,
				Freshness:  freshness,
			}
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
				headers := []string{"kind", "title", "url", "description", "age"}
				rows := make([][]string, 0, len(results))
				for _, r := range results {
					rows = append(rows, []string{r.Kind, r.Title, r.URL, r.Description, r.Age})
				}
				return output.New("table").WriteRows(headers, rows)
			}
			return output.New(formatFlag()).WriteValue(results)
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "web", "search vertical: web|news|video")
	cmd.Flags().IntVar(&limit, "limit", 10, "max results (web max 20; news/video max 50)")
	cmd.Flags().IntVar(&offset, "offset", 0, "result offset (max 9)")
	cmd.Flags().StringVar(&after, "after", "", "freshness start date (YYYY-MM-DD); ignored if --freshness set")
	cmd.Flags().StringVar(&before, "before", "", "freshness end date (YYYY-MM-DD); ignored if --freshness set")
	cmd.Flags().StringVar(&freshness, "freshness", "", "pd|pw|pm|py or YYYY-MM-DDtoYYYY-MM-DD")
	cmd.Flags().StringVar(&country, "country", "", "2-letter country code (e.g. CA, US)")
	cmd.Flags().StringVar(&searchLang, "search-lang", "", "content language code (e.g. en, fr)")
	cmd.Flags().StringVar(&safeSearch, "safesearch", "", "off|moderate|strict")
	return cmd
}
