package cli

import (
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/coldcanuk/catfu/pkg/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print catfu version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.New(formatFlag()).WriteValue(map[string]string{
				"version": version.Version,
				"name":    "catfu",
			})
		},
	}
}
