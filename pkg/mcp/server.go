package mcp

import (
	"context"
	"fmt"
	"sapa-memos-mcp/pkg/memos"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Options struct {
	ReadOnly    bool
	NeedApprove bool
}

func NewServer(client *memos.Client, options Options) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "sapa-memos-mcp",
		Version: "0.1.0",
	}, &mcp.ServerOptions{})

	createMemoTool := &mcp.Tool{
		Name:        "create_memo",
		Description: "Create a new memo in Memos",
	}

	type CreateMemoArgs struct {
		Content    string `json:"content" jsonschema:"The content of the memo"`
		Visibility string `json:"visibility" jsonschema:"PUBLIC or PRIVATE"`
		Confirm    bool   `json:"confirm" jsonschema:"Confirm the mutation when needApprove is true"`
	}

	mcp.AddTool(s, createMemoTool, func(ctx context.Context, req *mcp.CallToolRequest, args CreateMemoArgs) (*mcp.CallToolResult, any, error) {
		if res := mutationGuard(options, args.Confirm); res != nil {
			return res, nil, nil
		}
		content := strings.TrimSpace(args.Content)
		if content == "" {
			return errorResult("content is required"), nil, nil
		}
		var visibility *string
		if v := strings.TrimSpace(args.Visibility); v != "" {
			visibility = &v
		}
		memo, err := client.CreateMemo(content, visibility)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to create memo: %v", err)), nil, nil
		}
		idText := memo.Name
		if idText == "" {
			idText = fmt.Sprintf("%d", memo.ID)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Memo created successfully (ID: %s)", idText)},
			},
		}, nil, nil
	})

	updateMemoTool := &mcp.Tool{
		Name:        "update_memo",
		Description: "Update a memo in Memos",
	}

	type UpdateMemoArgs struct {
		ID         string  `json:"id" jsonschema:"Memo ID or name"`
		Content    *string `json:"content" jsonschema:"Updated content"`
		Visibility *string `json:"visibility" jsonschema:"PUBLIC or PRIVATE"`
		Confirm    bool    `json:"confirm" jsonschema:"Confirm the mutation when needApprove is true"`
	}

	mcp.AddTool(s, updateMemoTool, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateMemoArgs) (*mcp.CallToolResult, any, error) {
		if res := mutationGuard(options, args.Confirm); res != nil {
			return res, nil, nil
		}
		if strings.TrimSpace(args.ID) == "" {
			return errorResult("id is required"), nil, nil
		}
		if args.Content == nil && args.Visibility == nil {
			return errorResult("content or visibility is required"), nil, nil
		}
		if args.Content != nil {
			trimmed := strings.TrimSpace(*args.Content)
			if trimmed == "" {
				return errorResult("content cannot be empty"), nil, nil
			}
			args.Content = &trimmed
		}
		if args.Visibility != nil {
			trimmed := strings.TrimSpace(*args.Visibility)
			if trimmed == "" {
				args.Visibility = nil
			} else {
				args.Visibility = &trimmed
			}
		}
		memo, err := client.UpdateMemo(args.ID, args.Content, args.Visibility)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to update memo: %v", err)), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: formatMemo(memo)},
			},
		}, nil, nil
	})

	deleteMemoTool := &mcp.Tool{
		Name:        "delete_memo",
		Description: "Delete a memo in Memos",
	}

	type DeleteMemoArgs struct {
		ID      string `json:"id" jsonschema:"Memo ID or name"`
		Confirm bool   `json:"confirm" jsonschema:"Confirm the mutation when needApprove is true"`
	}

	mcp.AddTool(s, deleteMemoTool, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteMemoArgs) (*mcp.CallToolResult, any, error) {
		if res := mutationGuard(options, args.Confirm); res != nil {
			return res, nil, nil
		}
		if strings.TrimSpace(args.ID) == "" {
			return errorResult("id is required"), nil, nil
		}
		if err := client.DeleteMemo(args.ID); err != nil {
			return errorResult(fmt.Sprintf("Failed to delete memo: %v", err)), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Memo deleted successfully (ID: %s)", args.ID)},
			},
		}, nil, nil
	})

	readMemoTool := &mcp.Tool{
		Name:        "read_memo",
		Description: "Read a memo by ID or get the latest memo",
	}

	type ReadMemoArgs struct {
		ID   string `json:"id" jsonschema:"Memo ID or name"`
		Last bool   `json:"last" jsonschema:"Read the latest memo when true"`
	}

	mcp.AddTool(s, readMemoTool, func(ctx context.Context, req *mcp.CallToolRequest, args ReadMemoArgs) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(args.ID) == "" && !args.Last {
			return errorResult("id or last is required"), nil, nil
		}
		if args.Last {
			memosList, err := client.ListMemos()
			if err != nil {
				return errorResult(fmt.Sprintf("Failed to read memo: %v", err)), nil, nil
			}
			if len(memosList) == 0 {
				return errorResult("No memos available"), nil, nil
			}
			latest := memosList[0]
			for _, memo := range memosList[1:] {
				if memos.MemoCreatedTs(memo) > memos.MemoCreatedTs(latest) {
					latest = memo
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: formatMemo(&latest)},
				},
			}, nil, nil
		}
		memo, err := client.GetMemo(args.ID)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to read memo: %v", err)), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: formatMemo(memo)},
			},
		}, nil, nil
	})

	searchMemosTool := &mcp.Tool{
		Name:        "search_memos",
		Description: "Search for memos by content",
	}

	type SearchMemosArgs struct {
		Query    string   `json:"query" jsonschema:"Search query"`
		Tags     []string `json:"tags" jsonschema:"Tags without #"`
		DateFrom string   `json:"dateFrom" jsonschema:"Start date YYYY-MM-DD"`
		DateTo   string   `json:"dateTo" jsonschema:"End date YYYY-MM-DD"`
	}

	mcp.AddTool(s, searchMemosTool, func(ctx context.Context, req *mcp.CallToolRequest, args SearchMemosArgs) (*mcp.CallToolResult, any, error) {
		from, to, err := parseDateRange(args.DateFrom, args.DateTo)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		memosList, err := client.SearchMemosSmart(args.Query, args.Tags, from, to)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to search memos: %v", err)), nil, nil
		}

		if len(memosList) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No memos found matching the query."},
				},
			}, nil, nil
		}

		var resultText string
		for _, m := range memosList {
			idText := m.Name
			if idText == "" {
				idText = fmt.Sprintf("%d", m.ID)
			}
			resultText += fmt.Sprintf("- ID: %s\n  Content: %s\n  CreatedTs: %d\n\n", idText, truncate(m.Content, 200), memos.MemoCreatedTs(m))
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: resultText},
			},
		}, nil, nil
	})

	return s
}

func mutationGuard(options Options, confirm bool) *mcp.CallToolResult {
	if options.ReadOnly {
		return errorResult("readonly is active; mutation is disabled")
	}
	if options.NeedApprove && !confirm {
		return errorResult("needApprove is active; confirm=true is required")
	}
	return nil
}

func errorResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

func formatMemo(memo *memos.Memo) string {
	idText := memo.Name
	if idText == "" {
		idText = fmt.Sprintf("%d", memo.ID)
	}
	return fmt.Sprintf("ID: %s\nContent: %s\nCreatedTs: %d\nUpdatedTs: %d\nVisibility: %s", idText, memo.Content, memos.MemoCreatedTs(*memo), memos.MemoUpdatedTs(*memo), memo.Visibility)
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "..."
}

func parseDateRange(dateFrom string, dateTo string) (*time.Time, *time.Time, error) {
	var from *time.Time
	var to *time.Time
	if strings.TrimSpace(dateFrom) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateFrom), time.UTC)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid dateFrom format, expected YYYY-MM-DD")
		}
		from = &parsed
	}
	if strings.TrimSpace(dateTo) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateTo), time.UTC)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid dateTo format, expected YYYY-MM-DD")
		}
		end := parsed.Add(24*time.Hour - time.Second)
		to = &end
	}
	return from, to, nil
}
