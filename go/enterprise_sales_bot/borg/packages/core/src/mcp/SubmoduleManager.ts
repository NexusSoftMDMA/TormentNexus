import { spawn } from "child_process";
import path from "path";
import fs from "fs";

export class SubmoduleManager {
    private submodulesDir: string;

    constructor(rootDir: string) {
        this.submodulesDir = path.join(rootDir, 'submodules');
        if (!fs.existsSync(this.submodulesDir)) {
            fs.mkdirSync(this.submodulesDir, { recursive: true });
        }
    }

    public async addSubmodule(name: string, repoUrl: string): Promise<string> {
        const targetPath = path.join(this.submodulesDir, name);

        if (fs.existsSync(targetPath)) {
            return `Submodule ${name} already exists at ${targetPath}`;
        }

        console.log(`[SubmoduleManager] Cloning ${repoUrl} to ${targetPath}...`);

        return new Promise((resolve, reject) => {
            const git = spawn("git", ["clone", repoUrl, name], {
                cwd: this.submodulesDir,
                stdio: 'inherit'
            });

            git.on('close', (code) => {
                if (code === 0) resolve(`Successfully cloned ${name}`);
                else reject(new Error(`Git clone failed with code ${code}`));
            });
        });
    }

    public async updateSubmodule(name: string): Promise<string> {
        const targetPath = path.join(this.submodulesDir, name);
        if (!fs.existsSync(targetPath)) {
            throw new Error(`Submodule ${name} not found.`);
        }

        return new Promise((resolve, reject) => {
            const git = spawn("git", ["pull"], {
                cwd: targetPath,
                stdio: 'inherit'
            });

            git.on('close', (code) => {
                if (code === 0) resolve(`Successfully updated ${name}`);
                else reject(new Error(`Git pull failed with code ${code}`));
            });
        });
    }

    public listSubmodules(): string[] {
        if (!fs.existsSync(this.submodulesDir)) return [];
        return fs.readdirSync(this.submodulesDir).filter(f => {
            return fs.statSync(path.join(this.submodulesDir, f)).isDirectory();
        });
    }
}
