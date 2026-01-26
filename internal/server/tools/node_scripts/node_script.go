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

type NodeScript struct{}

func (NodeScript) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{
		DeleteNodeScript{},
		ReadNodeScript{},
		UpdateNodeScript{},
		AddTagToNodeScript{},
		DownloadNodeScript{},
		RemoveTagFromNodeScript{},
	}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type DeleteNodeScript struct{}

func (DeleteNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"delete-node-script",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The script's name."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Delete Node Script", false, true, false, true)),
		mcp.WithDescription("Deletes a script with the given name."),
	)
}

func (DeleteNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	scriptName, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeleteNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/scripts/%s", scriptName)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[DeleteNodeScript] Deleting script with name: %s", scriptName))
	resultData, err := client.Do(ctx, maas_client.RequestTypeDelete, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to delete script %s err=%v", scriptName, err)
		zap.L().Error(fmt.Sprintf("[DeleteNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[DeleteNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ReadNodeScript struct{}

func (ReadNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"read-node-script",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The script's name."),
		),
		mcp.WithString(
			"include_script",
			mcp.Description("Include the base64 encoded script content if any value is given for include_script."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Read Node Script", true, false, false, true)),
		mcp.WithDescription("Return metadata belonging to the script with the given name."),
	)
}

func (ReadNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	scriptName, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/scripts/%s", scriptName)

	if includeScript := request.GetString("include_script", ""); includeScript != "" {
		queryParams := url.Values{}
		queryParams.Add("include_script", includeScript)
		path += "?" + queryParams.Encode()
	}

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ReadNodeScript] Retrieving script with name: %s", scriptName))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to read script %s err=%v", scriptName, err)
		zap.L().Error(fmt.Sprintf("[ReadNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ReadNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type UpdateNodeScript struct{}

func (UpdateNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"update-node-script",
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
			mcp.Description("The type defines when the script should be used. Can be commissioning, testing or release. It defaults to testing."),
		),
		mcp.WithString(
			"hardware_type",
			mcp.Enum("cpu", "memory", "storage", "network", "node"),
			mcp.Description("The hardware_type defines what type of hardware the script is associated with. May be cpu, memory, storage, network, or node."),
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
		mcp.WithBoolean(
			"destructive",
			mcp.Description("Whether or not the script overwrites data on any drive on the running system. Destructive scripts can not be run on deployed systems. Defaults to false."),
		),
		mcp.WithBoolean(
			"may_reboot",
			mcp.Description("Whether or not the script may reboot the system while running."),
		),
		mcp.WithBoolean(
			"recommission",
			mcp.Description("Whether built-in commissioning scripts should be rerun after successfully running this script."),
		),
		mcp.WithBoolean(
			"apply_configured_networking",
			mcp.Description("Whether to apply the provided network configuration before the script runs."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Update Node Script", false, false, false, true)),
		mcp.WithDescription("Update a script with the given name."),
	)
}

func (UpdateNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	scriptName, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[UpdateNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)

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

	destructive := request.GetBool("destructive", false)
	if destructive {
		form.Add("destructive", "1")
	} else {
		form.Add("destructive", "0")
	}

	mayReboot := request.GetBool("may_reboot", false)
	if mayReboot {
		form.Add("may_reboot", "1")
	} else {
		form.Add("may_reboot", "0")
	}

	recommission := request.GetBool("recommission", false)
	if recommission {
		form.Add("recommission", "1")
	} else {
		form.Add("recommission", "0")
	}

	applyConfiguredNetworking := request.GetBool("apply_configured_networking", false)
	if applyConfiguredNetworking {
		form.Add("apply_configured_networking", "1")
	} else {
		form.Add("apply_configured_networking", "0")
	}

	path := fmt.Sprintf("/MAAS/api/2.0/scripts/%s", scriptName)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[UpdateNodeScript] Updating script with name: %s", scriptName))
	resultData, err := client.Do(ctx, maas_client.RequestTypePut, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to update script %s err=%v", scriptName, err)
		zap.L().Error(fmt.Sprintf("[UpdateNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[UpdateNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type AddTagToNodeScript struct{}

func (AddTagToNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"add-tag-to-node-script",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the script."),
		),
		mcp.WithString(
			"tag",
			mcp.Description("The tag being added."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Add Tag to Node Script", false, false, false, true)),
		mcp.WithDescription("Add a single tag to a script with the given name."),
	)
}

func (AddTagToNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	scriptName, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[AddTagToNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)
	if tag := request.GetString("tag", ""); tag != "" {
		form.Add("tag", tag)
	}

	path := fmt.Sprintf("/MAAS/api/2.0/scripts/%sop-add_tag", scriptName)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[AddTagToNodeScript] Adding tag to script with name: %s", scriptName))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to add tag to script %s err=%v", scriptName, err)
		zap.L().Error(fmt.Sprintf("[AddTagToNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[AddTagToNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type DownloadNodeScript struct{}

func (DownloadNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"download-node-script",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the script."),
		),
		mcp.WithString(
			"revision",
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("What revision to download, latest by default. Can use rev as a shortcut."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Download Node Script", true, false, false, true)),
		mcp.WithDescription("Download a script with the given name."),
	)
}

func (DownloadNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	scriptName, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DownloadNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/scripts/%sop-download", scriptName)

	if revision := request.GetString("revision", ""); revision != "" {
		queryParams := url.Values{}
		queryParams.Add("revision", revision)
		path += "?" + queryParams.Encode()
	}

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[DownloadNodeScript] Downloading script with name: %s", scriptName))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to download script %s err=%v", scriptName, err)
		zap.L().Error(fmt.Sprintf("[DownloadNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", resultData)), nil
}

type RemoveTagFromNodeScript struct{}

func (RemoveTagFromNodeScript) Create() mcp.Tool {
	return mcp.NewTool(
		"remove-tag-from-node-script",
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the script."),
		),
		mcp.WithString(
			"tag",
			mcp.Description("The tag being removed."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Remove Tag from Node Script", false, false, false, true)),
		mcp.WithDescription("Remove a tag from a script with the given name."),
	)
}

func (RemoveTagFromNodeScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	scriptName, err := request.RequireString("name")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[RemoveTagFromNodeScript] Required parameter name not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)
	if tag := request.GetString("tag", ""); tag != "" {
		form.Add("tag", tag)
	}

	path := fmt.Sprintf("/MAAS/api/2.0/scripts/%sop-remove_tag", scriptName)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[RemoveTagFromNodeScript] Removing tag from script with name: %s", scriptName))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to remove tag from script %s err=%v", scriptName, err)
		zap.L().Error(fmt.Sprintf("[RemoveTagFromNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[RemoveTagFromNodeScript] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
