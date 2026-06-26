
import fs from 'fs';
import path from 'path';
import { spawn } from 'child_process';
import { RepoGraphService } from './RepoGraphService.js';

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

export class AutoTestService {
    private watcher: fs.FSWatcher | null = null;
    private rootDir: string;
    private processingParams: Set<string> = new Set();
    public repoGraph: RepoGraphService;
    public testResults: Map<string, { status: 'pass' | 'fail' | 'running', timestamp: number, output?: string }> = new Map();
    public isRunning: boolean = false;

    constructor(rootDir: string) {
        this.rootDir = rootDir;
        this.repoGraph = new RepoGraphService(rootDir);
    }

    async start() {
        if (this.watcher) return;

        console.log(`[AutoTest] Building Dependency Graph...`);
        await this.repoGraph.buildGraph();

        console.log(`[AutoTest] Watching ${this.rootDir} for changes...`);
        try {
            this.watcher = fs.watch(this.rootDir, { recursive: true }, (eventType, filename) => {
                if (filename && !this.isIgnored(filename)) {
                    this.handleFileChange(filename);
                }
            });
            this.isRunning = true;
        } catch (e: unknown) {
            console.error(`[AutoTest] Watch failed: ${getErrorMessage(e)}`);
        }
    }

    stop() {
        if (this.watcher) {
            this.watcher.close();
            this.watcher = null;
        }
        this.isRunning = false;
    }

    private isIgnored(filename: string): boolean {
        if (filename.includes('node_modules')) return true;
        if (filename.includes('.git')) return true;
        if (filename.includes('dist')) return true;
        if (filename.includes('.turbo')) return true;
        if (filename.endsWith('.log')) return true;
        return false;
    }

    private handleFileChange(filename: string) {
        if (this.processingParams.has(filename)) return;

        this.processingParams.add(filename);
        setTimeout(() => this.processingParams.delete(filename), 2000);

        console.log(`[AutoTest] File changed: ${filename}`);

        const testFile = this.findTestFile(filename);
        if (testFile) {
            console.log(`[AutoTest] Found test file: ${testFile}. Running...`);
            this.runTest(testFile);
        }

        // Check consumers
        const consumers = this.repoGraph.getConsumers(filename);
        if (consumers.length > 0) {
            console.log(`[AutoTest] Forward Dependencies: ${consumers.join(', ')}`);
            for (const consumer of consumers) {
                const consumerTest = this.findTestFile(consumer);
                if (consumerTest && consumerTest !== testFile) {
                    console.log(`[AutoTest] Triggering dependent test: ${consumerTest}`);
                    this.runTest(consumerTest);
                }
            }
        }
    }

    private findTestFile(sourceFile: string): string | null {
        // Normalize slashes
        const normalized = sourceFile.split(path.sep).join('/');

        // 1. Is it a test file?
        if (normalized.endsWith('.test.ts') || normalized.endsWith('.spec.ts')) {
            return path.join(this.rootDir, sourceFile);
        }

        // 2. Sibling test: foo.ts -> foo.test.ts
        const siblingTest = sourceFile.replace(/\.ts$/, '.test.ts');
        if (fs.existsSync(path.join(this.rootDir, siblingTest))) {
            return path.join(this.rootDir, siblingTest);
        }

        // 3. Central test dir heuristic
        // packages/pkg/src/foo.ts -> packages/pkg/test/foo.test.ts
        const parts = sourceFile.split(path.sep);
        // Find 'src' and backtrack to package root
        const srcIndex = parts.lastIndexOf('src');
        if (srcIndex > 0) { // e.g. packages/core/src/...
            const pkgRoot = parts.slice(0, srcIndex).join(path.sep); // packages/core
            const fileName = path.basename(sourceFile, '.ts');
            const testPath = path.join(this.rootDir, pkgRoot, 'test', `${fileName}.test.ts`);
            if (fs.existsSync(testPath)) {
                return testPath;
            }
        }

        return null;
    }

    private runTest(testFilePath: string) {
        console.log(`[AutoTest] Spawning vitest for ${testFilePath}`);
        this.testResults.set(testFilePath, { status: 'running', timestamp: Date.now() });

        const p = spawn('npx', ['vitest', 'run', testFilePath], {
            cwd: this.rootDir,
            shell: true,
            stdio: 'pipe' // Pipe to capture output
        });

        let output = '';
        p.stdout?.on('data', (d) => { output += d.toString(); process.stderr.write(d); });
        p.stderr?.on('data', (d) => { output += d.toString(); process.stderr.write(d); });

        p.on('close', (code) => {
            if (code === 0) {
                console.log(`[AutoTest] ✅ Passed: ${testFilePath}`);
                this.testResults.set(testFilePath, { status: 'pass', timestamp: Date.now(), output });
            } else {
                console.error(`[AutoTest] ❌ Failed: ${testFilePath}`);
                this.testResults.set(testFilePath, { status: 'fail', timestamp: Date.now(), output });
            }
        });
    }
}
