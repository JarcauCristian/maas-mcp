package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/JarcauCristian/ztp-mcp/internal/server/middleware"
	"github.com/JarcauCristian/ztp-mcp/internal/server/registry"
	"github.com/JarcauCristian/ztp-mcp/internal/server/tools"
	"github.com/JarcauCristian/ztp-mcp/internal/server/tools/fabrics"
	"github.com/JarcauCristian/ztp-mcp/internal/server/tools/subnets"
	"github.com/JarcauCristian/ztp-mcp/internal/server/tools/tags"
	"github.com/JarcauCristian/ztp-mcp/internal/server/tools/vlans"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mark3labs/mcp-go/server"
)

func init() {
	var logger *zap.Logger

	config := zap.NewDevelopmentConfig()

	config.DisableCaller = false
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	logger = zap.Must(config.Build())

	zap.ReplaceGlobals(logger)
}

func registerTools(mcpServer *server.MCPServer) {
	registries := []registry.Registry{
		tools.VMHosts{},
		tools.Machines{},
		tools.Power{},
		tools.Templates{},
		tags.Tags{},
		tags.Tag{},
		subnets.Subnets{},
		subnets.Subnet{},
		fabrics.Fabrics{},
		fabrics.Fabric{},
		vlans.Vlans{},
		vlans.Vlan{},
	}

	for _, reg := range registries {
		reg.Register(mcpServer)
	}
}

func main() {
	var version string
	info, ok := debug.ReadBuildInfo()

	if !ok {
		version = "0.1.0"
	} else {
		version = info.Main.Version
	}

	mcpTransportRaw := flag.String("mcp-transport", "stdio", "MCP transport to use when strating the server.")
	mcpAddressRaw := flag.String("mcp-address", "localhost:8080", "MCP address in the form of <host>:<port> for SSE and HTTP transport modes.")
	flag.Parse()

	mcpTransport := *mcpTransportRaw
	mcpAddress := *mcpAddressRaw

	mcpServer := server.NewMCPServer(
		"Zero-Touch Provisioning MPC Server",
		version,
		server.WithInstructions("This server is used to communicate with the ZTP agent in order to deploy, interact and retrieve the status of machines inside an Ubuntu MAAS instance."),
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	if err := godotenv.Load(".env"); err != nil {
		zap.L().Warn("Failed to load environment variables from .env. Using the envs in environ...")
	}

	registerTools(mcpServer)

	switch mcpTransport {
	case "SSE", "sse":
		zap.L().Info("Starting MCP server in SSE mode...")
		sseServer := server.NewSSEServer(mcpServer)
		if err := sseServer.Start(mcpAddress); err != nil {
			zap.L().Fatal(err.Error())
		}
	case "HTTP", "http":
		zap.L().Info(fmt.Sprintf("Starting MCP server in Streamable HTTP mode on %s...", mcpAddress))

		mux := http.NewServeMux()

		mux.Handle("/mcp", server.NewStreamableHTTPServer(mcpServer))
		handler := middleware.Logging(middleware.Auth(mux))

		if err := http.ListenAndServe(mcpAddress, handler); err != nil {
			zap.L().Fatal(err.Error())
		}
	case "STDIO", "stdio":
		zap.L().Info("Starting MCP server in stdio mode...")
		if err := server.ServeStdio(mcpServer); err != nil {
			zap.L().Fatal(err.Error())
		}
	}
}
