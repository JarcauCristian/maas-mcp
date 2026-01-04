package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/JarcauCristian/ztp-mcp/internal/server/executor"
	"github.com/JarcauCristian/ztp-mcp/internal/server/maas_client"
	"github.com/JarcauCristian/ztp-mcp/internal/server/vault"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Commands struct{}

func (Commands) Register(mcpServer *server.MCPServer) {
	mcpTools := []MCPTool{ExecuteCommand{}, ExecuteScript{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

func getHostname(ctx context.Context, machineId string) (string, error) {
	path := fmt.Sprintf("/MAAS/api/2.0/machines/%s/", machineId)
	client := maas_client.MustClient()
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		return "", err
	}

	var machine map[string]any
	err = json.Unmarshal([]byte(resultData), &machine)

	ipAddrs, ok := machine["ip_addresses"]
	if !ok {
		return "", errors.New("no ip_addresses key found")
	}

	ipAddrsList, ok := ipAddrs.([]string)
	if !ok {
		return "", errors.New("couldn't convert ipAddrs to a slice of strings")
	}

	return ipAddrsList[0], nil
}

type ExecuteCommand struct{}

func (ExecuteCommand) Create() mcp.Tool {
	return mcp.NewTool(
		"execute_command",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to execute the command on."),
		),
		mcp.WithString(
			"command",
			mcp.Required(),
			mcp.Description("The command to execute on the machine."),
		),
		mcp.WithDescription("Execute a single line command on the specified machine."),
	)
}

func (ExecuteCommand) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command, err := request.RequireString("command")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	machineId, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	hostname, err := getHostname(ctx, machineId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	vault := vault.MustVault()
	e := executor.NewExecutor(vault, hostname)
	output, err := e.Execute(ctx, []string{command})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(output), nil
}

type ExecuteScript struct{}

func (ExecuteScript) Create() mcp.Tool {
	return mcp.NewTool(
		"execute_script",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to execute the command on."),
		),
		mcp.WithArray(
			"commands",
			mcp.WithStringItems(),
			mcp.Required(),
			mcp.Description("The command to execute on the machine."),
		),
		mcp.WithDescription("Execute a single line command on the specified machine."),
	)
}

func (ExecuteScript) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command, err := request.RequireStringSlice("command")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	machineId, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	hostname, err := getHostname(ctx, machineId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	vault := vault.MustVault()
	e := executor.NewExecutor(vault, hostname)
	output, err := e.Execute(ctx, command)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(output), nil
}
