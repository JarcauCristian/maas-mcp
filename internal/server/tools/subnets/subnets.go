package subnets

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

type Subnets struct{}

func (Subnets) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{ListSubnets{}, CreateSubnet{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListSubnets struct{}

func (ListSubnets) Create() mcp.Tool {
	return mcp.NewTool(
		"list-subnets",
		mcp.WithInputSchema[struct{}](),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("List Subnets", true, false, false, true)),
		mcp.WithDescription("Returns all subnets that are currently defined on the running instance of MAAS."),
	)
}

func (ListSubnets) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/subnets/"

	client := maas_client.MustClient()

	zap.L().Info("[ListSubnets] Retrieving all subnets...")
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve all the subnets: %v", err)
		zap.L().Error(fmt.Sprintf("[ListSubnets] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListSubnets] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type CreateSubnet struct{}

func (CreateSubnet) Create() mcp.Tool {
	return mcp.NewTool(
		"create-subnet",
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Create Subnet", false, true, false, true)),
		mcp.WithString(
			"cidr",
			mcp.Required(),
			mcp.Description("The network CIDR for this subnet (e.g., 192.168.1.0/24)."),
		),
		mcp.WithString(
			"name",
			mcp.Description("A human-readable name for this subnet."),
		),
		mcp.WithString(
			"description",
			mcp.Description("A description of this subnet."),
		),
		mcp.WithString(
			"vlan",
			mcp.Description("VLAN this subnet belongs to. If not provided, the subnet will be created in the default VLAN."),
		),
		mcp.WithString(
			"fabric",
			mcp.Description("Fabric for the subnet. If not provided, the subnet will be created in the default fabric."),
		),
		mcp.WithString(
			"vid",
			mcp.Description("VLAN ID. Only used when VLAN is not provided."),
		),
		mcp.WithString(
			"space",
			mcp.Description("Space this subnet should be placed in."),
		),
		mcp.WithString(
			"gateway_ip",
			mcp.Description("Gateway IP address for the subnet."),
		),
		mcp.WithString(
			"dns_servers",
			mcp.Description("Comma-separated list of DNS servers for this subnet."),
		),
		mcp.WithBoolean(
			"managed",
			mcp.DefaultBool(true),
			mcp.Description("Whether MAAS manages DHCP and DNS for this subnet."),
		),
		mcp.WithDescription("Create a new subnet on the running instance of MAAS with the provided information."),
	)
}

func (CreateSubnet) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	path := "/MAAS/api/2.0/subnets/"

	cidr, err := request.RequireString("cidr")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CreateSubnet] Required parameter cidr not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)
	form.Add("cidr", cidr)

	if name := request.GetString("name", ""); name != "" {
		form.Add("name", name)
	}
	if description := request.GetString("description", ""); description != "" {
		form.Add("description", description)
	}
	if vlan := request.GetString("vlan", ""); vlan != "" {
		form.Add("vlan", vlan)
	}
	if fabric := request.GetString("fabric", ""); fabric != "" {
		form.Add("fabric", fabric)
	}
	if vid := request.GetString("vid", ""); vid != "" {
		form.Add("vid", vid)
	}
	if space := request.GetString("space", ""); space != "" {
		form.Add("space", space)
	}
	if gatewayIP := request.GetString("gateway_ip", ""); gatewayIP != "" {
		form.Add("gateway_ip", gatewayIP)
	}
	if dnsServers := request.GetString("dns_servers", ""); dnsServers != "" {
		form.Add("dns_servers", dnsServers)
	}

	managed := request.GetBool("managed", true)
	if managed {
		form.Add("managed", "1")
	} else {
		form.Add("managed", "0")
	}

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[CreateSubnet] Creating subnet with CIDR: %s", cidr))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to create subnet err=%v", err)
		zap.L().Error(fmt.Sprintf("[CreateSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
