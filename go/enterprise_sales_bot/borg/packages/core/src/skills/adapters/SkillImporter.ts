
import fs from 'fs/promises';
import path from 'path';

interface ExternalSkill {
    name: string;
    description: string;
    commands?: string[];
    // ... typical structure varies
}

export class SkillImporter {
    private importDir: string;

    constructor() {
        this.importDir = path.join(process.cwd(), 'packages/core/data/raw_skills');
    }

    async scanAndImport() {
        try {
            await fs.mkdir(this.importDir, { recursive: true });
            const files = await fs.readdir(this.importDir);

            const skills = [];

            for (const file of files) {
                if (file.endsWith('.json')) {
                    const content = await fs.readFile(path.join(this.importDir, file), 'utf-8');
                    const raw = JSON.parse(content);

                    // Simple adapter: Assume raw has name/desc
                    skills.push({
                        name: raw.name || file.replace('.json', ''),
                        description: raw.description || "Imported skill",
                        source: 'local-import',
                        path: path.join(this.importDir, file)
                    });
                }
            }

            return skills;
        } catch (e) {
            console.error("Skill import failed", e);
            return [];
        }
    }
}
