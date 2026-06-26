import { ApiKeyType } from "../../types/mcp-admin/index.js";

export class ApiKeysSerializer {
    static serializeApiKey(dbApiKey: {
        uuid: string;
        name: string;
        key: string;
        type: ApiKeyType;
        created_at: Date;
        is_active: boolean;
    }) {
        return {
            uuid: dbApiKey.uuid,
            name: dbApiKey.name,
            key: dbApiKey.key,
            type: dbApiKey.type,
            created_at: dbApiKey.created_at,
            is_active: dbApiKey.is_active,
        };
    }

    static serializeApiKeyList(
        dbApiKeys: Array<{
            uuid: string;
            name: string;
            key: string;
            type: ApiKeyType;
            created_at: Date;
            is_active: boolean;
            user_id: string | null;
        }>,
    ) {
        return dbApiKeys.map((apiKey) => ({
            uuid: apiKey.uuid,
            name: apiKey.name,
            key: apiKey.key,
            type: apiKey.type,
            created_at: apiKey.created_at, // Keep as Date if frontend handles it, or serialize to string? Original keeps as Date in list but string in single? No, original keeps Date.
            is_active: apiKey.is_active,
            user_id: apiKey.user_id,
        }));
    }

    static serializeCreateApiKeyResponse(dbApiKey: {
        uuid: string;
        name: string;
        key: string;
        type: ApiKeyType;
        user_id: string | null;
        created_at: Date;
    }) {
        return {
            uuid: dbApiKey.uuid,
            name: dbApiKey.name,
            key: dbApiKey.key,
            type: dbApiKey.type,
            created_at: dbApiKey.created_at,
        };
    }
}
