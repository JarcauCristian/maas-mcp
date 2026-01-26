package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JarcauCristian/ztp-mcp/internal/server/templates"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type Templates struct{}

func (Templates) Register(mcpServer *server.MCPServer) {
	mcpTools := []MCPTool{RetrieveTemplates{}, RetrieveTemplateContents{}, RetrieveTemplateById{}, CreateTemplate{}, DeleteTemplate{}}

	for _, tool := range mcpTools {
		mcpServer.AddTool(tool.Create(), tool.Handle)
	}
}

type RetrieveTemplates struct{}

func (RetrieveTemplates) Create() mcp.Tool {
	return mcp.NewTool(
		"retrieve-templates",
		mcp.WithBoolean(
			"only_ids",
			mcp.DefaultBool(false),
			mcp.Description("If true return only the ids of the templates."),
		),
		mcp.WithDescription("Returns all deployment Cloud-Init templates that are available on the system."),
	)
}

func (RetrieveTemplates) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var jsonData []byte
	var errMsg string
	var err error
	templateStore := templates.MustTemplateStore()
	onlyIDs := request.GetBool("only_ids", false)

	if onlyIDs {
		zap.L().Info("[RetrieveTemplates] Retrieving all template descriptions...")
		descriptions := templateStore.ListDescriptions()

		jsonData, err = json.Marshal(descriptions)
		if err != nil {
			errMsg = fmt.Sprintf("failed to marshal result: %v", err)
			zap.L().Error(fmt.Sprintf("[RetrieveTemplates] %s", errMsg))
			return mcp.NewToolResultError(errMsg), nil
		}
	} else {
		zap.L().Info("[RetrieveTemplates] Retrieving all template IDs...")
		templateIDs := templateStore.ListIDs()
		jsonData, err = json.Marshal(templateIDs)
		if err != nil {
			errMsg = fmt.Sprintf("failed to marshal result: %v", err)
			zap.L().Error(fmt.Sprintf("[RetrieveTemplates] %s", errMsg))
			return mcp.NewToolResultError(errMsg), nil
		}
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type RetrieveTemplateById struct{}

func (RetrieveTemplateById) Create() mcp.Tool {
	return mcp.NewTool(
		"retrieve-template-by-id",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[a-z0-9_-]*$"),
			mcp.Description("The id of the template to retrieve."),
		),
		mcp.WithDescription("Return the information about a particular template specified by ID."),
	)
}

func (RetrieveTemplateById) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	templateStore := templates.MustTemplateStore()

	templateId, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[RetrieveTemplateById] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	zap.L().Info(fmt.Sprintf("[RetrieveTemplateById] Retrieving template with id %s...", templateId))
	descriptions, err := templateStore.GetDescription(templateId)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve description for template with id %s: %v", templateId, err)
		zap.L().Error(fmt.Sprintf("[RetrieveTemplateById] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	jsonData, err := json.Marshal(descriptions)
	if err != nil {
		errMsg = fmt.Sprintf("failed to marshal result: %v", err)
		zap.L().Error(fmt.Sprintf("[RetrieveTemplateById] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

type RetrieveTemplateContents struct{}

func (RetrieveTemplateContents) Create() mcp.Tool {
	return mcp.NewTool(
		"retrieve-template-content",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[a-z0-9_-]*$"),
			mcp.Description("The id of the template to retrieve the contents for."),
		),
		mcp.WithDescription("Return contents of a particular template specified by ID."),
	)
}

func (RetrieveTemplateContents) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var errMsg string

	templateStore := templates.MustTemplateStore()

	templateId, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[RetrieveTemplateContents] Required parameter id not present err=%v", err))
		return mcp.NewToolResultError(err.Error()), nil
	}

	zap.L().Info(fmt.Sprintf("[RetrieveTemplateContents] Retrieving template content for id %s...", templateId))
	templateContent, err := templateStore.GetContent(templateId)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to retrieve template content for id %s: %v", templateId, err)
		zap.L().Error(fmt.Sprintf("[RetrieveTemplateContents] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(templateContent), nil
}

type CreateTemplate struct{}

func (CreateTemplate) Create() mcp.Tool {
	return mcp.NewTool(
		"create-template",
		mcp.WithInputSchema[templates.GenericTemplate](),
		mcp.WithDescription("Create and add a new template based on the html template files required: description.json and template.yaml."),
	)
}

func (CreateTemplate) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templateStore := templates.MustTemplateStore()

	argumentsJSON, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to marshal arguments: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateTemplate] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	var genericTemplate templates.GenericTemplate
	if err := json.Unmarshal(argumentsJSON, &genericTemplate); err != nil {
		errMsg := fmt.Sprintf("Failed to unmarshal arguments to GenericTemplate: %v", err)
		zap.L().Error(fmt.Sprintf("[CreateTemplate] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	zap.L().Info(fmt.Sprintf("[CreateTemplate] Creating template with id: %s", genericTemplate.Id))

	err = templateStore.Create(genericTemplate)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create template for id %s: %v", genericTemplate.Id, err)
		zap.L().Error(fmt.Sprintf("[CreateTemplate] %s", errMsg))
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created the template with id=%s", genericTemplate.Id)), nil
}

type DeleteTemplate struct{}

func (DeleteTemplate) Create() mcp.Tool {
	return mcp.NewTool(
		"delete-template",
		mcp.WithString(
			"id",
			mcp.Required(),
			mcp.Pattern("^[0-9a-z_-]+$"),
			mcp.Description("The id of the template to be deleted."),
		),
		mcp.WithDescription("Delete the templated specified by the id."),
	)
}

func (DeleteTemplate) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templateStore := templates.MustTemplateStore()
	templateId, err := request.RequireString("id")
	if err != nil {
		zap.L().Error(fmt.Sprintf("[Delete Template] %s", err.Error()))
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = templateStore.Delete(templateId)
	if err != nil {
		zap.L().Error(fmt.Sprintf("[Delete Template] %s", err.Error()))
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully delete template with id: %s", templateId)), nil
}
