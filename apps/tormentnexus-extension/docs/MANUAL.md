# TormentNexus Extension User Manual

## Overview

TormentNexus Extension is a Chrome extension that bridges the Model Context Protocol (MCP) with web-based AI platforms like ChatGPT, Claude, Perplexity, and others. It allows you to use your local tools and data directly within these AI interfaces, enhancing their capabilities with file system access, command execution, and more.
# tormentnexus Extension User Manual

## Overview

tormentnexus Extension is a Chrome extension that bridges the Model Context Protocol (MCP) with web-based AI platforms like ChatGPT, Claude, Perplexity, and others. It allows you to use your local tools and data directly within these AI interfaces, enhancing their capabilities with file system access, command execution, and more.

## Getting Started

### Installation

1.  **Install the Extension**: Load the extension in Chrome (Developer Mode) or install from the Chrome Web Store.
2.  **Install the Proxy**: To connect to local MCP servers, you need to run the TormentNexus Extension Proxy.
2.  **Install the Proxy**: To connect to local MCP servers, you need to run the tormentnexus Extension Proxy.

### Proxy Setup

The proxy bridges the browser (extension) to your local MCP servers.

1.  **Create a Configuration File**: Create a `config.json` file defining your MCP servers.

    ```json
    {
      "mcpServers": {
        "filesystem": {
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/folder"],
          "env": {
            "DEBUG": "true"
          }
        }
      }
    }
    ```

2.  **Run the Proxy**:

    ```bash
    npx -y @srbhptl39/tormentnexus-extension-proxy@latest --config ./config.json
    npx -y @srbhptl39/tormentnexus-extension-proxy@latest --config ./config.json
    ```

    By default, this starts an SSE server on port 3006. You can change the port using environment variables: `PORT=3007 npx ...`

## Connection Configuration

Open the extension sidebar and navigate to the **Settings** tab.

### Connection Types

*   **SSE (Server-Sent Events)**: The default connection type. Standard HTTP streaming.
    *   URI: `http://localhost:3006/sse`
*   **WebSocket**: Faster, full-duplex communication. Requires running the proxy with WebSocket support or using a WebSocket-enabled MCP server.
    *   URI: `ws://localhost:3006/message`
*   **Streamable HTTP**: Standard MCP transport over HTTP.
    *   URI: `http://localhost:3006/mcp`

### Server URI

Enter the URL where your local proxy or remote MCP server is running.

### Testing Connection

You can test the latency of your connection by clicking the **Test Server Context** button in the Settings > Network tab. This will measure the round-trip time to your MCP server.

## Features

### Tool Detection
The extension automatically detects when the AI wants to call a tool based on the conversation context. It presents a "Call Tool" card in the chat interface.

### Automation Settings

These settings can be found in the **Settings** tab.

*   **Auto Insert**: Automatically inserts the tool result into the chat input box after the tool finishes execution.
    *   *Delay*: Time in seconds to wait before inserting. Set to 0 for instant insertion.
*   **Auto Submit**: Automatically submits the chat message after the tool result has been inserted.
    *   *Requirement*: Requires 'Auto Insert' to be enabled.
    *   *Delay*: Time in seconds to wait before submitting.
*   **Auto Execute**: Automatically runs the tool when a "Call Tool" card appears, without requiring you to click "Run".
    *   *Warning*: Use with caution. This enables fully autonomous tool execution.

### Push Content Mode
Toggle this in the Sidebar settings (Appearance Tab). It adjusts the page layout so the sidebar pushes the main content aside instead of overlaying it. This is useful for smaller screens to prevent the sidebar from blocking chat content.

### Feature Flags & Experiments
Also inside the Settings tab, you can navigate to the **Advanced** section to view and toggle experimental feature flags. This allows you to opt-in to bleeding-edge UI changes or new protocol handling routes before they are fully released.

### Tool Management
In the **Available Tools** tab, you can:
*   **Search**: Filter tools by name or description.
*   **Favorites**: Click the star icon next to a tool to mark it as a favorite. Toggle sort order to prioritize favorites.
*   **Enable/Disable**: Toggle individual tools or entire server groups.
*   **View Details**: Click on a tool to expand its description and view the JSON schema.

### Resource Browser
In the **Resources** tab, you can browse any raw data or templates exposed by your MCP servers.
*   **List Resources**: View all active `resources` and `resourceTemplates` supplied.
*   **Read Content**: Click on a resource to trigger the `read_resource` protocol and display the parsed text, JSON, or markdown contents directly in the extension.

### Prompt Templates
In the **Prompts** tab, you can execute predefined prompt chains supplied by your MCP servers. Selecting a prompt allows you to fill in any required arguments before instantly dispatching the text into your active chat window.

### Activity Monitoring & Logs
The new **Activity** tab provides a real-time timeline of extension actions:
*   **Log Entries**: Tracks every tool execution, connection event, and error.
*   **Filtering**: Filter logs by type (Tools, Connection, Errors).
*   **Details**: Click on any log entry to view full details, including execution metadata and raw JSON results.
*   **Persistence**: Logs are saved locally (up to 50 entries) so you can review recent history even after reloading the page.

### Notifications & Alerts
The extension provides two notification channels:
*   **Toasts**: Non-intrusive popups for successful connections, tool executions, and settings saves.
*   **Notification Center**: Click the bell icon at the top of the sidebar to view a persistent history of System Alerts and Remote Campaign Notifications. You can dismiss messages individually or clear all at once.

## Advanced Usage

### Developing Custom Tools

You can create your own MCP server to expose custom tools (e.g., internal API access, database queries). Refer to the [Official MCP Documentation](https://modelcontextprotocol.io/docs/server) to learn how to build an MCP server in Python or TypeScript.

Once built, simply add it to your `config.json` and restart the proxy.

### Security Best Practices

*   **Local Execution**: Tools run on your local machine. Be careful with tools that modify files or execute system commands. Always review the tool call before clicking "Run" unless you trust the source completely.
*   **Data Privacy**: Your data (files, database content) remains local. It is only sent to the AI provider (OpenAI, Anthropic, etc.) when a tool result is explicitly inserted into the chat.
*   **API Keys**: Never hardcode API keys in your `config.json` if you plan to share it. Use environment variables instead.

## Troubleshooting

### Connection Refused / 404
*   Ensure your proxy server is running.
*   Check if the port (default 3006) matches the URI in Server Status.
*   Verify `config.json` syntax.

### Tools Not Showing
*   Click the "Refresh" button in the Available Tools tab.
*   Ensure your MCP server is healthy and sending the tool list.

### Extension Context Invalidated
This happens if the extension is updated or reloaded while the page is open. Simply refresh the web page to reconnect.

### Latency Issues
If the "Test Connection" shows high latency (>200ms) for a local server:
*   Check your CPU usage.
*   Ensure no other process is blocking the port.
*   Try switching to WebSocket if supported.

## FAQ

**Q: Is my data secure?**
A: Yes. The extension communicates directly with your local proxy. Your data (files, etc.) stays local unless you explicitly send it to the AI as a tool result.

**Q: Which AI models work best?**
A: Models with strong function calling capabilities (like GPT-4, Claude 3.5 Sonnet) work best.

## Macros & Agentic Mode

### Overview
Macros allow you to automate sequences of tool executions. With "Agentic Mode," you can add conditional logic and loops to create powerful workflows that adapt based on tool outputs.

### Creating a Macro
1.  Go to the **Macros** tab in the sidebar.
2.  Click **New Macro**.
3.  Enter a name and description.
4.  Add steps:
    *   **Tool**: Execute a specific tool with predefined arguments.
    *   **Condition**: Check a condition (JavaScript expression) and branch execution (Continue, Stop, Go to Step).
    *   **Delay**: Wait for a specified duration.

### Agentic Capabilities
*   **Variables**: Access previous results using `{{lastResult}}` in tool arguments.
*   **Conditionals**: Evaluate expressions like `lastResult.status === 'success'` to decide the next step.
*   **Loops**: Use "Go to Step" actions to retry steps or iterate until a condition is met.

### Running Macros
Click the **Play** button on a macro card to execute it. Progress and results are shown in the Activity Log and via toast notifications.

### Macro Management
*   **Import/Export**: Use the 'Export' button in the Macro Builder to save your workflows as JSON files. Use the 'Import' button in the main Macro list to load them. This allows you to share complex automations with others.

### Dashboard Analytics
The Dashboard now provides deeper insights:
*   **Activity Chart**: View your tool usage over the last 7 days.
*   **Quick Access**: Run your most recently updated macros directly from the dashboard.

### Context Manager
Located in the input area (click the book icon), the Context Manager allows you to:
*   **Save Context**: Store frequently used prompts, instructions, or data snippets.
*   **Insert Context**: Quickly inject saved snippets into your current conversation.
*   **Manage**: Edit or delete snippets as your workflow evolves.
