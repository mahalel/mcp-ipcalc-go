# MCP Server IP Calculator

A Go-based MCP (Model Control Protocol) server that provides IP address calculation functionality, similar to the `ipcalc` command-line tool in Linux.

https://github.com/user-attachments/assets/44ece6c5-2fea-4dac-bc5a-07a575a4c950

## Features

- **IP Information**: Get detailed information about an IP address including network, netmask, broadcast, and usable hosts
- **Netmask Details**: Display detailed information about a network mask in decimal, binary, and hex formats

## Installation

### Option 1: Install directly with Go (recommended)

If you have Go installed, you can install the tool directly to your PATH:

```bash
go install github.com/mahalel/mcp-ipcalc-go@latest
```

This will compile and install the binary as `mcp-ipcalc-go` in your Go binary directory (usually `~/go/bin/`).

### Option 2: Manual build

1. Clone the repository:
   ```bash
   git clone https://github.com/mahalel/mcp-ipcalc-go.git
   cd mcp-ipcalc-go
   ```

2. Build the binary:
   ```bash
   go build -o mcp-ipcalc-go
   ```

3. Make the binary executable (Unix systems):
   ```bash
   chmod +x mcp-ipcalc-go
   ```

## Configure Claude Desktop Integration

To use this tool with Claude Desktop, add the following configuration to your `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "ipcalc": {
      "command": "/path/to/your/mcp-ipcalc-go"
    }
  }
}
```

## Usage

Once configured, you can use the IP calculator in Claude Desktop with the following operations:

### Get IP Information

```
@ipcalc operation: "info", ip: "192.168.1.0/24"
```

### Get Netmask Details

```
@ipcalc operation: "netmask", ip: "192.168.1.0/24"
```

Or you can use natural language like:

Info:
```
show me ip info for 192.168.1.0/24
```

Netmask:
```
show me the netmask for 10.0.0.0/8
```

## Development

### Requirements

- Go 1.20 or higher
- The [mcp-go](https://github.com/mark3labs/mcp-go) library

### Running Tests

```bash
go test -v
```

## License

GPLv3

## About MCP

The Model Control Protocol (MCP) allows AI assistants like Claude to interact with external tools. This implementation provides IP calculation capabilities to Claude via MCP.

# TODO

- Refactor the Splitting operation to be like `ipcalc`
