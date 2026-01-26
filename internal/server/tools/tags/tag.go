package tags

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

type Tag struct{}

func (Tag) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{DeleteTag{}, ReadTag{}, UpdateTag{}, ListByTag{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type DeleteTag struct{}

func (DeleteTag) Create() mcp.Tool {
	return mcp.NewTool(
		"delete-tag",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the tag that will be deleted."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Delete Tag", false, true, false, true)),
		mcp.WithDescription("Tool used to delete a specific tag from the running instance."),
	)
}

func (DeleteTag) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	name, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeleteTag] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := "/MAAS/api/2.0/tags/" + name + "/"

	client := maas_client.MustClient()

	resultData, err := client.Do(ctx, maas_client.RequestTypeDelete, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to delete tag %s err=%v", name, err)
		zap.L().Error(fmt.Sprintf("[DeleteTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[DeleteTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ReadTag struct{}

func (ReadTag) Create() mcp.Tool {
	return mcp.NewTool(
		"read-tag",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the tag to query."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Read Tag", true, false, false, true)),
		mcp.WithDescription("Return information about a specified tag by name."),
	)
}

func (ReadTag) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	name, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadTag] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := "/MAAS/api/2.0/tags/" + name + "/"

	client := maas_client.MustClient()

	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to read tag %s err=%v", name, err)
		zap.L().Error(fmt.Sprintf("[ReadTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ReadTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type UpdateTag struct{}

func (UpdateTag) Create() mcp.Tool {
	return mcp.NewTool(
		"update-tag",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the tag to update."),
		),
		mcp.WithString(
			"new_name",
			mcp.Description("The new name of the tag."),
		),
		mcp.WithString(
			"comment",
			mcp.Description("Updated description of the tag."),
		),
		mcp.WithString(
			"definition",
			mcp.Description("An XPATH query that is evaluated against the hardware_details stored for all nodes (i.e. the output of `lshw -xml`)."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Read Tag", false, true, false, true)),
		mcp.WithDescription("Return information about a specified tag by name."),
	)
}

func (UpdateTag) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	name, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadTag] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)
	newName := request.GetString("new_name", "")
	if newName != "" {
		form.Add("name", newName)
	}

	comment := request.GetString("comment", "")
	if newName != "" {
		form.Add("comment", comment)
	}

	definition := request.GetString("definition", "")
	if newName != "" {
		form.Add("definition", definition)
	}

	path := "/MAAS/api/2.0/tags/" + name + "/"

	client := maas_client.MustClient()

	resultData, err := client.Do(ctx, maas_client.RequestTypePut, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to read tag %s err=%v", name, err)
		zap.L().Error(fmt.Sprintf("[ReadTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ReadTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ListByTag struct{}

func (ListByTag) Create() mcp.Tool {
	return mcp.NewTool(
		"list-by-tag",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the tag to retrieve elements of type."),
		),
		mcp.WithString(
			"type",
			mcp.Enum(
				"nodes",
				"devices",
				"machines",
				"rack_controllers",
				"region_controllers",
			),
			mcp.Required(),
			mcp.Description("The type of element to return that contain the tag."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("List Node Type by Tag", true, false, false, true)),
		mcp.WithDescription("Returns all the elements of the specified type that have the tag."),
	)
}

func (ListByTag) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	name, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ListByTag] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}
	path := "/MAAS/api/2.0/tags/" + name + "/"

	nodeType, err := request.RequireString("type")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ListByTag] Required parameter type not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	switch nodeType {
	case "nodes":
		path += "op-nodes"
	case "devices":
		path += "op-devices"
	case "machines":
		path += "op-machines"
	case "rack_controllers":
		path += "op-rack_controllers"
	case "region_controllers":
		path += "op-region_controllers"
	}

	client := maas_client.MustClient()

	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to get elements of type %s for tag %s err=%v", nodeType, name, err)
		zap.L().Error(fmt.Sprintf("[ListByTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListByTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
