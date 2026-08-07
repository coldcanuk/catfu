package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"catalogues"},
		Short:   "List catalogued channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			chs, err := st.ListChannels(ctx)
			if err != nil {
				return err
			}
			if formatFlag() == "table" {
				headers := []string{"id", "title", "custom_url", "videos", "last_catalogued"}
				rows := make([][]string, 0, len(chs))
				for _, c := range chs {
					rows = append(rows, []string{c.ID, c.Title, c.CustomURL, fmt.Sprintf("%d", c.VideoCount), c.LastCatalogued})
				}
				return output.New("table").WriteRows(headers, rows)
			}
			return output.New(formatFlag()).WriteValue(chs)
		},
	}
	return cmd
}
