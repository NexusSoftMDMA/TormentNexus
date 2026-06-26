export interface SavedScriptRow {
    uuid: string;
    name: string;
    description: string;
    code: string;
}

function isObject(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null;
}

export function normalizeSavedScripts(payload: unknown): SavedScriptRow[] {
    if (!Array.isArray(payload)) {
        return [];
    }

    return payload.reduce<SavedScriptRow[]>((acc, item, index) => {
        if (!isObject(item)) {
            return acc;
        }

        const rawUuid = typeof item.uuid === 'string' ? item.uuid.trim() : '';
        const rawName = typeof item.name === 'string' ? item.name.trim() : '';
        const rawDescription = typeof item.description === 'string' ? item.description.trim() : '';
        const rawCode = typeof item.code === 'string' ? item.code : '';

        acc.push({
            uuid: rawUuid.length > 0 ? rawUuid : `script-${index}`,
            name: rawName.length > 0 ? rawName : 'Unnamed script',
            description: rawDescription,
            code: rawCode,
        });

        return acc;
    }, []);
}