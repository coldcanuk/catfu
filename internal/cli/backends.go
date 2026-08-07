package cli

import (
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newBackendsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backends",
		Short: "List available search/catalogue backends and status",
		RunE: func(cmd *cobra.Command, args []string) error {
			backends := []map[string]any{
				{
					"name":        "catalogue",
					"type":        "local",
					"description": "SQLite FTS5 local catalogue",
					"configured":  true,
					"db":          app.Config.DBPath,
				},
				{
					"name":        "brave",
					"type":        "web",
					"description": "Brave Search plan (web, news, video) via X-Subscription-Token",
					"configured":  app.Config.BraveAPIKey != "",
					"plan":        "Search (not Answers)",
					"env":         "BRAVE_API_KEY",
					"kinds":       []string{"web", "news", "video"},
				},
				{
					"name":        "yt-dlp",
					"type":        "catalogue",
					"description": "YouTube metadata via external yt-dlp",
					"binary":      app.Config.YTDLP,
				},
			}
			return output.New(formatFlag()).WriteValue(backends)
		},
	}
}
