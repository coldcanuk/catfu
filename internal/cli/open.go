package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/coldcanuk/catfu/internal/output"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <video-id>",
		Short: "Open a video in the default browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			url := "https://www.youtube.com/watch?v=" + args[0]
			st, err := openStore(ctx)
			if err == nil {
				defer st.Close()
				if v, _ := st.GetVideo(ctx, args[0]); v != nil && v.WebpageURL != "" {
					url = v.WebpageURL
				}
			}
			if err := openBrowser(url); err != nil {
				return err
			}
			return output.New(formatFlag()).WriteValue(map[string]string{"opened": url})
		},
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported OS for open: %s", runtime.GOOS)
	}
	return cmd.Start()
}
