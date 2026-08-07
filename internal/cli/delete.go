package cli

import (
	"fmt"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <channel>",
		Short: "Delete a channel and its videos from the catalogue",
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
			if ch == nil {
				return fmt.Errorf("channel not found: %s", args[0])
			}
			if !force && formatFlag() != "json" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Delete channel %s (%s) and %d videos? Pass --force to confirm.\n", ch.ID, ch.Title, ch.VideoCount)
				return fmt.Errorf("refusing to delete without --force")
			}
			if err := st.DeleteChannel(ctx, ch.ID); err != nil {
				return err
			}
			return output.New(formatFlag()).WriteValue(map[string]any{
				"deleted": ch.ID,
				"title":   ch.Title,
				"status":  "ok",
			})
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")
	return cmd
}
