package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/JarcauCristian/ztp-mcp/internal/server/maas_client"
	"github.com/JarcauCristian/ztp-mcp/internal/server/parser"
	"github.com/JarcauCristian/ztp-mcp/internal/server/templates"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type Machines struct{}

func (Machines) Register(mcpServer *server.MCPServer) {
	mcpTools := []MCPTool{ListMachines{}, ListMachine{}, CommissionMachine{}, DeployMachine{}, TestMachine{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type ListMachines struct{}

func (ListMachines) Create() mcp.Tool {
	return mcp.NewTool(
		"list_machines",
		mcp.WithString(
			"status",
			mcp.Enum(
				"new",
				"commissioning",
				"failed_commissioning",
				"ready",
				"deploying",
				"deployed",
				"releasing",
				"failed_deployment",
				"allocated",
				"retired",
				"broken",
				"recommissioning",
				"testing",
				"failed_testing",
				"rescuing",
				"disk_erasing",
				"failed_disk_erasing",
			),
			mcp.Description("The status of the machine that will be retrieved. Returns all machines if not provided."),
		),
		mcp.WithDescription("List all the available machines on the current ZTP agent conected."),
	)
}

func (ListMachines) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var path, errMsg string

	status := request.GetString("status", "")

	if status == "" {
		path = "/MAAS/api/2.0/machines/"
	} else {
		path = fmt.Sprintf("/MAAS/api/2.0/machines/?status=%s", status)
	}

	client := maas_client.MustClient()

	zap.L().Info("[ListMachines] Retrieving all the machines...")
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve all the machines: %v", err)
		zap.L().Error(fmt.Sprintf("[ListMachines] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	var machines []map[string]any

	err = json.Unmarshal([]byte(resultData), &machines)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to unmarshal the result err=%v", err)
		zap.L().Error(fmt.Sprintf("[ListMachines] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	// Filter out machines with the "protected" tag
	var filteredMachines []map[string]any
	for _, machine := range machines {
		if !parser.CheckForProtectedTag(machine) {
			filteredMachines = append(filteredMachines, machine)
		}
	}

	response, err := json.Marshal(filteredMachines)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to marshal filtered machines: %v", err)
		zap.L().Error(fmt.Sprintf("[ListMachines] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(response)), nil
}

type ListMachine struct{}

func (ListMachine) Create() mcp.Tool {
	return mcp.NewTool(
		"list_machine",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to retrieve information for."),
		),
		mcp.WithDescription("Return the information about a particular machine."),
	)
}

func (ListMachine) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string
	jsonData := make([]byte, 0)

	machineID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[ListMachine] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/machines/%s/", machineID)

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[ListMachine] Retrieving machine with id %s...", machineID))
	resultData, err := client.Do(ctx, maas_client.RequestTypeGet, path, nil)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve the machine with id %s err=%v", machineID, err)
		zap.L().Error(fmt.Sprintf("[ListMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	var machine map[string]any
	err = json.Unmarshal([]byte(resultData), &machine)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to unmarshal the result: %v", err)
		zap.L().Error(fmt.Sprintf("[ListMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	if !parser.CheckForProtectedTag(machine) {
		jsonData, err = json.Marshal(resultData)
		if err != nil {
			errMsg = fmt.Sprintf("failed to marshal result: %v", err)
			zap.L().Error(fmt.Sprintf("[ListMachine] %s", errMsg))
			return mcp.NewToolResultError(errMsg), nil
		}
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type CommissionMachine struct{}

func (CommissionMachine) Create() mcp.Tool {
	return mcp.NewTool(
		"commission_machine",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to commission."),
		),
		mcp.WithDescription("Start the commissioning process on a particular machine."),
	)
}

func (CommissionMachine) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	machineID, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[CommissionMachine] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/machines/%s/op-commission", machineID)

	client := maas_client.MustClient()

	form := make(url.Values)
	form.Add("enable_ssh", "1")

	zap.L().Info(fmt.Sprintf("[CommissionMachine] Commissioning machine with id %s...", machineID))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to commission the machine with id %s err=%v", machineID, err)
		zap.L().Error(fmt.Sprintf("[CommissionMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[CommissionMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type DeployMachine struct{}

func (DeployMachine) Create() mcp.Tool {
	return mcp.NewTool(
		"deploy_machine",
		mcp.WithString(
			"machineId",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to deploy."),
		),
		mcp.WithString(
			"templateId",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z-_]*$"),
			mcp.Description("The id of the templates to use for deployment."),
		),
		mcp.WithString(
			"templateParameters",
			mcp.Required(),
			mcp.Description("The parameters that will be used to replace the values in the templates. They are represented as a JSON valid object. If the template does not require parameters enter an empty JSON map {}."),
		),
		mcp.WithDescription("Deploys a machine with the specified id and template."),
	)
}

func (DeployMachine) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	machineId, err := request.RequireString("machineId")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeployMachine] Required parameter machineId not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	templateId, err := request.RequireString("templateId")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeployMachine] Required parameter templateId not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	parameters, err := request.RequireString("templateParameters")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeployMachine] Required parameter templateParameters not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	templateExecutor, err := templates.RetrieveExecutor(templateId, parameters)
	if err != nil {
		zap.L().Error(fmt.Sprintf("[DeployMachine] Failed to retrieve the template executor for parameters %s.", parameters))
		return mcp.NewToolResultError(err.Error()), nil
	}

	userData, err := templateExecutor.Execute()
	if err != nil {
		errMsg = "Failed to execute the template to retrieve the userData."
		zap.L().Error(fmt.Sprintf("[DeployMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	client := maas_client.MustClient()

	path := fmt.Sprintf("/MAAS/api/2.0/machines/%s/op-deploy", machineId)

	form := make(url.Values)
	form.Add("user_data", userData)

	zap.L().Info(fmt.Sprintf("[DeployMachine] Deploying machine with id %s and template %s...", machineId, templateId))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to deploy the machine with id %s err=%v", machineId, err)
		zap.L().Error(fmt.Sprintf("[DeployMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[DeployMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type TestMachine struct{}

func (TestMachine) Create() mcp.Tool {
	return mcp.NewTool(
		"test_machine",
		mcp.WithString(
			"system_id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z]{6}$"),
			mcp.Description("The id of the machine to retrieve information for."),
		),
		mcp.WithBoolean(
			"enable_ssh",
			mcp.Description("Whether to enable SSH for the testing environment using the user's SSH key(s). 0 = false. 1 = true."),
		),
		mcp.WithString(
			"parameters",
			mcp.Description("Scripts selected to run may define their own parameters. These parameters may be passed using the parameter name. Optionally a parameter may have the script name prepended to have that parameter only apply to that specific script."),
		),
		mcp.WithString(
			"testing_scripts",
			mcp.Pattern("^[^,]+(?:,[^,]+)*$"),
			mcp.Description("A comma-separated list of testing script names and tags to be run. By default all tests tagged 'commissioning' will be run."),
		),
		mcp.WithToolAnnotation(CreateToolAnnotation("Test Machine", false, true, false, true)),
		mcp.WithDescription("Tool used to test the machine with the specified testing scripts."),
	)
}

func (TestMachine) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	systemID, err := request.RequireString("system_id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[TestMachine] Required parameter system_id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := fmt.Sprintf("/MAAS/api/2.0/machines/%s/op-test", systemID)

	form := make(url.Values)

	enableSSH := request.GetBool("enable_ssh", false)
	if enableSSH {
		form.Add("enable_ssh", "1")
	} else {
		form.Add("enable_ssh", "0")
	}

	if parameters := request.GetString("parameters", ""); parameters != "" {
		form.Add("parameters", parameters)
	}

	if testingScripts := request.GetString("testing_scripts", ""); testingScripts != "" {
		form.Add("testing_scripts", testingScripts)
	}

	client := maas_client.MustClient()

	zap.L().Info(fmt.Sprintf("[TestMachine] Testing machine with system ID %s...", systemID))
	resultData, err := client.Do(ctx, maas_client.RequestTypePost, path, strings.NewReader(form.Encode()))
	if err != nil {
		errMsg = fmt.Sprintf("Failed to test machine with system ID %s err=%v", systemID, err)
		zap.L().Error(fmt.Sprintf("[TestMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(resultData)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[TestMachine] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
