
import glob from 'fast-glob';
import fs from 'fs/promises';
import path from 'path';
import matter from 'gray-matter';

/**
 * Skill — A portable, self-contained instruction set (runbook).
 *
 * Skills are markdown files with YAML frontmatter (name, description) stored
 * as `SKILL.md` in directories under the configured search paths.
 *
 * Example: `.tormentnexus/skills/deploy-to-prod/SKILL.md`
 */
export interface Skill {
    /** Unique identifier — derived from frontmatter `name` or parent directory name */
    id: string;
    /** Human-readable skill name */
    name: string;
    /** Brief description of what this skill does */
    description: string;
    /** The full markdown body (instructions) of the skill */
    content: string;
    /** Absolute path to the SKILL.md file on disk */
    path: string;
}

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

/**
 * SkillRegistry
 *
 * Discovers, loads, and manages TormentNexus's skill library.
 * Skills are portable runbooks (SKILL.md files with YAML frontmatter) that
 * agents can read and follow to perform specialized tasks.
 *
 * Discovery:
 * - Scans configured `searchPaths` directories for `SKILL.md` files
 * - Uses `fast-glob` with max depth 3 to find skill files
 * - Parses YAML frontmatter via `gray-matter` for metadata
 *
 * MCP Tool Exposure:
 * - `getSkillTools()` returns MCP-compatible tool definitions (list, read, create, update)
 * - `skillsRouter.ts` exposes skills via tRPC for the dashboard frontend
 *
 * Master Index:
 * - Optional `masterIndexPath` points to a JSONC file cataloging all available
 *   MCP servers, harnesses, and skills for the library UI.
 */
export class SkillRegistry {
    /** In-memory cache of loaded skills, keyed by skill ID */
    private skills: Map<string, Skill> = new Map();
    /** Directories to scan for SKILL.md files */
    private searchPaths: string[];
    /** Optional path to master library index (JSONC) */
    private masterIndexPath?: string;

    constructor(searchPaths: string[]) {
        this.searchPaths = searchPaths;
    }


    setMasterIndexPath(indexPath: string) {
        this.masterIndexPath = indexPath;
    }

    async getLibraryIndex() {
        if (!this.masterIndexPath) {
            return { categories: { mcp_servers: [], universal_harness: [], skills: [] } };
        }

        try {
            const content = await fs.readFile(this.masterIndexPath, 'utf-8');
            // Remove comments (JSONC)
            const cleanJSON = content.replace(/\\"|"(?:\\"|[^"])*"|(\/\/.*|\/\*[\s\S]*?\*\/)/g, (m, g) => g ? "" : m);
            return JSON.parse(cleanJSON);
        } catch (e) {
            console.error("Error reading master index:", e);
            return { categories: { mcp_servers: [], universal_harness: [], skills: [] } };
        }
    }

    hasSkill(id: string): boolean {
        // Skill IDs in the map are usually the folders/names
        return this.skills.has(id);
    }

    getSkills(): Skill[] {
        return Array.from(this.skills.values());
    }

    async loadSkills() {
        this.skills.clear();

        for (const loc of this.searchPaths) {
            try {
                // Find all SKILL.md files (case-insensitive for Windows, but explicit for Linux)
                const entries = await glob(['**/SKILL.md', '**/skill.md'], {
                    cwd: loc,
                    absolute: true,
                    deep: 3 // Go deeper just in case
                });

                for (const file of entries) {
                    await this.parseSkill(file);
                }
            } catch (e) {
                // Ignore missing directories
            }
        }
        console.log(`TormentNexus Core: Loaded ${this.skills.size} skills.`);
    }

    private async parseSkill(filePath: string) {
        try {
            const raw = await fs.readFile(filePath, 'utf-8');
            const { data, content } = matter(raw);

            // Use 'name' from frontmatter or folder name
            const id = data.name || path.basename(path.dirname(filePath));

            const skill: Skill = {
                id,
                name: data.name || id,
                description: data.description || "No description provided",
                content,
                path: filePath
            };

            this.skills.set(id, skill);
        } catch (e) {
            console.error(`Failed to parse skill at ${filePath}`, e);
        }
    }

    getSkillTools() {
        return [
            {
                name: "list_skills",
                description: "List all available skills (runbooks)",
                inputSchema: { type: "object", properties: {} }
            },
            {
                name: "read_skill",
                description: "Read the content/instructions of a specific skill",
                inputSchema: {
                    type: "object",
                    properties: {
                        skillName: { type: "string" }
                    },
                    required: ["skillName"]
                }
            },
            {
                name: "create_skill",
                description: "Create a new skill (runbook)",
                inputSchema: {
                    type: "object",
                    properties: {
                        id: { type: "string", description: "Unique ID (folder name)" },
                        name: { type: "string" },
                        description: { type: "string" }
                    },
                    required: ["id", "name", "description"]
                }
            },
            {
                name: "search_skills",
                description: "Search for skills (runbooks) by name or description",
                inputSchema: {
                    type: "object",
                    properties: {
                        query: { type: "string" }
                    },
                    required: ["query"]
                }
            },
            {
                name: "update_skill",
                description: "Update content of an existing skill",
                inputSchema: {
                    type: "object",
                    properties: {
                        id: { type: "string" },
                        content: { type: "string" }
                    },
                    required: ["id", "content"]
                }
            }
        ];
    }

    async searchSkills(query: string) {
        const queryLower = query.toLowerCase();
        const matches = Array.from(this.skills.values())
            .filter(s =>
                s.id.toLowerCase().includes(queryLower) ||
                s.name.toLowerCase().includes(queryLower) ||
                s.description.toLowerCase().includes(queryLower)
            )
            .map(s => ({
                id: s.id,
                name: s.name,
                description: s.description
            }));

        return {
            content: [{
                type: "text",
                text: JSON.stringify({ matches }, null, 2)
            }]
        };
    }

    async listSkills() {
        // Progressive disclosure: only return the skill names/ids without descriptions
        const skillList = Array.from(this.skills.values()).map(s => s.id);

        return {
            content: [{
                type: "text",
                text: JSON.stringify({ skills: skillList }, null, 2)
            }]
        };
    }

    async readSkill(skillName: string) {
        const skill = this.skills.get(skillName);
        if (!skill) {
            return {
                content: [{ type: "text", text: `Skill '${skillName}' not found.` }]
            };
        }

        return {
            content: [{
                type: "text",
                text: skill.content
            }]
        };
    }

    async createSkill(id: string, name: string, description: string) {
        // Default to the first search path (usually .tormentnexus/skills in cwd)
        const targetDir = this.searchPaths[0];
        const skillDir = path.join(targetDir, id);
        const skillFile = path.join(skillDir, 'SKILL.md');

        try {
            await fs.mkdir(skillDir, { recursive: true });

            const content = `---
name: ${name}
description: ${description}
---

# ${name}

${description}

## Instructions
1. ...
`;
            await fs.writeFile(skillFile, content, 'utf-8');

            // Reload to pick up new skill
            await this.parseSkill(skillFile);

            return {
                content: [{ type: "text", text: `Created skill '${name}' at ${skillFile}` }]
            };
        } catch (e: unknown) {
            return { content: [{ type: "text", text: `Error creating skill: ${getErrorMessage(e)}` }] };
        }
    }

    async saveSkill(id: string, content: string) {
        const skill = this.skills.get(id);
        if (!skill) {
            return { content: [{ type: "text", text: `Skill '${id}' not found.` }] };
        }

        try {
            await fs.writeFile(skill.path, content, 'utf-8');

            // Update in-memory
            skill.content = content;
            // Re-parse to update frontmatter if changed
            await this.parseSkill(skill.path);

            return { content: [{ type: "text", text: `Saved skill '${id}'.` }] };
        } catch (e: unknown) {
            return { content: [{ type: "text", text: `Error saving skill: ${getErrorMessage(e)}` }] };
        }
    }
}
