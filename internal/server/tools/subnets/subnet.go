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

type Subnet struct{}

func (Subnet) Register(mcpServer *server.MCPServer) {
	mcpTools := []tools.MCPTool{
		ReadSubnet{},
		UpdateSubnet{},
		DeleteSubnet{},
		SubnetIPAddresses{},
		SubnetReservedIPRanges{},
		SubnetStatistics{},
		SubnetUnreservedIPRanges{},
	}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ReadSubnet struct{}

func (ReadSubnet) Create() mcp.Tool {
	return mcp.NewTool(
		"read-subnet",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet to retrieve."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Read Subnet", true, false, false, true)),
		mcp.WithDescription("Get information about a subnet with the given ID."),
	)
}

func (ReadSubnet) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ReadSubnet] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := "/MAAS/api/2.0/subnets/" + subnetID + "/"

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ReadSubnet] Retrieving subnet with ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to read subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[ReadSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ReadSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type UpdateSubnet struct{}

func (UpdateSubnet) Create() mcp.Tool {
	return mcp.NewTool(
		"update-subnet",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet to update."),
		),
		mcp.WithString(
			"cidr",
			mcp.Description("The network CIDR for this subnet."),
		),
		mcp.WithString(
			"name",
			mcp.Description("The subnet's name."),
		),
		mcp.WithString(
			"description",
			mcp.Description("The subnet's description."),
		),
		mcp.WithString(
			"vlan",
			mcp.Description("VLAN this subnet belongs to."),
		),
		mcp.WithString(
			"fabric",
			mcp.Description("Fabric for the subnet."),
		),
		mcp.WithString(
			"vid",
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("VID of the VLAN this subnet belongs to."),
		),
		mcp.WithString(
			"gateway_ip",
			mcp.Description("The gateway IP address for this subnet."),
		),
		mcp.WithString(
			"dns_servers",
			mcp.Description("Comma-separated list of DNS servers for this subnet."),
		),
		mcp.WithString(
			"disabled_boot_architectures",
			mcp.Description("Comma or space separated list of boot architectures which will not be responded to by isc-dhcpd."),
		),
		mcp.WithBoolean(
			"managed",
			mcp.Description("Whether MAAS manages this subnet."),
		),
		mcp.WithBoolean(
			"allow_dns",
			mcp.Description("Configure MAAS DNS to allow DNS resolution from this subnet."),
		),
		mcp.WithBoolean(
			"allow_proxy",
			mcp.Description("Configure maas-proxy to allow requests from this subnet."),
		),
		mcp.WithString(
			"rdns_mode",
			mcp.Enum("0", "1", "2"),
			mcp.Description("How reverse DNS is handled: 0=Disabled, 1=Enabled, 2=RFC2317."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Update Subnet", false, false, false, true)),
		mcp.WithDescription("Update a subnet with the given ID."),
	)
}

func (UpdateSubnet) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[UpdateSubnet] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)

	// Optional parameters - only add if provided
	if cidr := request.GetString("cidr", ""); cidr != "" {
		form.Add("cidr", cidr)
	}
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
	if gatewayIP := request.GetString("gateway_ip", ""); gatewayIP != "" {
		form.Add("gateway_ip", gatewayIP)
	}
	if dnsServers := request.GetString("dns_servers", ""); dnsServers != "" {
		form.Add("dns_servers", dnsServers)
	}
	if disabledBootArchs := request.GetString("disabled_boot_architectures", ""); disabledBootArchs != "" {
		form.Add("disabled_boot_architectures", disabledBootArchs)
	}
	if rdnsMode := request.GetString("rdns_mode", ""); rdnsMode != "" {
		form.Add("rdns_mode", rdnsMode)
	}

	managed := request.GetBool("managed", false)
	if managed {
		form.Add("managed", "1")
	} else {
		form.Add("managed", "0")
	}

	allowDNS := request.GetBool("allow_dns", false)
	if allowDNS {
		form.Add("allow_dns", "1")
	} else {
		form.Add("allow_dns", "0")
	}

	allowProxy := request.GetBool("allow_proxy", false)
	if allowProxy {
		form.Add("allow_proxy", "1")
	} else {
		form.Add("allow_proxy", "0")
	}

	path := "/MAAS/api/2.0/subnets/" + subnetID + "/"

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[UpdateSubnet] Updating subnet with ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypePut, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to update subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[UpdateSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[UpdateSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type DeleteSubnet struct{}

func (DeleteSubnet) Create() mcp.Tool {
	return mcp.NewTool(
		"delete-subnet",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet to delete."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Delete Subnet", false, true, false, true)),
		mcp.WithDescription("Delete a subnet with the given ID."),
	)
}

func (DeleteSubnet) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeleteSubnet] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := "/MAAS/api/2.0/subnets/" + subnetID + "/"

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[DeleteSubnet] Deleting subnet with ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeDelete, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to delete subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[DeleteSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[DeleteSubnet] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type SubnetIPAddresses struct{}

func (SubnetIPAddresses) Create() mcp.Tool {
	return mcp.NewTool(
		"subnet-ip-addresses",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet."),
		),
		mcp.WithBoolean(
			"with_username",
			mcp.DefaultBool(true),
			mcp.Description("If false, suppresses the display of usernames associated with each address."),
		),
		mcp.WithBoolean(
			"with_summary",
			mcp.DefaultBool(true),
			mcp.Description("If false, suppresses the display of nodes, BMCs, and DNS records associated with each address."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Subnet IP Addresses", true, false, false, true)),
		mcp.WithDescription("Returns a summary of IP addresses assigned to this subnet."),
	)
}

func (SubnetIPAddresses) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[SubnetIPAddresses] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	withUsername := request.GetBool("with_username", true)
	withSummary := request.GetBool("with_summary", true)

	path := fmt.Sprintf("/MAAS/api/2.0/subnets/%s/op-ip_addresses?with_username=%d&with_summary=%d",
		subnetID,
		boolToInt(withUsername),
		boolToInt(withSummary))

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[SubnetIPAddresses] Retrieving IP addresses for subnet ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to get IP addresses for subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[SubnetIPAddresses] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[SubnetIPAddresses] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type SubnetReservedIPRanges struct{}

func (SubnetReservedIPRanges) Create() mcp.Tool {
	return mcp.NewTool(
		"subnet-reserved-ip-ranges",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Subnet Reserved IP Ranges", true, false, false, true)),
		mcp.WithDescription("Lists IP ranges currently reserved in the subnet."),
	)
}

func (SubnetReservedIPRanges) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[SubnetReservedIPRanges] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := "/MAAS/api/2.0/subnets/" + subnetID + "/op-reserved_ip_ranges"

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[SubnetReservedIPRanges] Retrieving reserved IP ranges for subnet ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to get reserved IP ranges for subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[SubnetReservedIPRanges] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[SubnetReservedIPRanges] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type SubnetStatistics struct{}

func (SubnetStatistics) Create() mcp.Tool {
	return mcp.NewTool(
		"subnet-statistics",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet."),
		),
		mcp.WithBoolean(
			"include_ranges",
			mcp.DefaultBool(false),
			mcp.Description("If true, includes detailed information about the usage of this range."),
		),
		mcp.WithBoolean(
			"include_suggestions",
			mcp.DefaultBool(false),
			mcp.Description("If true, includes the suggested gateway and dynamic range for this subnet."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Subnet Statistics", true, false, false, true)),
		mcp.WithDescription("Returns statistics for the specified subnet, including usage and availability information."),
	)
}

func (SubnetStatistics) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[SubnetStatistics] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	includeRanges := request.GetBool("include_ranges", false)
	includeSuggestions := request.GetBool("include_suggestions", false)

	path := fmt.Sprintf("/MAAS/api/2.0/subnets/%s/op-statistics?include_ranges=%d&include_suggestions=%d",
		subnetID,
		boolToInt(includeRanges),
		boolToInt(includeSuggestions))

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[SubnetStatistics] Retrieving statistics for subnet ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to get statistics for subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[SubnetStatistics] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[SubnetStatistics] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type SubnetUnreservedIPRanges struct{}

func (SubnetUnreservedIPRanges) Create() mcp.Tool {
	return mcp.NewTool(
		"subnet-unreserved-ip-ranges",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The ID of the subnet."),
		),
		mcp.WithToolAnnotation(tools.CreateToolAnnotation("Subnet Unreserved IP Ranges", true, false, false, true)),
		mcp.WithDescription("Lists IP ranges currently unreserved in the subnet."),
	)
}

func (SubnetUnreservedIPRanges) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	subnetID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[SubnetUnreservedIPRanges] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := "/MAAS/api/2.0/subnets/" + subnetID + "/op-unreserved_ip_ranges"

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[SubnetUnreservedIPRanges] Retrieving unreserved IP ranges for subnet ID: %s", subnetID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to get unreserved IP ranges for subnet %s err=%v", subnetID, err)
		zap.L().Error(fmt.Sprintf("[SubnetUnreservedIPRanges] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[SubnetUnreservedIPRanges] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
