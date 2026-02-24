package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"sapa-memos-mcp/pkg/mcp"
	"sapa-memos-mcp/pkg/memos"

	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type kvConfig struct {
	URL            string
	Token          string
	ReadOnly       bool
	Disabled       bool
	NeedApprove    bool
	HasURL         bool
	HasToken       bool
	HasReadOnly    bool
	HasDisabled    bool
	HasNeedApprove bool
}

func parseKVArgs(args []string) kvConfig {
	cfg := kvConfig{}
	for i := 0; i+1 < len(args); i += 2 {
		key := strings.ToLower(strings.TrimSpace(args[i]))
		val := strings.TrimSpace(args[i+1])
		switch key {
		case "url":
			cfg.URL = val
			cfg.HasURL = true
		case "token":
			cfg.Token = val
			cfg.HasToken = true
		case "readonly":
			if parsed, err := strconv.ParseBool(val); err == nil {
				cfg.ReadOnly = parsed
				cfg.HasReadOnly = true
			}
		case "disabled":
			if parsed, err := strconv.ParseBool(val); err == nil {
				cfg.Disabled = parsed
				cfg.HasDisabled = true
			}
		case "needapprove":
			if parsed, err := strconv.ParseBool(val); err == nil {
				cfg.NeedApprove = parsed
				cfg.HasNeedApprove = true
			}
		}
	}
	return cfg
}

func main() {
	// Define flags
	mcpEnabled := flag.Bool("mcp-enabled", true, "Enable MCP server mode")
	memosURL := flag.String("memos-url", "http://localhost:5230", "Memos API URL")
	memosToken := flag.String("memos-token", "", "Memos API Token")

	// Parse flags
	flag.Parse()

	kv := parseKVArgs(flag.Args())

	// Check environment variables if flags are default (optional, but good practice)
	if *memosURL == "http://localhost:5230" {
		if envURL := os.Getenv("MEMOS_URL"); envURL != "" {
			*memosURL = envURL
		}
	}
	if *memosToken == "" {
		if envToken := os.Getenv("MEMOS_TOKEN"); envToken != "" {
			*memosToken = envToken
		}
	}

	// Check MCP Enabled flag/env
	if envMCP := os.Getenv("MCP_ENABLED"); envMCP == "false" {
		*mcpEnabled = false
	}

	readOnly := false
	needApprove := false
	disabled := false

	if kv.HasURL {
		*memosURL = kv.URL
	}
	if kv.HasToken {
		*memosToken = kv.Token
	}
	if kv.HasReadOnly {
		readOnly = kv.ReadOnly
	}
	if kv.HasNeedApprove {
		needApprove = kv.NeedApprove
	}
	if kv.HasDisabled {
		disabled = kv.Disabled
	}

	if disabled {
		fmt.Println("MCP Server disabled via configuration.")
		os.Exit(0)
	}

	if !*mcpEnabled {
		fmt.Println("MCP Server disabled via configuration.")
		os.Exit(0)
	}

	if *memosToken == "" {
		log.Println("Warning: No Memos API token provided. Authentication may fail.")
	}

	// Initialize Memos Client
	client := memos.NewClient(*memosURL, *memosToken)

	// Initialize MCP Server
	server := mcp.NewServer(client, mcp.Options{
		ReadOnly:    readOnly,
		NeedApprove: needApprove,
	})

	// Run Server
	log.Println("Starting MCP Server...")
	// Based on SDK examples, Run is the correct method
	if err := server.Run(context.Background(), &mcp_sdk.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
