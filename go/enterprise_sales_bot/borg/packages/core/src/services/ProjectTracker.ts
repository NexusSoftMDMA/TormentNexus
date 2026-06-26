/**
 * ProjectTracker
 *
 * Scans `task.md` and `ROADMAP.md` for the next actionable task and provides
 * progress metrics for the Director and dashboard.
 *
 * Architecture:
 * - Reads markdown task lists (`- [ ]`, `- [/]`, `- [x]`) to determine task state
 * - Supports nested sub-tasks (indentation-based hierarchy)
 * - Validates parsed task lines with Zod schemas
 * - Can mark tasks as started (`[/]`) or completed (`[x]`) by rewriting the source file
 *
 * State Persistence (Phase 62-8 Hardening):
 * - Maintains a JSON history file (`.tormentnexus/tracker_history.json`) for durable state
 * - Records completed tasks with timestamps for audit trail
 * - History survives markdown file rewrites and provides historical context
 *
 * Integration:
 * - Used by `Director` to auto-select idle tasks (Phase 59)
 * - Exposed via `systemProcedures.ts` → `getTaskStatus` tRPC route
 * - Dashboard shows real progress percentage from this service
 */

import fs from 'fs';
import path from 'path';
import { z } from 'zod';

/**
 * A single actionable task parsed from task.md or ROADMAP.md.
 */
export interface ProjectTask {
    /** Unique identifier (derived from filename + line number) */
    id: string;
    /** Human-readable task description (the text after the checkbox) */
    description: string;
    /** Current state: TODO ([ ]), IN_PROGRESS ([/]), DONE ([x]) */
    status: 'TODO' | 'IN_PROGRESS' | 'DONE';
    /** Absolute path to the source markdown file */
    sourceFile: string;
    /** Line number (1-indexed) in the source file */
    lineNumber: number;
}

/** Entry in the persistent task history JSON file */
interface TaskHistoryEntry {
    id: string;
    description: string;
    completedAt: string; // ISO timestamp
    sourceFile: string;
}

/** Zod schema for validating parsed checkbox states */
const TaskStateSchema = z.enum([' ', 'x', '/']);

/** Zod schema for validating a fully parsed task line */
const ParsedTaskLineSchema = z.object({
    indent: z.number().int().nonnegative(),
    state: TaskStateSchema,
    text: z.string().min(1),
    lineNumber: z.number().int().positive(),
});

type ParsedTaskLine = z.infer<typeof ParsedTaskLineSchema>;

export class ProjectTracker {
    /** Root directory of the project (used to locate task.md, ROADMAP.md) */
    private rootDir: string;

    /** Path to the JSON task history file for durable persistence */
    private historyPath: string;

    constructor(rootDir: string) {
        this.rootDir = rootDir;
        // Store history in .tormentnexus directory to keep project root clean
        this.historyPath = path.join(rootDir, '.tormentnexus', 'tracker_history.json');
    }

    /**
     * Scans task.md and ROADMAP.md for the next actionable task.
     * Priority:
     * 1. In Progress (but not done) in task.md
     * 2. Todo in task.md
     * 3. In Progress in ROADMAP.md
     * 4. Todo in ROADMAP.md
     */
    public getNextTask(): ProjectTask | null {
        // Default to local task.md in root
        const localTaskMd = path.join(this.rootDir, 'task.md');
        const docsTaskMd = path.join(this.rootDir, 'docs', 'task.md');

        let targetPath = fs.existsSync(localTaskMd) ? localTaskMd : (fs.existsSync(docsTaskMd) ? docsTaskMd : null);

        if (targetPath) {
            const task = this.findTaskInFile(targetPath);
            if (task) return task;
        }

        const roadmapPath = path.join(this.rootDir, 'ROADMAP.md');
        if (fs.existsSync(roadmapPath)) {
            const task = this.findTaskInFile(roadmapPath);
            if (task) return task;
        }

        return null;
    }

    /**
     * Returns overall project progress metrics.
     * Counts all checkbox items in task.md and calculates completion percentage.
     */
    public getStatus() {
        const taskMdPath = path.join(this.rootDir, 'task.md');
        let total = 0;
        let done = 0;

        if (fs.existsSync(taskMdPath)) {
            const content = fs.readFileSync(taskMdPath, 'utf-8');
            total += (content.match(/-\s*\[[ x/]\]/g) || []).length;
            done += (content.match(/-\s*\[x\]/g) || []).length;
        }

        const currentTask = this.getNextTask();

        return {
            taskId: currentTask ? currentTask.id : 'idle',
            status: currentTask ? 'busy' : 'idle',
            progress: total > 0 ? Math.round((done / total) * 100) : 0,
            currentTask: currentTask ? currentTask.description : 'Idle'
        };
    }

    /**
     * Returns the task completion history from the persistent JSON file.
     * Used for audit trail and dashboard display of completed work.
     */
    public getHistory(): TaskHistoryEntry[] {
        try {
            if (fs.existsSync(this.historyPath)) {
                const raw = fs.readFileSync(this.historyPath, 'utf-8');
                return JSON.parse(raw);
            }
        } catch (e) {
            console.warn('[ProjectTracker] Failed to read history:', e);
        }
        return [];
    }

    /**
     * Appends a completed task to the persistent history file.
     * Creates the .tormentnexus directory if it doesn't exist.
     */
    private appendToHistory(task: ProjectTask): void {
        try {
            const history = this.getHistory();
            history.push({
                id: task.id,
                description: task.description,
                completedAt: new Date().toISOString(),
                sourceFile: task.sourceFile,
            });

            // Ensure .tormentnexus directory exists
            const dir = path.dirname(this.historyPath);
            if (!fs.existsSync(dir)) {
                fs.mkdirSync(dir, { recursive: true });
            }

            fs.writeFileSync(this.historyPath, JSON.stringify(history, null, 2));
        } catch (e) {
            console.warn('[ProjectTracker] Failed to write history:', e);
        }
    }

    /**
     * Searches a markdown file for the first actionable task.
     * Priority: In-progress tasks first, then todo tasks.
     * For in-progress parents, drills down to find the first incomplete sub-task.
     */
    private findTaskInFile(filePath: string): ProjectTask | null {
        const content = fs.readFileSync(filePath, 'utf-8');
        const lines = content.split('\n');

        // 1. Check for In Progress task: - [/]
        for (let i = 0; i < lines.length; i++) {
            const parsed = this.parseTaskLine(lines[i], i + 1);
            if (parsed?.state === '/') {
                const subTask = this.findFirstSubTask(lines, i + 1, parsed.indent);
                if (subTask) {
                    return {
                        id: `task-${path.basename(filePath)}-${subTask.line}`,
                        description: subTask.text,
                        status: 'TODO',
                        sourceFile: filePath,
                        lineNumber: subTask.line
                    };
                }

                return {
                    id: `task-${path.basename(filePath)}-${i + 1}`,
                    description: parsed.text,
                    status: 'IN_PROGRESS',
                    sourceFile: filePath,
                    lineNumber: i + 1
                };
            }
        }

        // 2. Check for Todo task: - [ ]
        for (let i = 0; i < lines.length; i++) {
            const parsed = this.parseTaskLine(lines[i], i + 1);
            if (parsed?.state === ' ') {
                return {
                    id: `task-${path.basename(filePath)}-${i + 1}`,
                    description: parsed.text,
                    status: 'TODO',
                    sourceFile: filePath,
                    lineNumber: i + 1
                };
            }
        }

        return null;
    }

    /**
     * Parses a single line of markdown for a task checkbox pattern.
     * Regex matches: `- [ ] task text`, `- [x] task text`, `- [/] task text`
     * with optional leading whitespace for indentation.
     */
    private parseTaskLine(line: string, lineNumber: number): ParsedTaskLine | null {
        const match = line.match(/^(\s*)-\s*\[([ x/])\]\s*(.+?)\s*$/);
        if (!match) {
            return null;
        }

        const parsed = ParsedTaskLineSchema.safeParse({
            indent: match[1].length,
            state: match[2],
            text: match[3].trim(),
            lineNumber,
        });

        return parsed.success ? parsed.data : null;
    }

    /**
     * Starting from `startIndex`, finds the first incomplete sub-task
     * (deeper indentation than `parentIndent`).
     * Stops when encountering a sibling or parent-level item, or a heading.
     */
    private findFirstSubTask(lines: string[], startIndex: number, parentIndent: number): { text: string, line: number } | null {

        for (let i = startIndex; i < lines.length; i++) {
            const line = lines[i];
            if (line.trim() === '') continue;

            const parsed = this.parseTaskLine(line, i + 1);
            if (!parsed) {
                if (line.match(/^\s*#/)) break;
                continue;
            }

            if (parsed.indent <= parentIndent) {
                break;
            }

            if (parsed.state === ' ' || parsed.state === '/') {
                return { text: parsed.text, line: parsed.lineNumber };
            }
        }

        return null;
    }

    /**
     * Marks a task as completed ([x]) in its source markdown file.
     * Also records the completion in the persistent JSON history.
     */
    public completeTask(task: ProjectTask): void {
        const content = fs.readFileSync(task.sourceFile, 'utf-8');
        const lines = content.split('\n');

        if (lines[task.lineNumber - 1]) {
            lines[task.lineNumber - 1] = lines[task.lineNumber - 1].replace('[ ]', '[x]').replace('[/]', '[x]');
            fs.writeFileSync(task.sourceFile, lines.join('\n'));
            console.log(`[ProjectTracker] Marked task as done: ${task.description}`);

            // Persist to JSON history for audit trail
            this.appendToHistory(task);
        }
    }

    /**
     * Marks a task as in-progress ([/]) in its source markdown file.
     */
    public startTask(task: ProjectTask): void {
        const content = fs.readFileSync(task.sourceFile, 'utf-8');
        const lines = content.split('\n');

        if (lines[task.lineNumber - 1]) {
            lines[task.lineNumber - 1] = lines[task.lineNumber - 1].replace('[ ]', '[/]');
            fs.writeFileSync(task.sourceFile, lines.join('\n'));
            console.log(`[ProjectTracker] Marked task as in-progress: ${task.description}`);
        }
    }
}
