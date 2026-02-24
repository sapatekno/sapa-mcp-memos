# Memos MCP Server (SAPA)

Server Model Context Protocol (MCP) untuk [Memos](https://usememos.com/) yang ditulis dengan Go.
Server ini memungkinkan AI agent membaca, mencari, serta mengelola memo di instance Memos Anda.

## Features
- **create_memo**: Tambah catatan.
- **update_memo**: Edit catatan.
- **delete_memo**: Hapus catatan.
- **read_memo**: Baca catatan (ID atau terbaru).
- **search_memos**: Cari catatan pintar (konten, tag, rentang tanggal).

## Installation

### Prerequisites
- Go 1.23+

### Build
```bash
go build -o memos-mcp.exe ./cmd/server
```

## Configuration

Server dapat dikonfigurasi lewat args key-value pada MCP client, flags, atau environment variables.

| Parameter | Flag | Env Var | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Enable MCP** | `--mcp-enabled` | `MCP_ENABLED` | `true` | Set ke `false` untuk menonaktifkan server. |
| **Memos URL** | `--memos-url` | `MEMOS_URL` | `http://localhost:5230` | URL instance Memos. |
| **API Token** | `--memos-token` | `MEMOS_TOKEN` | `""` | Memos API Bearer Token. |
| **ReadOnly** | (args) `readonly` | - | `false` | Jika `true`, semua aksi tambah/edit/hapus ditolak. |
| **Disabled** | (args) `disabled` | - | `false` | Jika `true`, MCP server tidak akan berjalan. |
| **Need Approve** | (args) `needApprove` | - | `false` | Jika `true`, aksi mutasi perlu `confirm=true`. |

## Usage

### Run as MCP Server
Jalankan binary. Komunikasi melalui Stdio.
```bash
./memos-mcp.exe --memos-url "https://domainmemos.com" --memos-token "memos_pat_****************"
```

### With Claude Desktop / Trae
Tambahkan pada file konfigurasi MCP Anda:

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

## Contoh Pemanggilan Tools

### Baca catatan terbaru
```json
{
  "last": true
}
```

### Baca catatan berdasarkan ID
```json
{
  "id": "memos/5mggyKHrZvL9mNijQbZ8uG"
}
```

### Tambah catatan (needApprove=true)
```json
{
  "content": "Catatan baru",
  "visibility": "PRIVATE",
  "confirm": true
}
```

### Edit catatan
```json
{
  "id": "memos/5mggyKHrZvL9mNijQbZ8uG",
  "content": "Isi diperbarui",
  "confirm": true
}
```

### Hapus catatan
```json
{
  "id": "memos/5mggyKHrZvL9mNijQbZ8uG",
  "confirm": true
}
```

### Cari catatan pintar
```json
{
  "query": "laporan",
  "tags": ["projek", "pilot"],
  "dateFrom": "2026-02-01",
  "dateTo": "2026-02-28"
}
```
