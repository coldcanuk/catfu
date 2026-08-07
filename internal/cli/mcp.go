package cli

import (
	"github.com/coldcanuk/catfu/internal/mcp"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run the catfu MCP server on stdio",
		Long:  "Starts a Model Context Protocol server on stdin/stdout. Logs go to stderr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return mcp.Run(ctx, mcp.Options{
				DBPath:      app.Config.DBPath,
				YTDLP:       app.Config.YTDLP,
				BraveAPIKey: app.Config.BraveAPIKey,
				SleepReq:    app.Config.SleepReq,
				SleepMin:    app.Config.SleepMin,
				SleepMax:    app.Config.SleepMax,
				Logger:      app.Logger,
			})
		},
	}
}
