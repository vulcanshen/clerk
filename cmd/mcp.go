package cmd

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/vulcanshen/clerk/internal/mcpserver"
)

var mcpCmd = &cobra.Command{
	Use:               "mcp",
	Short:             "Start MCP stdio server",
	Hidden:            true,
	ValidArgsFunction: noFileComp,
	Long:  "Starts a Model Context Protocol server over stdio. Used by Claude Code to interact with clerk tools.",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := mcpserver.NewServer(Version)
		return server.ServeStdio(s)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
