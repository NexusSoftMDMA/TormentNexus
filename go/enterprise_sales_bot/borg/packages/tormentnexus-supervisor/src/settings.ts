import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { z } from 'zod';

export const DEFAULT_ACTION_LABELS = [
    'Run',
    'Expand',
    'Always Allow',
    'Retry',
    'Accept all',
    'Accept',
    'Allow',
    'Approve',
    'Proceed',
    'Keep',
    'Accept all changes',
    'Accept All Changes',
    'Accept All',
    'Approve All',
    'Run command',
    'Allow all'
] as const satisfies readonly string[];

const supervisorSettingsSchema = z.object({
    bumpText: z.string().default('keep going'),
    bumpSentences: z.array(z.string()).default([
        'keep going',
        'proceed',
        'outstanding',
        'perfect',
        'onward',
        'continue',
        'great work, keep it up',
        'excellent, please proceed',
        'magnificent, continue',
        'onward ho!'
    ]),
    actionLabels: z.array(z.string()).default([...DEFAULT_ACTION_LABELS]),
    focusDelayMs: z.number().int().nonnegative().default(100),
    afterClickDelayMs: z.number().int().nonnegative().default(150),
    inputSettleDelayMs: z.number().int().nonnegative().default(120)
});

export type SupervisorSettings = z.infer<typeof supervisorSettingsSchema>;
export type SupervisorSettingsUpdate = Partial<SupervisorSettings>;

export const DEFAULT_SETTINGS = supervisorSettingsSchema.parse({});

export class SupervisorSettingsManager {
    private readonly settingsPath: string;

    constructor(settingsPath?: string) {
        this.settingsPath = settingsPath ?? path.join(os.homedir(), '.tormentnexus', 'supervisor-settings.json');
    }

    getSettingsPath(): string {
        return this.settingsPath;
    }

    async getSettings(): Promise<SupervisorSettings> {
        try {
            const raw = await fs.readFile(this.settingsPath, 'utf-8');
            const parsed = JSON.parse(raw) as unknown;
            const parsedObject = isRecord(parsed) ? parsed : {};
            return supervisorSettingsSchema.parse({ ...DEFAULT_SETTINGS, ...parsedObject });
        } catch (error: unknown) {
            if (isFileNotFound(error)) {
                return DEFAULT_SETTINGS;
            }

            throw error;
        }
    }

    async updateSettings(update: SupervisorSettingsUpdate): Promise<SupervisorSettings> {
        const current = await this.getSettings();
        const next = supervisorSettingsSchema.parse({
            ...current,
            ...update
        });

        await fs.mkdir(path.dirname(this.settingsPath), { recursive: true });
        await fs.writeFile(this.settingsPath, JSON.stringify(next, null, 2), 'utf-8');
        return next;
    }
}

function isFileNotFound(error: unknown): error is NodeJS.ErrnoException {
    return error !== null && typeof error === 'object' && 'code' in error && error.code === 'ENOENT';
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return value !== null && typeof value === 'object' && !Array.isArray(value);
}
