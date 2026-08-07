package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/coldcanuk/catfu/internal/secrets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage stored Brave Search credentials (OS keychain preferred)",
		Long: `Store the Brave Search plan subscription token in the OS keychain
(macOS Keychain, Windows Credential Manager, Linux Secret Service / libsecret).

If no keychain is available (common on headless servers), catfu falls back to
a mode-0600 file at $XDG_CONFIG_HOME/catfu/secrets.

Precedence when running commands:
  1. --brave-api-key flag
  2. BRAVE_API_KEY / CATFU_BRAVE_API_KEY environment
  3. config.yaml brave_api_key
  4. OS keychain / secrets file (set via catfu auth set)

Agents/CI can still use env or flags; interactive users should prefer auth set.
`,
	}
	cmd.AddCommand(newAuthSetCmd(), newAuthStatusCmd(), newAuthClearCmd())
	return cmd
}

func newAuthSetCmd() *cobra.Command {
	var token string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Save Brave Search plan token to keychain (or secure file)",
		RunE: func(cmd *cobra.Command, args []string) error {
			tok := strings.TrimSpace(token)
			if tok == "" {
				// Non-interactive pipe support
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					tok = strings.TrimSpace(string(b))
				} else {
					fmt.Fprint(cmd.ErrOrStderr(), "Brave Search plan token (input hidden): ")
					b, err := term.ReadPassword(int(os.Stdin.Fd()))
					fmt.Fprintln(cmd.ErrOrStderr())
					if err != nil {
						return fmt.Errorf("read token: %w", err)
					}
					tok = strings.TrimSpace(string(b))
				}
			}
			if tok == "" {
				return fmt.Errorf("empty token")
			}
			backend, err := secrets.SetBraveAPIKey(tok)
			if err != nil {
				return err
			}
			out := map[string]any{
				"status":  "saved",
				"backend": string(backend),
				"plan":    "Search",
			}
			if backend == secrets.BackendFile {
				out["note"] = "OS keychain unavailable; stored in 0600 secrets file under config dir"
			} else {
				out["note"] = "stored in OS keychain (service=catfu)"
			}
			return output.New(formatFlag()).WriteValue(out)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "token value (prefer omit and type interactively)")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether a Brave token is stored (never prints the secret)",
		RunE: func(cmd *cobra.Command, args []string) error {
			st := secrets.Status()
			// Effective resolution (env/flag may still override)
			st["effective_key_set"] = app.Config.BraveAPIKey != ""
			st["effective_source"] = app.Config.BraveAPIKeySource
			return output.New(formatFlag()).WriteValue(st)
		},
	}
}

func newAuthClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove stored Brave token from keychain and secrets file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := secrets.DeleteBraveAPIKey(); err != nil {
				return err
			}
			return output.New(formatFlag()).WriteValue(map[string]any{
				"status": "cleared",
			})
		},
	}
}
