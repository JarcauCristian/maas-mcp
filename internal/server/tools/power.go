package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JarcauCristian/ztp-mcp/internal/server/maas_client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type Power struct{}

func (Power) Register(mcpServer *server.MCPServer) {
	mcpTools := []MCPTool{PowerState{}, ChangePowerState{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type PowerState struct{}

func (PowerState) Create() mcp.Tool {
	return mcp.NewTool(
		"power-state",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to retrieve information for."),
		),
		mcp.WithDescription("Returns the power state of a particular machine."),
	)
}

func (PowerState) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	machineID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[PowerState] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/machines/%s/op-query_power_state", machineID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[PowerState] Retrieving power state for machine with id %s...", machineID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve power state for machine with id %s err=%v", machineID, err)
		zap.L().Error(fmt.Sprintf("[PowerState] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[PowerState] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type ChangePowerState struct{}

func (ChangePowerState) Create() mcp.Tool {
	return mcp.NewTool(
		"change-power-state",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to retrieve information for."),
		),
		mcp.WithBoolean(
			"state",
			mcp.Required(),
			mcp.Description("If true power on the machine else power off."),
		),
		mcp.WithDescription("Change the power state of a machine specified by id."),
	)
}

func (ChangePowerState) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	machineID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ChangePowerState] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	state, err := request.RequireBool("state")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ChangePowerState] Required parameter state not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	var path string

	if state {
		path = fmt.Sprintf("/MAAS/api/2.0/machines/%s/op-power_on", machineID)
	} else {
		path = fmt.Sprintf("/MAAS/api/2.0/machines/%s/op-power_off", machineID)
	}

	client := maas_client.MustClient()

	powerName := "on"

	if !state {
		powerName = "off"
	}

	zap.L().Info(fmt.Sprintf("[ChangePowerState] Power machine with id %s %s...", machineID, powerName))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to power %s machine with id %s err=%v", powerName, machineID, err)
		zap.L().Error(fmt.Sprintf("[ChangePowerState] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[ChangePowerState] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
