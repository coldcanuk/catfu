package cli

import (
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show effective configuration (secrets redacted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.New(formatFlag()).WriteValue(app.Config.Redacted())
		},
	}
}
