import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import fs from 'fs';
import path from 'path';
import os from 'os';
import { MCPConfigStore } from './configStore.js';

describe('MCPConfigStore', () => {
    let tempDir: string;
    let configPath: string;
    let lkgPath: string;
    let store: MCPConfigStore;

    beforeEach(() => {
        tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'mcp-config-test-'));
        configPath = path.join(tempDir, 'mcp_servers.json');
        lkgPath = path.join(tempDir, 'mcp_servers.lkg.json');
        store = new MCPConfigStore(configPath);
    });

    afterEach(() => {
        fs.rmSync(tempDir, { recursive: true, force: true });
    });

    it('should write to both primary and LKG on writeAll', () => {
        const config = {
            test: { command: 'test', args: [], enabled: true }
        };
        store.writeAll(config);

        expect(fs.existsSync(configPath)).toBe(true);
        expect(fs.existsSync(lkgPath)).toBe(true);
        
        const primary = JSON.parse(fs.readFileSync(configPath, 'utf8'));
        const lkg = JSON.parse(fs.readFileSync(lkgPath, 'utf8'));
        
        expect(primary).toEqual(config);
        expect(lkg).toEqual(config);
    });

    it('should fallback to LKG if primary is missing', () => {
        const config = {
            test: { command: 'test', args: [], enabled: true }
        };
        // Setup LKG manually
        fs.writeFileSync(lkgPath, JSON.stringify(config));

        const result = store.readAll();
        expect(result).toEqual(config);
    });

    it('should fallback to LKG if primary is corrupted', () => {
        const config = {
            test: { command: 'test', args: [], enabled: true }
        };
        // Setup LKG
        store.writeAll(config);
        
        // Corrupt primary
        fs.writeFileSync(configPath, 'invalid json {');

        const result = store.readAll();
        expect(result).toEqual(config);
    });

    it('should update LKG when primary is successfully read', () => {
        const initialLkg = { old: { command: 'old', args: [], enabled: true } };
        const newPrimary = { new: { command: 'new', args: [], enabled: true } };
        
        fs.writeFileSync(lkgPath, JSON.stringify(initialLkg));
        fs.writeFileSync(configPath, JSON.stringify(newPrimary));

        const result = store.readAll();
        expect(result).toEqual(newPrimary);
        
        const lkgAfterRead = JSON.parse(fs.readFileSync(lkgPath, 'utf8'));
        expect(lkgAfterRead).toEqual(newPrimary);
    });
});
