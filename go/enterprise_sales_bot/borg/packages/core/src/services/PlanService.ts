/**
 * Plan Service - Structured Development Workflow with Diff Sandbox
 * 
 * Implements Plan/Build modes for structured development workflow:
 * - PLAN mode: Exploration and planning, no file changes
 * - BUILD mode: Execution, apply changes via diff sandbox
 * 
 * Features:
 * - Cumulative diff review
 * - File-by-file approval
 * - Rollback support
 * - Branch experimentation
 * 
 * Inspired by Plandex's diff sandbox architecture.
 */

import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';

// Mode types
export type PlanMode = 'PLAN' | 'BUILD';

// Diff types
export interface FileDiff {
    id: string;
    filePath: string;
    originalContent: string | null;  // null if new file
    proposedContent: string;
    status: 'pending' | 'approved' | 'rejected' | 'applied';
    createdAt: Date;
    description?: string;
}

export interface DiffHunk {
    startLine: number;
    endLine: number;
    originalLines: string[];
    proposedLines: string[];
}

export interface DiffSandboxState {
    id: string;
    mode: PlanMode;
    diffs: Map<string, FileDiff>;
    checkpoints: Checkpoint[];
    createdAt: Date;
    updatedAt: Date;
}

export interface Checkpoint {
    id: string;
    name: string;
    timestamp: Date;
    diffIds: string[];
    description?: string;
}

export interface PlanServiceOptions {
    rootPath: string;
    sandboxDir?: string;
    autoCheckpoint?: boolean;
    maxCheckpoints?: number;
}

/**
 * Diff Sandbox - Manages cumulative diffs before applying
 */
export class DiffSandbox {
    private diffs = new Map<string, FileDiff>();
    private checkpoints: Checkpoint[] = [];

    constructor(private rootPath: string, private sandboxDir: string) {
        this.ensureSandboxDir();
    }

    private ensureSandboxDir(): void {
        if (!fs.existsSync(this.sandboxDir)) {
            fs.mkdirSync(this.sandboxDir, { recursive: true });
        }
    }

    /**
     * Generate a unique diff ID
     */
    private generateDiffId(): string {
        return crypto.randomBytes(8).toString('hex');
    }

    /**
     * Add a proposed file change to the sandbox
     */
    addDiff(filePath: string, proposedContent: string, description?: string): FileDiff {
        const absolutePath = path.resolve(this.rootPath, filePath);
        const originalContent = fs.existsSync(absolutePath)
            ? fs.readFileSync(absolutePath, 'utf-8')
            : null;

        const diff: FileDiff = {
            id: this.generateDiffId(),
            filePath,
            originalContent,
            proposedContent,
            status: 'pending',
            createdAt: new Date(),
            description,
        };

        this.diffs.set(diff.id, diff);

        // Save to sandbox for persistence
        this.saveDiffToSandbox(diff);

        return diff;
    }

    /**
     * Update an existing diff
     */
    updateDiff(diffId: string, proposedContent: string): FileDiff | null {
        const diff = this.diffs.get(diffId);
        if (!diff || diff.status !== 'pending') {
            return null;
        }

        diff.proposedContent = proposedContent;
        this.saveDiffToSandbox(diff);
        return diff;
    }

    /**
     * Save diff to sandbox directory
     */
    private saveDiffToSandbox(diff: FileDiff): void {
        const diffPath = path.join(this.sandboxDir, `${diff.id}.json`);
        fs.writeFileSync(diffPath, JSON.stringify(diff, null, 2));
    }

    /**
     * Get all pending diffs
     */
    getPendingDiffs(): FileDiff[] {
        return Array.from(this.diffs.values()).filter(d => d.status === 'pending');
    }

    /**
     * Get diff by ID
     */
    getDiff(diffId: string): FileDiff | null {
        return this.diffs.get(diffId) || null;
    }

    /**
     * Get diff for a specific file
     */
    getDiffForFile(filePath: string): FileDiff | null {
        for (const diff of this.diffs.values()) {
            if (diff.filePath === filePath && diff.status === 'pending') {
                return diff;
            }
        }
        return null;
    }

    /**
     * Approve a diff
     */
    approveDiff(diffId: string): boolean {
        const diff = this.diffs.get(diffId);
        if (!diff || diff.status !== 'pending') {
            return false;
        }
        diff.status = 'approved';
        this.saveDiffToSandbox(diff);
        return true;
    }

    /**
     * Reject a diff
     */
    rejectDiff(diffId: string): boolean {
        const diff = this.diffs.get(diffId);
        if (!diff || diff.status !== 'pending') {
            return false;
        }
        diff.status = 'rejected';
        this.saveDiffToSandbox(diff);
        return true;
    }

    /**
     * Apply an approved diff to the file system
     */
    applyDiff(diffId: string): boolean {
        const diff = this.diffs.get(diffId);
        if (!diff || diff.status !== 'approved') {
            return false;
        }

        const absolutePath = path.resolve(this.rootPath, diff.filePath);

        // Ensure directory exists
        const dir = path.dirname(absolutePath);
        if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true });
        }

        // Write the file
        fs.writeFileSync(absolutePath, diff.proposedContent, 'utf-8');
        diff.status = 'applied';
        this.saveDiffToSandbox(diff);
        return true;
    }

    /**
     * Apply all approved diffs
     */
    applyAllApproved(): { applied: string[]; failed: string[] } {
        const applied: string[] = [];
        const failed: string[] = [];

        for (const diff of this.diffs.values()) {
            if (diff.status === 'approved') {
                if (this.applyDiff(diff.id)) {
                    applied.push(diff.filePath);
                } else {
                    failed.push(diff.filePath);
                }
            }
        }

        return { applied, failed };
    }

    /**
     * Create a checkpoint of current diffs
     */
    createCheckpoint(name: string, description?: string): Checkpoint {
        const checkpoint: Checkpoint = {
            id: crypto.randomBytes(8).toString('hex'),
            name,
            timestamp: new Date(),
            diffIds: Array.from(this.diffs.keys()),
            description,
        };

        this.checkpoints.push(checkpoint);
        this.saveCheckpoint(checkpoint);
        return checkpoint;
    }

    /**
     * Save checkpoint to sandbox
     */
    private saveCheckpoint(checkpoint: Checkpoint): void {
        const checkpointsPath = path.join(this.sandboxDir, 'checkpoints.json');
        const checkpoints = fs.existsSync(checkpointsPath)
            ? JSON.parse(fs.readFileSync(checkpointsPath, 'utf-8'))
            : [];
        checkpoints.push(checkpoint);
        fs.writeFileSync(checkpointsPath, JSON.stringify(checkpoints, null, 2));
    }

    /**
     * Rollback to a checkpoint
     */
    rollbackToCheckpoint(checkpointId: string): boolean {
        const checkpointIndex = this.checkpoints.findIndex(c => c.id === checkpointId);
        if (checkpointIndex === -1) {
            return false;
        }

        const checkpoint = this.checkpoints[checkpointIndex];

        // Remove diffs created after this checkpoint
        const validDiffIds = new Set(checkpoint.diffIds);
        for (const [diffId, diff] of this.diffs) {
            if (!validDiffIds.has(diffId)) {
                // Revert applied changes
                if (diff.status === 'applied' && diff.originalContent !== null) {
                    const absolutePath = path.resolve(this.rootPath, diff.filePath);
                    fs.writeFileSync(absolutePath, diff.originalContent, 'utf-8');
                } else if (diff.status === 'applied' && diff.originalContent === null) {
                    // Delete new file
                    const absolutePath = path.resolve(this.rootPath, diff.filePath);
                    if (fs.existsSync(absolutePath)) {
                        fs.unlinkSync(absolutePath);
                    }
                }
                this.diffs.delete(diffId);
            }
        }

        // Remove later checkpoints
        this.checkpoints = this.checkpoints.slice(0, checkpointIndex + 1);
        return true;
    }

    /**
     * Get all checkpoints
     */
    getCheckpoints(): Checkpoint[] {
        return [...this.checkpoints];
    }

    /**
     * Generate unified diff format for a file diff
     */
    generateUnifiedDiff(diff: FileDiff): string {
        const originalLines = (diff.originalContent || '').split('\n');
        const proposedLines = diff.proposedContent.split('\n');

        const output: string[] = [
            `--- a/${diff.filePath}`,
            `+++ b/${diff.filePath}`,
        ];

        // Simple line-by-line diff
        const maxLines = Math.max(originalLines.length, proposedLines.length);
        let hunkStart = -1;
        let hunk: string[] = [];

        const flushHunk = () => {
            if (hunk.length > 0 && hunkStart >= 0) {
                output.push(`@@ -${hunkStart + 1},${originalLines.length} +${hunkStart + 1},${proposedLines.length} @@`);
                output.push(...hunk);
                hunk = [];
            }
        };

        for (let i = 0; i < maxLines; i++) {
            const orig = originalLines[i] ?? '';
            const prop = proposedLines[i] ?? '';

            if (orig !== prop) {
                if (hunkStart === -1) {
                    hunkStart = Math.max(0, i - 3);
                    // Add context lines
                    for (let j = hunkStart; j < i; j++) {
                        hunk.push(` ${originalLines[j] || ''}`);
                    }
                }
                if (originalLines[i] !== undefined) {
                    hunk.push(`-${orig}`);
                }
                if (proposedLines[i] !== undefined) {
                    hunk.push(`+${prop}`);
                }
            } else if (hunk.length > 0) {
                hunk.push(` ${orig}`);
                if (hunk.filter(l => !l.startsWith(' ')).length === 0) {
                    flushHunk();
                    hunkStart = -1;
                }
            }
        }

        flushHunk();
        return output.join('\n');
    }

    /**
     * Get summary of all diffs
     */
    getSummary(): string {
        const pending = this.getPendingDiffs();
        const approved = Array.from(this.diffs.values()).filter(d => d.status === 'approved');
        const applied = Array.from(this.diffs.values()).filter(d => d.status === 'applied');
        const rejected = Array.from(this.diffs.values()).filter(d => d.status === 'rejected');

        return [
            `Diff Sandbox Summary:`,
            `  Pending: ${pending.length}`,
            `  Approved: ${approved.length}`,
            `  Applied: ${applied.length}`,
            `  Rejected: ${rejected.length}`,
            `  Checkpoints: ${this.checkpoints.length}`,
        ].join('\n');
    }

    /**
     * Clear all diffs
     */
    clear(): void {
        this.diffs.clear();
        this.checkpoints = [];

        // Clear sandbox directory
        if (fs.existsSync(this.sandboxDir)) {
            const files = fs.readdirSync(this.sandboxDir);
            for (const file of files) {
                fs.unlinkSync(path.join(this.sandboxDir, file));
            }
        }
    }
}

/**
 * Plan Service - Manages Plan/Build mode workflow
 */
export class PlanService {
    private mode: PlanMode = 'PLAN';
    private sandbox: DiffSandbox;
    private options: Required<PlanServiceOptions>;

    constructor(options: PlanServiceOptions) {
        this.options = {
            rootPath: options.rootPath,
            sandboxDir: options.sandboxDir || path.join(options.rootPath, '.tormentnexus', 'sandbox'),
            autoCheckpoint: options.autoCheckpoint ?? true,
            maxCheckpoints: options.maxCheckpoints ?? 20,
        };

        this.sandbox = new DiffSandbox(this.options.rootPath, this.options.sandboxDir);
    }

    /**
     * Get current mode
     */
    getMode(): PlanMode {
        return this.mode;
    }

    /**
     * Switch to PLAN mode
     */
    enterPlanMode(): void {
        if (this.mode !== 'PLAN' && this.options.autoCheckpoint) {
            this.sandbox.createCheckpoint('mode-switch-to-plan', 'Auto-checkpoint before switching to PLAN mode');
        }
        this.mode = 'PLAN';
    }

    /**
     * Switch to BUILD mode
     */
    enterBuildMode(): void {
        if (this.mode !== 'BUILD' && this.options.autoCheckpoint) {
            this.sandbox.createCheckpoint('mode-switch-to-build', 'Auto-checkpoint before switching to BUILD mode');
        }
        this.mode = 'BUILD';
    }

    /**
     * Propose a file change (any mode, but only applied in BUILD mode)
     */
    proposeChange(filePath: string, content: string, description?: string): FileDiff {
        return this.sandbox.addDiff(filePath, content, description);
    }

    /**
     * Read a file (considers pending diffs in sandbox)
     */
    readFile(filePath: string): string | null {
        // Check if there's a pending diff for this file
        const diff = this.sandbox.getDiffForFile(filePath);
        if (diff) {
            return diff.proposedContent;
        }

        // Read from filesystem
        const absolutePath = path.resolve(this.options.rootPath, filePath);
        if (fs.existsSync(absolutePath)) {
            return fs.readFileSync(absolutePath, 'utf-8');
        }

        return null;
    }

    /**
     * Get pending diffs
     */
    getPendingChanges(): FileDiff[] {
        return this.sandbox.getPendingDiffs();
    }

    /**
     * Review a specific diff
     */
    reviewDiff(diffId: string): { diff: FileDiff; unifiedDiff: string } | null {
        const diff = this.sandbox.getDiff(diffId);
        if (!diff) return null;

        return {
            diff,
            unifiedDiff: this.sandbox.generateUnifiedDiff(diff),
        };
    }

    /**
     * Approve a diff
     */
    approveDiff(diffId: string): boolean {
        return this.sandbox.approveDiff(diffId);
    }

    /**
     * Reject a diff
     */
    rejectDiff(diffId: string): boolean {
        return this.sandbox.rejectDiff(diffId);
    }

    /**
     * Apply all approved diffs (only in BUILD mode)
     */
    applyApprovedChanges(): { applied: string[]; failed: string[] } | null {
        if (this.mode !== 'BUILD') {
            return null;  // Can only apply in BUILD mode
        }

        if (this.options.autoCheckpoint) {
            this.sandbox.createCheckpoint('pre-apply', 'Auto-checkpoint before applying changes');
        }

        return this.sandbox.applyAllApproved();
    }

    /**
     * Create a manual checkpoint
     */
    createCheckpoint(name: string, description?: string): Checkpoint {
        return this.sandbox.createCheckpoint(name, description);
    }

    /**
     * Rollback to a checkpoint
     */
    rollback(checkpointId: string): boolean {
        return this.sandbox.rollbackToCheckpoint(checkpointId);
    }

    /**
     * Get all checkpoints
     */
    getCheckpoints(): Checkpoint[] {
        return this.sandbox.getCheckpoints();
    }

    /**
     * Get status summary
     */
    getStatus(): string {
        return [
            `Mode: ${this.mode}`,
            this.sandbox.getSummary(),
        ].join('\n');
    }

    /**
     * Clear all pending changes
     */
    clearPendingChanges(): void {
        this.sandbox.clear();
    }

    /**
     * Get the diff sandbox (for advanced operations)
     */
    getSandbox(): DiffSandbox {
        return this.sandbox;
    }
}

export default PlanService;
