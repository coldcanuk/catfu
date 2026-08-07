// Command catfu is the YouTube channel metadata catalogue CLI and MCP server.
package main

import (
	"os"

	"github.com/coldcanuk/catfu/internal/cli"
)

func main() {
	// Ensure Cobra context cancellation on signals is available via default.
	cli.Execute()
	os.Exit(0)
}
