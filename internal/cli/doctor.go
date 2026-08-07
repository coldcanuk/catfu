package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/coldcanuk/catfu/internal/catalogue"
	"github.com/coldcanuk/catfu/internal/output"
	"github.com/coldcanuk/catfu/pkg/version"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check dependencies, DB accessibility, and credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()

			report := map[string]any{
				"version": version.Version,
				"go":      runtime.Version(),
				"os":      runtime.GOOS + "/" + runtime.GOARCH,
			}

			// yt-dlp
			yt := &catalogue.YTDLP{Bin: app.Config.YTDLP}
			ytPath, ytErr := yt.LookPath()
			ytInfo := map[string]any{"path": ytPath, "ok": ytErr == nil}
			if ytErr != nil {
				ytInfo["error"] = ytErr.Error()
			} else if ver, err := yt.Version(ctx); err == nil {
				ytInfo["version"] = ver
			} else {
				ytInfo["version_error"] = err.Error()
			}
			report["ytdlp"] = ytInfo

			// DB
			dbInfo := map[string]any{"path": app.Config.DBPath}
			st, err := openStore(ctx)
			if err != nil {
				dbInfo["ok"] = false
				dbInfo["error"] = err.Error()
			} else {
				defer st.Close()
				stats, err := st.Stats(ctx)
				if err != nil {
					dbInfo["ok"] = false
					dbInfo["error"] = err.Error()
				} else {
					dbInfo["ok"] = true
					dbInfo["channels"] = stats.Channels
					dbInfo["videos"] = stats.Videos
				}
			}
			report["database"] = dbInfo

			// Brave Search plan token (single product key for web/news/video)
			braveOK := app.Config.BraveAPIKey != ""
			report["brave"] = map[string]any{
				"api_key_set":   braveOK,
				"plan":          "Search",
				"header":        "X-Subscription-Token",
				"endpoints":     []string{"web", "news", "video"},
				"env":           []string{"BRAVE_API_KEY", "CATFU_BRAVE_API_KEY"},
				"note":          "Use a Search plan subscription token, not Answers",
			}

			// write perms on data dir
			dir := app.Config.DBPath
			if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
				report["db_file_exists"] = true
			}

			// PATH
			if p, err := exec.LookPath(app.Config.YTDLP); err == nil {
				report["ytdlp_lookpath"] = p
			}

			if formatFlag() == "table" {
				fmt.Fprintf(cmd.OutOrStdout(), "catfu doctor\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  version:  %s\n", version.Version)
				fmt.Fprintf(cmd.OutOrStdout(), "  go:       %s\n", runtime.Version())
				fmt.Fprintf(cmd.OutOrStdout(), "  yt-dlp:   %v\n", ytInfo)
				fmt.Fprintf(cmd.OutOrStdout(), "  database: %v\n", dbInfo)
				fmt.Fprintf(cmd.OutOrStdout(), "  brave:    set=%v (Search plan key)\n", braveOK)
				return nil
			}
			return output.New(formatFlag()).WriteValue(report)
		},
	}
}
