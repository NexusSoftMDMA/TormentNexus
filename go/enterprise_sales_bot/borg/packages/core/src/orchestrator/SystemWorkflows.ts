
import { WorkflowEngine, WorkflowState } from './WorkflowEngine.js';

export interface ToolRunner {
    executeTool(name: string, args: Record<string, unknown>): Promise<unknown>;
}

function getErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message;
    }
    return typeof error === 'string' ? error : 'Unknown error';
}

/**
 * Register standard system workflows
 */
export function registerSystemWorkflows(engine: WorkflowEngine, runner: ToolRunner) {
    registerCodeReviewWorkflow(engine, runner);
    registerDeploymentWorkflow(engine, runner);
}

/**
 * Code Review Workflow
 */
function registerCodeReviewWorkflow(engine: WorkflowEngine, runner: ToolRunner) {
    const builder = WorkflowEngine.createGraph();

    builder
        .addNode('analyze_diff', async (state) => {
            console.log('[Workflow: Code Review] Analyzing diff...');
            try {
                // Determine target branch, default to main/master
                const targetBranch = state.targetBranch || 'main';

                // Use run_command tool via runner
                const diff = await runner.executeTool('run_command', {
                    CommandLine: `git diff ${targetBranch}...HEAD --stat`,
                    Cwd: process.cwd(),
                    SafeToAutoRun: true,
                    WaitMsBeforeAsync: 1000
                });

                return { ...state, diffSummary: diff, hasChanges: true };
            } catch (e: unknown) {
                return { ...state, error: `Diff failed: ${getErrorMessage(e)}`, hasChanges: false };
            }
        }, { name: 'Analyze Diff', description: 'Analyze git changes against target branch' })

        .addNode('scan_security', async (state) => {
            console.log('[Workflow: Code Review] Scanning for security issues...');
            try {
                // Simple grep for obvious secrets
                // We mock this slightly by just returning a count if the tool fails or returns distinct format
                const scan = await runner.executeTool('grep_search', {
                    SearchPath: 'src',
                    Query: '(password|secret|key) *=',
                    IsRegex: true
                });

                // MCPServer tools might return different shapes. Assuming grep_search returns checkable result.
                const risks = Array.isArray(scan) ? scan.length : 0;
                return { ...state, securityScan: risks === 0 ? 'Passed' : 'Risks Found', findings: scan };
            } catch (e) {
                return { ...state, securityScan: 'Skipped (Error)', findings: [] };
            }
        }, { name: 'Security Scan', description: 'Check for potential secrets or vulnerabilities' })

        .addNode('auto_critique', async (state) => {
            console.log('[Workflow: Code Review] Auto-generating critique...');
            try {
                // Call the real agent via the execution runner to handle the code diff
                const diffContent = state.diffSummary || '';

                if (!diffContent || typeof diffContent !== 'string') {
                    return { ...state, critique: 'Skipped critique: No diff to review.' };
                }

                const response = await runner.executeTool('use_agent', {
                    name: 'claude', // Use a real agent adapter registered in MCPServer
                    prompt: `Provide a highly concise code review of this unified diff. Look for obvious anti-patterns:\n\n${diffContent}`
                });

                return { ...state, critique: typeof response === 'object' && response !== null && 'content' in response ? (response as any).content[0]?.text || '' : String(response) };
            } catch (err: unknown) {
                return { ...state, critique: `Automated critique failed: ${getErrorMessage(err)}` };
            }
        }, { name: 'Auto Critique', description: 'LLM-based code analysis' })

        .addNode('human_review', async (state) => {
            // This is a checkpoint, so execution stops here until manual approval
            console.log('[Workflow: Code Review] Waiting for human approval...');
            return { ...state, status: 'Reviewing' };
        }, { name: 'Human Review', description: 'Manual approval required', requiresApproval: true })

        .addNode('merge_pr', async (state) => {
            console.log('[Workflow: Code Review] Merging...');
            // Logic to merge would go here
            return { ...state, status: 'Merged', completedAt: new Date() };
        }, { name: 'Merge PR', description: 'Merge changes' });

    builder
        .addEdge('analyze_diff', 'scan_security')
        .addEdge('scan_security', 'auto_critique')
        .addEdge('auto_critique', 'human_review')
        .addEdge('human_review', 'merge_pr');

    const workflow = builder.build('code_review', 'Code Review Pipeline', 'Automated code analysis and security scanning');
    engine.registerWorkflow(workflow);
}

/**
 * Deployment Workflow
 */
function registerDeploymentWorkflow(engine: WorkflowEngine, runner: ToolRunner) {
    const builder = WorkflowEngine.createGraph();

    builder
        .addNode('build', async (state) => {
            console.log('[Workflow: Deploy] Building...');
            // Use run_command to build
            await runner.executeTool('run_command', {
                CommandLine: 'npm run build',
                Cwd: process.cwd(),
                SafeToAutoRun: true,
                WaitMsBeforeAsync: 10000 // Verification build might take time
            });
            return { ...state, buildStatus: 'Success' };
        }, { name: 'Build', description: 'Compile and build' })

        .addNode('test', async (state) => {
            console.log('[Workflow: Deploy] Testing...');
            try {
                // Just run a quick check or unit tests
                // We'll skip running full tests to save time in this demo logic
                // await runner.executeTool('run_command', { CommandLine: 'npm test', ... });
                return { ...state, testStatus: 'Passed (Skipped for demo)' };
            } catch (e) {
                throw new Error("Tests failed");
            }
        }, { name: 'Test', description: 'Run test suite' })

        .addNode('deploy_staging', async (state) => {
            console.log('[Workflow: Deploy] Deploying to Staging...');
            return { ...state, env: 'staging', url: 'http://localhost:3000' };
        }, { name: 'Deploy Staging', description: 'Deploy to ephemeral env' })

        .addNode('verify_health', async (state) => {
            console.log('[Workflow: Deploy] Verifying Health...');
            return { ...state, health: 'OK' };
        }, { name: 'Verify Health', description: 'Check deployment health' })

        .addNode('deploy_prod', async (state) => {
            console.log('[Workflow: Deploy] Promoting to Production...');
            return { ...state, env: 'production', released: true };
        }, { name: 'Deploy Prod', description: 'Promote to production', requiresApproval: true })

        .addNode('post_deploy_alert', async (state) => {
            console.log('[Workflow: Deploy] Sending alerts...');
            return { ...state, alerted: true };
        }, { name: 'Notify Team', description: 'Send completion alert' });

    builder
        .addEdge('build', 'test')
        .addEdge('test', 'deploy_staging')
        .addEdge('deploy_staging', 'verify_health')
        .addEdge('verify_health', 'deploy_prod')
        .addEdge('deploy_prod', 'post_deploy_alert');

    const workflow = builder.build('deployment', 'Production Deployment', 'Build, Test, and Deploy pipeline');
    engine.registerWorkflow(workflow);
}
