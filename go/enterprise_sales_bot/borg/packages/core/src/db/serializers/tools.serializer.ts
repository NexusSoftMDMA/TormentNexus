import { DatabaseTool, Tool } from "../../types/mcp-admin/index.js";

export class ToolsSerializer {
    static serializeTool(dbTool: DatabaseTool): Tool {
        return {
            uuid: dbTool.uuid,
            name: dbTool.name,
            description: dbTool.description,
            toolSchema: dbTool.toolSchema,
            is_deferred: dbTool.is_deferred ?? false,
            always_on: dbTool.always_on ?? false,
            created_at: dbTool.created_at.toISOString(),
            updated_at: dbTool.updated_at.toISOString(),
            mcp_server_uuid: dbTool.mcp_server_uuid,
        };
    }

    static serializeToolList(dbTools: DatabaseTool[]): Tool[] {
        return dbTools.map(this.serializeTool);
    }
}
