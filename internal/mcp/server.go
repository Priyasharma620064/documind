package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/priya-sharma/documind/internal/search"
)

// Server implements the Model Context Protocol (MCP) interface.
type Server struct {
	mcpServer *server.MCPServer
	engine    *search.Engine
}

// NewServer creates a new MCP server.
func NewServer(name, version string, engine *search.Engine) *Server {
	s := server.NewMCPServer(name, version)
	
	mcpSrv := &Server{
		mcpServer: s,
		engine:    engine,
	}

	mcpSrv.registerTools()
	return mcpSrv
}

// Start starts the MCP server using stdio transport.
func (s *Server) Start() error {
	slog.Info("Starting MCP server (stdio transport)")
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) registerTools() {
	// 1. docs_lookup tool
	s.mcpServer.AddTool(mcp.NewTool("docs_lookup",
		mcp.WithDescription("Perform semantic documentation lookup across repositories"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Semantic search query"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name to search in"),
		),
		mcp.WithNumber("top_k",
			mcp.Description("Number of results to return"),
		),
	), s.docsLookupHandler)
}

func (s *Server) docsLookupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Arguments is map[string]any in the current SDK version
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}
	
	query, _ := args["query"].(string)
	repo, _ := args["repo"].(string)
	
	topK := 5
	if val, ok := args["top_k"].(float64); ok {
		topK = int(val)
	}

	results, err := s.engine.Search(ctx, repo, query, topK, "")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	var output string
	for _, res := range results {
		output += fmt.Sprintf("### %s\nPath: %s\nScore: %.2f\n\n%s\n\n---\n\n",
			res.Chunk.HeadingPath, res.Chunk.FilePath, res.Score, res.Chunk.Content)
	}

	if len(results) == 0 {
		output = "No results found."
	}

	return mcp.NewToolResultText(output), nil
}
