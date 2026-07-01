# Agent Integration

gpu-mcp-server works with any MCP-compatible agent. Below are copy-paste configs.

## Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gpu": {
      "command": "/path/to/gpu-mcp-server"
    }
  }
}
```

## Goose

```yaml
extensions:
  gpu-metrics:
    type: stdio
    cmd: /path/to/gpu-mcp-server
```

## Cursor

Add to `.cursor/mcp.json` (project-level) or `~/.cursor/mcp.json` (global):

```json
{
  "mcpServers": {
    "gpu": {
      "type": "stdio",
      "command": "/path/to/gpu-mcp-server"
    }
  }
}
```

## Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "gpu": {
      "command": "/path/to/gpu-mcp-server"
    }
  }
}
```

## How it works

The agent spawns `gpu-mcp-server` as a child process using stdio transport.
There is no network port, no authentication surface — the server runs with
the agent's local process privileges.

The agent discovers the four tools (`list_gpus`, `get_gpu_metrics`,
`get_gpu_processes`, `gpu_summary`) via MCP and calls them as needed.
