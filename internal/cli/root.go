// Package cli implements the catfu Cobra command tree.
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/coldcanuk/catfu/internal/config"
	"github.com/coldcanuk/catfu/internal/store"
	"github.com/coldcanuk/catfu/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Dependencies shared across commands.
type App struct {
	Viper  *viper.Viper
	Config config.Config
	Logger *slog.Logger
	Store  *store.Store
}

var app = &App{}

// NewRoot builds the root command.
func NewRoot() *cobra.Command {
	var cfgFile string

	root := &cobra.Command{
		Use:   "catfu",
		Short: "YouTube channel metadata catalogue with local FTS and pluggable search",
		Long: `catfu catalogues YouTube channel metadata (no video download) into a local
SQLite database with FTS5 search, optional Brave web search, and an MCP server.

yt-dlp must be installed separately and available on PATH.

Examples:
  catfu doctor --json
  catfu catalogue @SomeChannel --limit 50
  export BRAVE_API_KEY='your-search-plan-token'
  catfu web "query" --country CA --json
  catfu web "query" --brave-api-key "$BRAVE_API_KEY" --kind news

Note: global flags alone do nothing — always pass a subcommand (web, doctor, …).`,
		SilenceUsage:  true,
		SilenceErrors: true,
		// Running `catfu --brave-api-key …` with no subcommand used to only dump help.
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("missing command: try `catfu doctor`, `catfu web \"query\"`, or `catfu --help`\n  Brave key example: catfu web \"golang\" --brave-api-key \"$BRAVE_API_KEY\" --json")
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			v, err := config.InitViper(cfgFile)
			if err != nil {
				return err
			}
			// Bind flags that were registered
			_ = v.BindPFlags(cmd.Flags())
			_ = v.BindPFlags(cmd.Root().PersistentFlags())
			// Re-bind after flag parse: merge persistent flags into viper
			if f := cmd.Flags().Lookup("db"); f != nil && f.Changed {
				v.Set("db", f.Value.String())
			}
			if f := cmd.Root().PersistentFlags().Lookup("db"); f != nil && f.Changed {
				v.Set("db", f.Value.String())
			}
			for _, key := range []string{"json", "quiet", "format", "brave-api-key", "ytdlp", "log-level", "sleep-requests", "sleep-interval", "max-sleep-interval"} {
				if f := cmd.Root().PersistentFlags().Lookup(key); f != nil && f.Changed {
					viperKey := strings.ReplaceAll(key, "-", "_")
					if key == "brave-api-key" {
						viperKey = "brave_api_key"
					}
					v.Set(viperKey, f.Value.String())
					if key == "json" || key == "quiet" {
						b, _ := cmd.Root().PersistentFlags().GetBool(key)
						v.Set(viperKey, b)
					}
					if key == "sleep-requests" || key == "sleep-interval" || key == "max-sleep-interval" {
						fv, _ := cmd.Root().PersistentFlags().GetFloat64(key)
						v.Set(viperKey, fv)
					}
				}
			}
			cfg, err := config.Load(v)
			if err != nil {
				return err
			}
			cfg.ConfigFile = cfgFile
			if f := cmd.Root().PersistentFlags().Lookup("brave-api-key"); f != nil && f.Changed {
				cfg.BraveAPIKey = f.Value.String()
				cfg.BraveAPIKeySource = "flag"
			}
			app.Viper = v
			app.Config = cfg
			app.Logger = newLogger(cfg)
			return nil
		},
	}

	pf := root.PersistentFlags()
	pf.StringVar(&cfgFile, "config", "", "config file (default: $XDG_CONFIG_HOME/catfu/config.yaml)")
	pf.String("db", "", "SQLite database path (env CATFU_DB)")
	pf.Bool("json", false, "emit JSON on stdout (agent-friendly)")
	pf.String("format", "table", "output format: table|json|csv")
	pf.BoolP("quiet", "q", false, "reduce progress noise")
	pf.String("brave-api-key", "", "Brave Search plan token (env BRAVE_API_KEY; not Answers)")
	pf.String("ytdlp", "yt-dlp", "yt-dlp binary name or path")
	pf.String("log-level", "info", "log level: debug|info|warn|error")
	pf.Float64("sleep-requests", 0.5, "yt-dlp --sleep-requests seconds")
	pf.Float64("sleep-interval", 1, "yt-dlp --sleep-interval seconds")
	pf.Float64("max-sleep-interval", 3, "yt-dlp --max-sleep-interval seconds")

	root.AddCommand(
		newCatalogueCmd(),
		newSearchCmd(),
		newWebCmd(),
		newListCmd(),
		newStatusCmd(),
		newUpdateCmd(),
		newInfoCmd(),
		newDeleteCmd(),
		newExportCmd(),
		newVersionCmd(),
		newDoctorCmd(),
		newConfigCmd(),
		newAuthCmd(),
		newCompletionCmd(),
		newOpenCmd(),
		newStatsCmd(),
		newBackendsCmd(),
		newMCPCmd(),
	)

	root.Version = version.Version
	return root
}

// Execute runs the CLI.
func Execute() {
	root := NewRoot()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newLogger(cfg config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(h)
}

func openStore(ctx context.Context) (*store.Store, error) {
	if err := config.EnsureDBDir(app.Config.DBPath); err != nil {
		return nil, err
	}
	s, err := store.Open(ctx, app.Config.DBPath)
	if err != nil {
		return nil, err
	}
	app.Store = s
	return s, nil
}

func formatFlag() string {
	if app.Config.JSON || app.Config.Format == "json" {
		return "json"
	}
	return app.Config.Format
}
