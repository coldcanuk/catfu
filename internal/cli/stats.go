package cli

import (
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show high-level database statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			stats, err := st.Stats(ctx)
			if err != nil {
				return err
			}
			return output.New(formatFlag()).WriteValue(stats)
		},
	}
}
