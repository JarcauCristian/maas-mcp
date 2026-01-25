package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/JarcauCristian/ztp-mcp/internal/server/maas_client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

const NUMBER_PATTERN = "^[0-9]+$"

type VMHosts struct{}

func (VMHosts) Register(mcpServer *server.MCPServer) {
	mcpTools := []MCPTool{ListVMHosts{}, ListVMHost{}, ListVirtualMachines{}, ComposeVM{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListVMHosts struct{}

func (ListVMHosts) Create() mcp.Tool {
	return mcp.NewTool(
		"list_vm_hosts",
		mcp.WithDescription("Returns the available VM hosts from the ZTP agent conected."),
	)
}

func (ListVMHosts) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	path := "/MAAS/api/2.0/vm-hosts/"

	client := maas_client.MustClient()

	zap.L().Info("[ListVMHosts] Retrieving all VM hosts...")
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve the VM hosts: %v", err)
		zap.L().Error(fmt.Sprintf("[ListVMHosts] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListVMHosts] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ListVMHost struct{}

func (ListVMHost) Create() mcp.Tool {
	return mcp.NewTool(
		"list_vm_host",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Description("The ID of the VM host to query information for."),
			mcp.Pattern(NUMBER_PATTERN),
		),
		mcp.WithDescription("Returns information about a particular VM host specified by id on the ZTP agent conected."),
	)
}

func (ListVMHost) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	vmID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ListVMHost] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/vm-hosts/%s/", vmID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ListVMHost] Retrieving VM host with ID %s...", vmID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retreive VM host with ID %s, err=%v", vmID, err)
		zap.L().Error(fmt.Sprintf("[ListVMHost] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListVMHost] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ComposeVM struct{}

func (ComposeVM) Create() mcp.Tool {
	return mcp.NewTool(
		"compose_vm_host",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Description("ID of the VM host to compose the machine on."),
			mcp.Pattern(NUMBER_PATTERN),
		),
		mcp.WithString(
			"cores",
			mcp.Required(),
			mcp.Description("The number of cores the composed VM should have."),
			mcp.Pattern(NUMBER_PATTERN),
		),
		mcp.WithString(
			"memory",
			mcp.Required(),
			mcp.Description("How much RAM the composed VM should have (Should be in MiB)."),
			mcp.Pattern(NUMBER_PATTERN),
		),
		mcp.WithString(
			"storage",
			mcp.Required(),
			mcp.Description("How much storage the composed VM should have (Should be in GB)."),
			mcp.Pattern(NUMBER_PATTERN),
		),
		mcp.WithString(
			"hostname",
			mcp.Required(),
			mcp.Description("The name of the created VM (Give something random if not provided)."),
			mcp.Pattern("^[a-zA-Z0-9.-]+$"),
		),
		mcp.WithDescription("Compose a VM on a particular VM host specified by ID."),
	)
}

func (ComposeVM) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	vmHostID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ComposeVM] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	cores, err := request.RequireString("cores")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ComposeVM] Required parameter cores not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	memory, err := request.RequireString("memory")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ComposeVM] Required parameter memory not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	storage, err := request.RequireString("storage")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ComposeVM] Required parameter storage not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	hostname, err := request.RequireString("hostname")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ComposeVM] Required parameter hostname not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	form := make(url.Values)
	form.Set("cores", cores)
	form.Set("memory", memory)
	form.Set("storage", storage)
	form.Set("hostname", hostname)

	path := fmt.Sprintf("/MAAS/api/2.0/vm-hosts/%s/op-compose", vmHostID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ComposeVM] Composing VM on host %s with the following configuration:\nCores: %s\nMemory: %s\nStorage: %s\nHostname: %s", vmHostID, cores, memory, storage, hostname))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to compose VM err=%v", err)
		zap.L().Error(fmt.Sprintf("[ComposeVM] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ComposeVM] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ListVirtualMachines struct{}

func (ListVirtualMachines) Create() mcp.Tool {
	return mcp.NewTool(
		"list-virtual-machines",
		mcp.WithString(
			"vm-host-id",
			mcp.Pattern("^[0-9]+$"),
			mcp.Description("The id of the VM host to search the VMs for."),
		),
		mcp.WithBoolean(
			"all",
			mcp.DefaultBool(true),
			mcp.Description("If true return all virtual machines and ignore vm-host-id."),
		),
		mcp.WithDescription("Retrieve all the virtual machines from a specified VM host or all of them."),
	)
}

func (l ListVirtualMachines) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var path string
	var machines []map[string]any
	vmHostId := request.GetString("vm-host-id", "")
	all := request.GetBool("all", true)

	if !all && vmHostId == "" {
		return mcp.NewToolResultError("vm-host-id is required when all=false"), nil
	}

	client := maas_client.MustClient()
	path = "/MAAS/api/2.0/machines/"

	machinesRaw, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve machines from the MAAS host; err=%s", err.Error())), nil
	}

	err = json.Unmarshal([]byte(machinesRaw), &machines)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to unmarshal machines result; err=%s", err.Error())), nil
	}

	virtualMachines := l.parseMachineList(machines)
	var output any
	if all {
		output = virtualMachines
	} else {
		if vms, ok := virtualMachines[vmHostId]; ok {
			output = vms
		} else {
			output = []map[string]any{}
		}
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal output: %v", err)), nil
	}

	return mcp.NewToolResultText(string(outputBytes)), nil
}

func (ListVirtualMachines) parseMachineList(machines []map[string]any) map[string][]map[string]any {
	output := make(map[string][]map[string]any)
	for _, machine := range machines {
		vmData, podId, ok := extractVMData(machine)
		if !ok {
			continue
		}

		output[podId] = append(output[podId], vmData)
	}

	return output
}

func extractVMData(machine map[string]any) (vmData map[string]any, podId string, ok bool) {
	vmId, ok := machine["virtualmachine_id"]
	if !ok {
		return nil, "", false
	}

	systemId, ok := machine["system_id"]
	if !ok {
		return nil, "", false
	}

	hostname, ok := machine["hostname"]
	if !ok {
		return nil, "", false
	}

	podRaw, ok := machine["pod"]
	if !ok {
		return nil, "", false
	}

	pod, ok := podRaw.(map[string]any)
	if !ok {
		return nil, "", false
	}

	podIdRaw, ok := pod["id"]
	if !ok {
		return nil, "", false
	}

	podId = fmt.Sprintf("%v", podIdRaw)

	vmData = map[string]any{
		"hostname":          hostname,
		"virtualmachine_id": vmId,
		"system_id":         systemId,
	}

	return vmData, podId, true
}
