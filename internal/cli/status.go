package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [channel]",
		Short: "Show database or channel catalogue status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			if len(args) == 0 {
				stats, err := st.Stats(ctx)
				if err != nil {
					return err
				}
				return output.New(formatFlag()).WriteValue(stats)
			}
			ch, err := st.GetChannel(ctx, args[0])
			if err != nil {
				return err
			}
			if ch == nil {
				return fmt.Errorf("channel not found: %s", args[0])
			}
			return output.New(formatFlag()).WriteValue(ch)
		},
	}
}
