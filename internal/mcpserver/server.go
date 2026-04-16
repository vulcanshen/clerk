package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/vulcanshen/clerk/internal/config"
)

var toolNames []string

func ToolNames() []string {
	return toolNames
}

func addTool(s *server.MCPServer, tool mcp.Tool, handler server.ToolHandlerFunc) {
	toolNames = append(toolNames, tool.Name)
	s.AddTool(tool, handler)
}

func NewServer(version string) *server.MCPServer {
	s := server.NewMCPServer(
		"clerk",
		version,
		server.WithToolCapabilities(false),
	)

	addTool(s, mcp.NewTool("clerk-resume",
			mcp.WithDescription(`Recover context from previous Claude Code sessions in this project.

Use this tool when:
- The user forgot to use "claude -c" or "--resume" and wants to continue previous work
- The user asks to "resume", "recover context", "what were we doing", or similar
- You need background on what was previously done in this project

Returns file paths to:
1. Summary files — daily digests of past sessions (read these first for a quick overview)
2. Transcript files — full conversation history from previous sessions (read for detailed context)

After receiving the paths, read the summary files first. If more detail is needed, read the relevant transcript files to fully rebuild context.`),
			mcp.WithString("cwd",
				mcp.Required(),
				mcp.Description("Absolute path to the project working directory."),
			),
		),
		handleResume,
	)

	addTool(s, mcp.NewTool("clerk-search",
			mcp.WithDescription(`Search previous sessions by keyword/tag.

Use this tool when:
- The user asks "what did we do with X", "find sessions about Y", or similar
- You need to find past work related to a specific technology, tool, or concept

Returns file paths to matching summaries and transcripts, organized by tag. Read the returned files to provide context.`),
			mcp.WithString("keyword",
				mcp.Required(),
				mcp.Description("Keyword to search for (e.g. mcp, refactor, auth, docker)"),
			),
		),
		handleSearch,
	)

	return s
}

func handleResume(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cwd := request.GetString("cwd", "")
	if cwd == "" {
		return mcp.NewToolResultError("cwd is required — pass the project working directory"), nil
	}

	cfg, err := config.Load()
	if err != nil {
		return mcp.NewToolResultError("failed to load config: " + err.Error()), nil
	}

	result, err := Resume(cwd, cfg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result), nil
}

func handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword := request.GetString("keyword", "")
	if keyword == "" {
		return mcp.NewToolResultError("keyword is required"), nil
	}

	cfg, err := config.Load()
	if err != nil {
		return mcp.NewToolResultError("failed to load config: " + err.Error()), nil
	}

	result, err := Search(keyword, cfg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result), nil
}
