package fabrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/JarcauCristian/ztp-mcp/internal/server/maas_client"
	"github.com/JarcauCristian/ztp-mcp/internal/server/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type Fabric struct{}

func (Fabric) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{DeleteFabric{}, ReadFabric{}, UpdateFabric{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type DeleteFabric struct{}

func (DeleteFabric) Create() mcp.Tool {
	return mcp.NewTool(
		"delete-fabric",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the fabric to delete."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Delete Fabric", false, true, false, true)),
		mcp.WithDescription("Delete a fabric with the given ID."),
	)
}

func (DeleteFabric) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeleteFabric] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/", fabricID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[DeleteFabric] Deleting fabric with ID: %s", fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeDelete, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to delete fabric %s err=%v", fabricID, err)
		zap.L().Error(fmt.Sprintf("[DeleteFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[DeleteFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ReadFabric struct{}

func (ReadFabric) Create() mcp.Tool {
	return mcp.NewTool(
		"read-fabric",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the fabric to retrieve."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Read Fabric", true, false, false, true)),
		mcp.WithDescription("Read a fabric with the given ID."),
	)
}

func (ReadFabric) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadFabric] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/", fabricID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ReadFabric] Retrieving fabric with ID: %s", fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to read fabric %s err=%v", fabricID, err)
		zap.L().Error(fmt.Sprintf("[ReadFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ReadFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type UpdateFabric struct{}

func (UpdateFabric) Create() mcp.Tool {
	return mcp.NewTool(
		"update-fabric",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the fabric to update."),
		),
		mcp.WithString(
			"name",
			mcp.Description("Name of the fabric."),
		),
		mcp.WithString(
			"description",
			mcp.Description("Description of the fabric."),
		),
		mcp.WithString(
			"class_type",
			mcp.Description("Class type of the fabric."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Update Fabric", false, false, false, true)),
		mcp.WithDescription("Update a fabric with the given ID."),
	)
}

func (UpdateFabric) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[UpdateFabric] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)

	if name := request.GetString("name", ""); name != "" {
		form.Add("name", name)
	}
	if description := request.GetString("description", ""); description != "" {
		form.Add("description", description)
	}
	if classType := request.GetString("class_type", ""); classType != "" {
		form.Add("class_type", classType)
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/", fabricID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[UpdateFabric] Updating fabric with ID: %s", fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypePut, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to update fabric %s err=%v", fabricID, err)
		zap.L().Error(fmt.Sprintf("[UpdateFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[UpdateFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
