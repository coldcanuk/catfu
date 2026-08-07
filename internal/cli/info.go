package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <channel|video-id>",
		Short: "Show details for a channel or video",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			st, err := openStore(ctx)
			if err != nil {
				return err
			}
			defer st.Close()
			id := args[0]
			if v, err := st.GetVideo(ctx, id); err != nil {
				return err
			} else if v != nil {
				return output.New(formatFlag()).WriteValue(v)
			}
			if ch, err := st.GetChannel(ctx, id); err != nil {
				return err
			} else if ch != nil {
				return output.New(formatFlag()).WriteValue(ch)
			}
			return fmt.Errorf("not found: %s", id)
		},
	}
}
