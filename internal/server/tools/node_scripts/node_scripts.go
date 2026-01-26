package nodescripts

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

type NodeScripts struct{}

func (NodeScripts) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{ListNodeScripts{}, CreateNodeScript{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListNodeScripts struct{}

func (ListNodeScripts) Create() mcp.Tool {
	return mcp.NewTool(
		"list-node-scripts",
		mcp.WithString(
			"type",
			mcp.Enum("commissioning", "testing", "release"),
			mcp.Description("Only return scripts with the given type. This can be commissioning, testing or release. Defaults to showing all."),
		),
		mcp.WithString(
			"hardware_type",
			mcp.Enum("cpu", "memory", "storage", "network", "node"),
			mcp.Description("Only return scripts for the given hardware type. Can be cpu, memory, storage, network, or node. Defaults to all."),
		),
		mcp.WithString(
			"include_script",
			mcp.Description("Include the base64-encoded script content."),
		),
		mcp.WithString(
			"filters",
			mcp.Description("A comma separated list to show only results with a script name or tag."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("List Node Scripts", true, false, false, true)),
		mcp.WithDescription("Return a list of stored scripts. Note that parameters should be passed in the URI."),
	)
}

func (ListNodeScripts) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/scripts/"

	// Build query parameters
	queryParams := url.Values{}

	if scriptType := request.GetString("type", ""); scriptType != "" {
		queryParams.Add("type", scriptType)
	}
	if hardwareType := request.GetString("hardware_type", ""); hardwareType != "" {
		queryParams.Add("hardware_type", hardwareType)
	}
	if includeScript := request.GetString("include_script", ""); includeScript != "" {
		queryParams.Add("include_script", includeScript)
	}
	if filters := request.GetString("filters", ""); filters != "" {
		queryParams.Add("filters", filters)
	}

	// Add query parameters to path if any exist
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	client := maas_client.MustClient()

	zap.L().Info("[ListNodeScripts] Retrieving all node scripts...")
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve all the node scripts: %v", err)
		zap.L().Error(fmt.Sprintf("[ListNodeScripts] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListNodeScripts] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type CreateNodeScript struct{}

func (CreateNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"create-node-script",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the script."),
		),
		mcp.WithString(
			"script",
			mcp.Description("The content of the script to be uploaded in binary form."),
		),
		mcp.WithString(
			"type",
			mcp.Enum("commissioning", "testing", "release"),
			mcp.Description("The script_type defines when the script should be used: commissioning or testing or release. Defaults to testing."),
		),
		mcp.WithString(
			"hardware_type",
			mcp.Enum("cpu", "memory", "storage", "network", "node"),
			mcp.Description("The hardware_type defines what type of hardware the script is associated with. May be CPU, memory, storage, network, or node."),
		),
		mcp.WithString(
			"title",
			mcp.Description("The title of the script."),
		),
		mcp.WithString(
			"description",
			mcp.Description("A description of what the script does."),
		),
		mcp.WithString(
			"tags",
			mcp.Description("A comma separated list of tags for this script."),
		),
		mcp.WithString(
			"timeout",
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("How long the script is allowed to run before failing. 0 gives unlimited time, defaults to 0."),
		),
		mcp.WithString(
			"comment",
			mcp.Description("A comment about what this change does."),
		),
		mcp.WithString(
			"for_hardware",
			mcp.Description("A list of modalias, PCI IDs, and/or USB IDs the script will automatically run on. Must start with modalias:, pci:, or usb:."),
		),
		mcp.WithString(
			"parallel",
			mcp.Pattern("^[0-1]$"),
			mcp.Description("Whether the script may be run in parallel with other scripts. 1 = True, 0 = False."),
		),
		mcp.WithString(
			"recommission",
			mcp.Description("Whether builtin commissioning scripts should be rerun after successfully running this script."),
		),
		mcp.WithBoolean(
			"destructive",
			mcp.Description("Whether or not the script overwrites data on any drive on the running system. Destructive scripts can not be run on deployed systems. Defaults to false."),
		),
		mcp.WithBoolean(
			"may_reboot",
			mcp.Description("Whether or not the script may reboot the system while running."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Create Node Script", false, false, false, true)),
		mcp.WithDescription("Create a new script."),
	)
}

func (CreateNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/scripts/"

	name, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CreateNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)
	form.Add("name", name)

	// Add optional string parameters
	if script := request.GetString("script", ""); script != "" {
		form.Add("script", script)
	}
	if scriptType := request.GetString("type", ""); scriptType != "" {
		form.Add("type", scriptType)
	}
	if hardwareType := request.GetString("hardware_type", ""); hardwareType != "" {
		form.Add("hardware_type", hardwareType)
	}
	if title := request.GetString("title", ""); title != "" {
		form.Add("title", title)
	}
	if description := request.GetString("description", ""); description != "" {
		form.Add("description", description)
	}
	if tags := request.GetString("tags", ""); tags != "" {
		form.Add("tags", tags)
	}
	if timeout := request.GetString("timeout", ""); timeout != "" {
		form.Add("timeout", timeout)
	}
	if comment := request.GetString("comment", ""); comment != "" {
		form.Add("comment", comment)
	}
	if forHardware := request.GetString("for_hardware", ""); forHardware != "" {
		form.Add("for_hardware", forHardware)
	}
	if parallel := request.GetString("parallel", ""); parallel != "" {
		form.Add("parallel", parallel)
	}
	if recommission := request.GetString("recommission", ""); recommission != "" {
		form.Add("recommission", recommission)
	}

	destructive := request.GetBool("destructive", false)
	if destructive {
		form.Add("destructive", "1")
	} else {
		form.Add("destructive", "0")
	}

	may_reboot := request.GetBool("may_reboot", false)
	if may_reboot {
		form.Add("may_reboot", "1")
	} else {
		form.Add("may_reboot", "0")
	}

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[CreateNodeScript] Creating node script with name: %s", name))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to create node script err=%v", err)
		zap.L().Error(fmt.Sprintf("[CreateNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
