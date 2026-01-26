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

type Tags struct{}

func (Tags) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{ListTags{}, CreateTag{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListTags struct{}

func (ListTags) Create() mcp.Tool {
	return mcp.NewTool(
		"read-tags",
		mcp.WithInputSchema[struct{}](),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("List Tags", true, false, false, true)),
		mcp.WithDescription("This tools is used to return all the tags that are currently defined on the running instance of MAAS."),
	)
}

func (ListTags) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/tags/"

	client := maas_client.MustClient()

	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve all the tags: %v", err)
		zap.L().Error(fmt.Sprintf("[ListTags] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListTags] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type CreateTag struct{}

func (CreateTag) Create() mcp.Tool {
	return mcp.NewTool(
		"create-tag",
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Create Tag", true, false, false, true)),
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the tag that will be created."),
		),
		mcp.WithString(
			"comment",
			mcp.Required(),
			mcp.Description("Description of the tag and what it should be used for."),
		),
		mcp.WithString(
			"definition",
			mcp.Description("An XPATH query that is evaluated against the hardware_details stored for all nodes (i.e. the output of `lshw -xml`)."),
		),
		mcp.WithString(
			"kernel_opts",
			mcp.Description("Nodes associated with this tag will add this string to their kernel options when booting. The value overrides the global `kernel_opts` setting. If more than one tag is associated with a node, command line will be concatenated from all associated tags, in alphabetic tag name order."),
		),
		mcp.WithDescription("Tool used to create a new tag on the running instance of MAAS with the provided information."),
	)
}

func (CreateTag) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/tags/"

	name, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CreateTag] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	comment, err := request.RequireString("comment")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CreateTag] Required parameter comment not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	definition := request.GetString("definition", "")
	kernelOpts := request.GetString("kernel_opts", "")

	form := make(url.Values)
	form.Add("name", name)
	form.Add("comment", comment)
	form.Add("definition", definition)
	form.Add("kernel_opts", kernelOpts)

	client := maas_client.MustClient()

	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to create tag err=%v", err)
		zap.L().Error(fmt.Sprintf("[CreateTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateTag] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
