package vlans

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

const NUMBER_PATTERN = "^[0-9]+$"

type Vlans struct{}

func (Vlans) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{ListVlans{}, CreateVlan{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListVlans struct{}

func (ListVlans) Create() mcp.Tool {
	return mcp.NewTool(
		"list-vlans",
		mcp.WithString(
			"fabric_id",
			mcp.Required(),
			mcp.Pattern(NUMBER_PATTERN),
			mcp.Description("The fabric ID for which to list the VLANs."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("List VLANs", true, false, false, true)),
		mcp.WithDescription("This tool is used to return all the VLANs that belong to the given fabric on the running instance of MAAS."),
	)
}

func (ListVlans) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("fabric_id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ListVlans] Required parameter fabric_id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/vlans/", fabricID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ListVlans] Retrieving all VLANs for fabric ID: %s", fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve all the VLANs for fabric %s: %v", fabricID, err)
		zap.L().Error(fmt.Sprintf("[ListVlans] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListVlans] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type CreateVlan struct{}

func (CreateVlan) Create() mcp.Tool {
	return mcp.NewTool(
		"create-vlan",
		mcp.WithString(
			"fabric_id",
			mcp.Required(),
			mcp.Pattern(NUMBER_PATTERN),
			mcp.Description("The fabric ID on which to add the new VLAN."),
		),
		mcp.WithString(
			"vid",
			mcp.Required(),
			mcp.Pattern(NUMBER_PATTERN),
			mcp.Description("VLAN ID of the new VLAN."),
		),
		mcp.WithString(
			"name",
			mcp.Description("Name of the VLAN."),
		),
		mcp.WithString(
			"description",
			mcp.Description("Description of the new VLAN."),
		),
		mcp.WithString(
			"mtu",
			mcp.Pattern(NUMBER_PATTERN),
			mcp.Description("The MTU to use on the VLAN."),
		),
		mcp.WithString(
			"space",
			mcp.Description("The space this VLAN should be placed in. Passing in an empty string (or the string 'undefined') will cause the VLAN to be placed in the 'undefined' space."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Create VLAN", false, false, false, true)),
		mcp.WithDescription("Tool used to create a new VLAN on the running instance of MAAS with the provided information."),
	)
}

func (CreateVlan) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("fabric_id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CreateVlan] Required parameter fabric_id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	vid, err := request.RequireString("vid")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CreateVlan] Required parameter vid not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/vlans/", fabricID)

	form := make(url.Values)
	form.Add("vid", vid)

	if name := request.GetString("name", ""); name != "" {
		form.Add("name", name)
	}
	if description := request.GetString("description", ""); description != "" {
		form.Add("description", description)
	}
	if mtu := request.GetString("mtu", ""); mtu != "" {
		form.Add("mtu", mtu)
	}
	if space := request.GetString("space", ""); space != "" {
		form.Add("space", space)
	}

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[CreateVlan] Creating VLAN with VID %s on fabric %s", vid, fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to create VLAN err=%v", err)
		zap.L().Error(fmt.Sprintf("[CreateVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
