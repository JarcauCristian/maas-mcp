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

type Vlan struct{}

func (Vlan) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{DeleteVlan{}, ReadVlan{}, UpdateVlan{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type DeleteVlan struct{}

func (DeleteVlan) Create() mcp.Tool {
	return mcp.NewTool(
		"delete-vlan",
		mcp.WithString(
			"fabric_id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("Fabric ID containing the VLAN to delete."),
		),
		mcp.WithString(
			"vid",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("VLAN ID of the VLAN to delete."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Delete VLAN", false, true, false, true)),
		mcp.WithDescription("Delete a VLAN on a given fabric."),
	)
}

func (DeleteVlan) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("fabric_id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeleteVlan] Required parameter fabric_id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	vid, err := request.RequireString("vid")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeleteVlan] Required parameter vid not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/vlans/%s/", fabricID, vid)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[DeleteVlan] Deleting VLAN %s on fabric %s", vid, fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeDelete, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to delete VLAN %s on fabric %s err=%v", vid, fabricID, err)
		zap.L().Error(fmt.Sprintf("[DeleteVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[DeleteVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ReadVlan struct{}

func (ReadVlan) Create() mcp.Tool {
	return mcp.NewTool(
		"read-vlan",
		mcp.WithString(
			"fabric_id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The fabric_id containing the VLAN."),
		),
		mcp.WithString(
			"vid",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The VLAN ID."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Read VLAN", true, false, false, true)),
		mcp.WithDescription("Retrieve information about a VLAN on a given fabric."),
	)
}

func (ReadVlan) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("fabric_id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadVlan] Required parameter fabric_id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	vid, err := request.RequireString("vid")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadVlan] Required parameter vid not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/vlans/%s/", fabricID, vid)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ReadVlan] Retrieving VLAN %s on fabric %s", vid, fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to read VLAN %s on fabric %s err=%v", vid, fabricID, err)
		zap.L().Error(fmt.Sprintf("[ReadVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ReadVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type UpdateVlan struct{}

func (UpdateVlan) Create() mcp.Tool {
	return mcp.NewTool(
		"update-vlan",
		mcp.WithString(
			"fabric_id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("Fabric ID containing the VLAN."),
		),
		mcp.WithString(
			"vid",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("VLAN ID of the VLAN."),
		),
		mcp.WithString(
			"name",
			mcp.Description("Name of the VLAN."),
		),
		mcp.WithString(
			"description",
			mcp.Description("Description of the VLAN."),
		),
		mcp.WithString(
			"mtu",
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The MTU to use on the VLAN."),
		),
		mcp.WithString(
			"space",
			mcp.Description("The space this VLAN should be placed in. Passing in an empty string (or the string 'undefined') will cause the VLAN to be placed in the 'undefined' space."),
		),
		mcp.WithBoolean(
			"dhcp_on",
			mcp.Description("Whether or not DHCP should be managed on the VLAN."),
		),
		mcp.WithString(
			"primary_rack",
			mcp.Description("The primary rack controller managing the VLAN (system_id)."),
		),
		mcp.WithString(
			"secondary_rack",
			mcp.Description("The secondary rack controller managing the VLAN (system_id)."),
		),
		mcp.WithString(
			"relay_vlan",
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("Relay VLAN ID. Only set when this VLAN will be using a DHCP relay to forward DHCP requests to another VLAN that MAAS is managing."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Update VLAN", false, false, false, true)),
		mcp.WithDescription("Update a VLAN on a given fabric."),
	)
}

func (UpdateVlan) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	fabricID, err := request.RequireString("fabric_id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[UpdateVlan] Required parameter fabric_id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	vid, err := request.RequireString("vid")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[UpdateVlan] Required parameter vid not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)

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
	if primaryRack := request.GetString("primary_rack", ""); primaryRack != "" {
		form.Add("primary_rack", primaryRack)
	}
	if secondaryRack := request.GetString("secondary_rack", ""); secondaryRack != "" {
		form.Add("secondary_rack", secondaryRack)
	}
	if relayVlan := request.GetString("relay_vlan", ""); relayVlan != "" {
		form.Add("relay_vlan", relayVlan)
	}

	dhcpOn := request.GetBool("dhcp_on", false)
	if dhcpOn {
		form.Add("dhcp_on", "1")
	} else {
		form.Add("dhcp_on", "0")
	}

	path := fmt.Sprintf("/MAAS/api/2.0/fabrics/%s/vlans/%s/", fabricID, vid)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[UpdateVlan] Updating VLAN %s on fabric %s", vid, fabricID))
	resultData, err := client.Do(ctx, maas_client.RequestTypePut, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to update VLAN %s on fabric %s err=%v", vid, fabricID, err)
		zap.L().Error(fmt.Sprintf("[UpdateVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[UpdateVlan] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
