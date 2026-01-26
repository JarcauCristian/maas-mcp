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

type Fabrics struct{}

func (Fabrics) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{ListFabrics{}, CreateFabric{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListFabrics struct{}

func (ListFabrics) Create() mcp.Tool {
	return mcp.NewTool(
		"list-fabrics",
		mcp.WithInputSchema[struct{}](),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("List Fabrics", true, false, false, true)),
		mcp.WithDescription("This tool is used to return all the fabrics that are currently defined on the running instance of MAAS."),
	)
}

func (ListFabrics) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/fabrics/"

	client := maas_client.MustClient()

	zap.L().Info("[ListFabrics] Retrieving all fabrics...")
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve all the fabrics: %v", err)
		zap.L().Error(fmt.Sprintf("[ListFabrics] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListFabrics] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type CreateFabric struct{}

func (CreateFabric) Create() mcp.Tool {
	return mcp.NewTool(
		"create-fabric",
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
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Create Fabric", false, false, false, true)),
		mcp.WithDescription("Tool used to create a new fabric on the running instance of MAAS with the provided information."),
	)
}

func (CreateFabric) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/fabrics/"

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

	client := maas_client.MustClient()

	zap.L().Info("[CreateFabric] Creating fabric...")
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to create fabric err=%v", err)
		zap.L().Error(fmt.Sprintf("[CreateFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateFabric] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
