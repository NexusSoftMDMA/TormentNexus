import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

async function runTest() {
    console.log("Starting MCP Server Connection Test...");

    const transport = new StdioClientTransport({
        command: "node",
        args: [path.join(__dirname, "packages", "core", "dist", "server-stdio.js")],
        env: {
            ...process.env,
            NODE_ENV: "development",
            TORMENTNEXUS_PORT: "4100"
        }
    });

    const client = new Client(
        { name: "tormentnexus-test-client", version: "1.0.0" },
        { capabilities: {} }
    );

    try {
        console.log("Connecting to MCP server...");
        await client.connect(transport);
        console.log("Connected successfully!");

        console.log("Waiting 30 seconds for server bootstrap to complete...");
        await new Promise(r => setTimeout(r, 30000));

        // 1. List Tools
        console.log("\n--- 1. Listing Tools ---");
        const toolsResult = await client.listTools();
        console.log(`Found ${toolsResult.tools.length} tools.`);
        // Show first 5 tools
        toolsResult.tools.slice(0, 5).forEach(t => console.log(` - ${t.name}: ${t.description?.split('\n')[0]}`));

        // 2. Call Internal Tool: router_status
        console.log("\n--- 2. Calling Internal Tool: router_status ---");
        const statusResult = await client.callTool({
            name: "router_status",
            arguments: {}
        }, undefined, { timeout: 120000 }); // 2 minute timeout
        console.log("Result:", JSON.stringify(statusResult, null, 2));

        // 2.5. Call Native Tool: system_diagnostics
        console.log("\n--- 2.5. Calling Native Tool: system_diagnostics ---");
        try {
            const sysStatusResult = await client.callTool({
                name: "system_diagnostics",
                arguments: {}
            }, undefined, { timeout: 60000 });
            console.log("Result:", JSON.stringify(sysStatusResult, null, 2));
        } catch (e) {
            console.error("system_diagnostics call failed:", e.message || String(e));
        }

        // 3. Call Standard Tool: list_directory
        console.log("\n--- 3. Calling Standard Tool: list_directory ---");
        try {
            const listDirResult = await client.callTool({
                name: "list_directory",
                arguments: { path: "." }
            }, undefined, { timeout: 60000 });
            console.log("listDirResult output:", JSON.stringify(listDirResult, null, 2));
            if (listDirResult.isError) {
                console.log("Error returned by tool.");
            } else if (listDirResult.content && listDirResult.content.length > 0) {
                const text = listDirResult.content[0].text;
                console.log(`Success! Result length: ${text?.length || 0} chars.`);
                console.log("Preview:", (text || "").substring(0, 100) + "...");
            } else {
                console.log("Success! (Empty content array returned)");
            }
        } catch (e) {
            console.log("list_directory call failed:", e.message || String(e));
        }

        // 4. Call Aggregated Tool: windows-mcp__SystemInfo (if available)
        const hasWindowsMcp = toolsResult.tools.some(t => t.name === "windows-mcp__SystemInfo");
        if (hasWindowsMcp) {
            console.log("\n--- 4. Calling Aggregated Tool: windows-mcp__SystemInfo ---");
            const sysInfoResult = await client.callTool({
                name: "windows-mcp__SystemInfo",
                arguments: {}
            });
            console.log("Success! System Info retrieved.");
        } else {
            console.log("\n--- 4. Skipping Aggregated Tool: windows-mcp__SystemInfo (not found) ---");
        }

    } catch (error) {
        console.error("Test failed:", error);
    } finally {
        console.log("\nClosing connection...");
        await transport.close();
        process.exit(0);
    }
}

runTest();
