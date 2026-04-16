package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/JarcauCristian/ztp-mcp/internal/server/maas_client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type Events struct{}

func (Events) Register(mcpServer *server.MCPServer) {
	mcpTools := []MCPTool{GetEvents{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type GetEvents struct{}

func (GetEvents) Create() mcp.Tool {
	return mcp.NewTool(
		"get-events",
		mcp.WithString(
			"level",
			mcp.Enum(
				"AUDIT",
				"CRITICAL",
				"DEBUG",
				"ERROR",
				"INFO",
				"WARNING",
			),
			mcp.Description("The minimum level type of the events to return."),
		),
		mcp.WithNumber(
			"limit",
			mcp.DefaultNumber(1000.0),
			mcp.Max(10000.0),
			mcp.Description("How many events to return."),
		),
		mcp.WithString(
			"before",
			mcp.DefaultString(""),
			mcp.Description("The id of the event to return the events before it."),
		),
		mcp.WithString(
			"after",
			mcp.DefaultString(""),
			mcp.Description("The id of the event to return the events after it."),
		),
		mcp.WithDescription("Get all the events from the MAAS envrionment."),
	)
}

func (GetEvents) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	level := request.GetString("level", "")
	limit := request.GetFloat("limit", 1000)
	before := request.GetString("before", "")
	after := request.GetString("after", "")

	query := url.Values{}
	query.Add("limit", fmt.Sprintf("%.0f", limit))

	if level != "" {
		query.Add("level", level)
	}
	if before != "" {
		query.Add("before", before)
	}
	if after != "" {
		query.Add("after", after)
	}

	path := fmt.Sprintf("/MAAS/api/2.0/events/?%s", query.Encode())

	client := maas_client.MustClient()

	zap.L().Info("[GetEvents] Retrieving events...")
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve events: %v", err)
		zap.L().Error(fmt.Sprintf("[GetEvents] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	var events []map[string]any
	if err := json.Unmarshal([]byte(resultData), &events); err != nil {
		errMsg = fmt.Sprintf("Failed to unmarshal events: %v", err)
		zap.L().Error(fmt.Sprintf("[GetEvents] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	response, err := json.Marshal(events)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to marshal events: %v", err)
		zap.L().Error(fmt.Sprintf("[GetEvents] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(response)), nil
}
