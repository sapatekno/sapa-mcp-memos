# Memos MCP Server (SAPA)

A Model Context Protocol (MCP) server for [Memos](https://usememos.com/) written in Go.
This server enables AI agents to read, search, and manage memos in your Memos instance.

## Features
- **create_memo**: Add a memo.
- **update_memo**: Edit a memo.
- **delete_memo**: Delete a memo.
- **read_memo**: Read a memo (by ID or latest).
- **search_memos**: Smart search (content, tags, date range).

## Installation

### Prerequisites
- Go 1.23+

### Build
```bash
go build -o memos-mcp.exe ./cmd/server
```

## Configuration

The server can be configured via key-value args in the MCP client, flags, or environment variables.

| Parameter | Flag | Env Var | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Enable MCP** | `--mcp-enabled` | `MCP_ENABLED` | `true` | Set to `false` to disable the server. |
| **Memos URL** | `--memos-url` | `MEMOS_URL` | `http://localhost:5230` | Memos instance URL. |
| **API Token** | `--memos-token` | `MEMOS_TOKEN` | `""` | Memos API Bearer Token. |
| **ReadOnly** | (args) `readonly` | - | `false` | When `true`, create/edit/delete are blocked. |
| **Disabled** | (args) `disabled` | - | `false` | When `true`, MCP server will not run. |
| **Need Approve** | (args) `needApprove` | - | `false` | When `true`, mutations require `confirm=true`. |

## Usage

### Run as MCP Server
Run the binary. Communication uses Stdio.
```bash
./memos-mcp.exe --memos-url "https://domainmemos.com" --memos-token "memos_pat_****************"
```

### With Claude Desktop / Trae
Add to your MCP configuration file:

```json
{
  "mcpServers": {
    "sapa-memos": {
      "command": "path/memos-mcp.exe",
      "args": [
        "url",
        "https://domainmemos.com",
        "token",
        "memos_pat_****************",
        "readonly",
        "false",
        "disabled",
        "false",
        "needApprove",
        "true"
      ],
      "readonly": false,
      "disabled": false,
      "needApprove": true
    }
  }
}
```

## Tool Examples

### Read latest memo
```json
{
  "last": true
}
```

### Read memo by ID
```json
{
  "id": "memos/5mggyKHrZvL9mNijQbZ8uG"
}
```

### Create memo (needApprove=true)
```json
{
  "content": "New memo",
  "visibility": "PRIVATE",
  "confirm": true
}
```

### Update memo
```json
{
  "id": "memos/5mggyKHrZvL9mNijQbZ8uG",
  "content": "Updated content",
  "confirm": true
}
```

### Delete memo
```json
{
  "id": "memos/5mggyKHrZvL9mNijQbZ8uG",
  "confirm": true
}
```

### Smart search
```json
{
  "query": "report",
  "tags": ["project", "pilot"],
  "dateFrom": "2026-02-01",
  "dateTo": "2026-02-28"
}
```
