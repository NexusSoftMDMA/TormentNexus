/**
 * Example Workflows - Demonstrations of WorkflowEngine Capabilities
 * 
 * This file contains example workflow definitions showcasing:
 * - Multi-step processing pipelines
 * - Conditional routing
 * - Human-in-the-loop checkpoints
 * - State management patterns
 * 
 * Uses GraphBuilder for proper type-safe workflow construction.
 */

import { WorkflowEngine, WorkflowDefinition, WorkflowState } from '../orchestrator/WorkflowEngine.js';

interface CodeReviewInput {
    code?: string;
}

interface CodeAnalysis {
    lines?: number;
    complexity?: number;
}

interface ResearchInput {
    query?: string;
    depth?: string;
}

interface SearchSource {
    name: string;
    results: number;
    confidence: number;
}

interface SearchResultsShape {
    sources?: SearchSource[];
    totalResults?: number;
    avgConfidence?: number;
}

interface SynthesisShape {
    summary?: string;
    confidence?: number;
}

interface FileProcessingInput {
    path?: string;
}

/**
 * Reason: workflow state payload fields are intentionally dynamic and can be absent.
 * What: narrows unknown state slots into plain object maps for safe keyed access.
 * Why: keeps examples type-safe without changing their demonstration behavior.
 */
function asRecord(value: unknown): Record<string, unknown> {
    if (!value || typeof value !== 'object') {
        return {};
    }
    return value as Record<string, unknown>;
}

/**
 * Reason: examples repeatedly read `state.input` with loose shape assumptions.
 * What: returns a typed input payload when the state input is an object.
 * Why: removes broad casts while preserving concise example workflow logic.
 */
function getInput<T>(state: WorkflowState): T {
    return asRecord(state.input) as unknown as T;
}

// ============================================================================
// EXAMPLE 1: Code Review Pipeline
// ============================================================================

/**
 * Code Review Workflow
 * 
 * Stages:
 * 1. analyze_code - Analyze the code for patterns and issues
 * 2. security_check - Check for security vulnerabilities  
 * 3. style_check - Check code style and formatting
 * 4. generate_report - Combine all findings into a report
 * 5. complete - Finalize the review
 */
export function createCodeReviewWorkflow(engine: WorkflowEngine): WorkflowDefinition {
    const graph = WorkflowEngine.createGraph();

    graph.addNode('analyze_code', async (state: WorkflowState) => {
        const input = getInput<CodeReviewInput>(state);
        const code = input.code || '';
        const analysis = {
            lines: code.split('\n').length,
            hasAsync: /async\s+/.test(code),
            hasExports: /export\s+/.test(code),
            complexity: Math.min(10, Math.floor(code.length / 100)),
        };
        return { ...state, analysis };
    }, { name: 'Analyze Code' });

    graph.addNode('security_check', async (state: WorkflowState) => {
        const input = getInput<CodeReviewInput>(state);
        const code = input.code || '';
        const securityIssues: string[] = [];

        if (/eval\s*\(/.test(code)) securityIssues.push('Avoid using eval()');
        if (/innerHTML\s*=/.test(code)) securityIssues.push('Consider using textContent instead of innerHTML');
        if (/exec\s*\(/.test(code)) securityIssues.push('Be careful with exec() - potential injection');
        if (/password\s*[:=]\s*['"]/.test(code)) securityIssues.push('Hardcoded password detected');

        return { ...state, securityIssues };
    }, { name: 'Security Check' });

    graph.addNode('style_check', async (state: WorkflowState) => {
        const input = getInput<CodeReviewInput>(state);
        const code = input.code || '';
        const styleIssues: string[] = [];

        if (/\t/.test(code)) styleIssues.push('Use spaces instead of tabs');
        if (/console\.log/.test(code)) styleIssues.push('Remove console.log statements');
        if (/TODO|FIXME/i.test(code)) styleIssues.push('Unresolved TODO/FIXME comments');

        return { ...state, styleIssues };
    }, { name: 'Style Check' });

    graph.addNode('generate_report', async (state: WorkflowState) => {
        const analysis = asRecord(state.analysis) as unknown as CodeAnalysis;
        const securityIssues = (state.securityIssues as string[]) || [];
        const styleIssues = (state.styleIssues as string[]) || [];

        const report = {
            summary: {
                lines: analysis.lines ?? 0,
                complexity: analysis.complexity ?? 0,
                securityIssuesCount: securityIssues.length,
                styleIssuesCount: styleIssues.length,
            },
            securityIssues,
            styleIssues,
            needsHumanReview: securityIssues.length > 0 || (analysis.complexity ?? 0) > 7,
        };
        return { ...state, report };
    }, { name: 'Generate Report' });

    graph.addNode('complete', async (state: WorkflowState) => {
        const report = asRecord(state.report);
        const needsHumanReview = report.needsHumanReview === true;
        return {
            ...state,
            completed: true,
            completedAt: new Date().toISOString(),
            finalStatus: needsHumanReview ? 'reviewed' : 'auto-approved',
        };
    }, { name: 'Complete' });

    // Set entry point
    graph.setEntryPoint('analyze_code');

    // Add edges
    graph.addEdge('analyze_code', 'security_check');
    graph.addEdge('security_check', 'style_check');
    graph.addEdge('style_check', 'generate_report');
    graph.addEdge('generate_report', 'complete');

    const workflow = graph.build(
        'code-review-v1',
        'Code Review Pipeline',
        'Multi-stage code review with security and style checks'
    );

    engine.registerWorkflow(workflow);
    return workflow;
}

// ============================================================================
// EXAMPLE 2: Research Agent Pipeline
// ============================================================================

/**
 * Research Agent Workflow
 * 
 * Stages:
 * 1. parse_query - Parse and understand the research query
 * 2. search_sources - Search multiple sources for information
 * 3. synthesize - Combine findings into coherent summary
 * 4. finalize - Prepare final output
 */
export function createResearchWorkflow(engine: WorkflowEngine): WorkflowDefinition {
    const graph = WorkflowEngine.createGraph();

    graph.addNode('parse_query', async (state: WorkflowState) => {
        const input = getInput<ResearchInput>(state);
        const query = input.query || '';
        const parsed = {
            topic: query.split(' ').slice(0, 5).join(' '),
            keywords: query.toLowerCase().match(/\b\w{4,}\b/g) || [],
            depth: input.depth || 'standard',
        };
        return { ...state, parsedQuery: parsed };
    }, { name: 'Parse Query' });

    graph.addNode('search_sources', async (state: WorkflowState) => {
        // Simulate searching multiple sources
        const sources = [
            { name: 'web', results: 3, confidence: 0.8 },
            { name: 'docs', results: 2, confidence: 0.9 },
            { name: 'code', results: 5, confidence: 0.7 },
        ];
        const totalResults = sources.reduce((sum, s) => sum + s.results, 0);
        const avgConfidence = sources.reduce((sum, s) => sum + s.confidence, 0) / sources.length;

        return {
            ...state,
            searchResults: { sources, totalResults, avgConfidence },
        };
    }, { name: 'Search Sources' });

    graph.addNode('synthesize', async (state: WorkflowState) => {
        const searchResults = asRecord(state.searchResults) as unknown as SearchResultsShape;
        const synthesis = {
            summary: `Found ${searchResults.totalResults ?? 0} results across ${searchResults.sources?.length ?? 0} sources`,
            confidence: searchResults.avgConfidence ?? 0,
        };
        return { ...state, synthesis };
    }, { name: 'Synthesize' });

    graph.addNode('finalize', async (state: WorkflowState) => {
        const synthesis = asRecord(state.synthesis) as unknown as SynthesisShape;
        const searchResults = asRecord(state.searchResults) as unknown as SearchResultsShape;
        return {
            ...state,
            output: {
                summary: synthesis.summary,
                confidence: synthesis.confidence,
                sources: searchResults.sources?.map((s) => s.name) || [],
                completedAt: new Date().toISOString(),
            },
            completed: true,
        };
    }, { name: 'Finalize' });

    graph.setEntryPoint('parse_query');
    graph.addEdge('parse_query', 'search_sources');
    graph.addEdge('search_sources', 'synthesize');
    graph.addEdge('synthesize', 'finalize');

    const workflow = graph.build(
        'research-agent-v1',
        'Research Agent Pipeline',
        'Multi-source research with synthesis'
    );

    engine.registerWorkflow(workflow);
    return workflow;
}

// ============================================================================
// EXAMPLE 3: File Processing Pipeline with Conditional Routing
// ============================================================================

/**
 * File Processing Workflow
 * 
 * Demonstrates a workflow with LLM-like routing based on file type
 */
export function createFileProcessingWorkflow(engine: WorkflowEngine): WorkflowDefinition {
    const graph = WorkflowEngine.createGraph();

    graph.addNode('detect_type', async (state: WorkflowState) => {
        const input = getInput<FileProcessingInput>(state);
        const filePath = input.path || '';
        const ext = filePath.split('.').pop()?.toLowerCase() || '';
        const typeMap: Record<string, string> = {
            ts: 'code', tsx: 'code', js: 'code', jsx: 'code',
            py: 'code', rs: 'code', go: 'code',
            md: 'docs', txt: 'docs',
            json: 'data', yaml: 'data', yml: 'data',
        };
        return { ...state, fileType: typeMap[ext] || 'unknown' };
    }, { name: 'Detect File Type' });

    graph.addNode('process_file', async (state: WorkflowState) => {
        const fileType = state.fileType as string;

        const handlerMap: Record<string, { handler: string; actions: string[] }> = {
            code: { handler: 'code', actions: ['lint', 'format', 'analyze'] },
            docs: { handler: 'docs', actions: ['spell-check', 'link-check'] },
            data: { handler: 'data', actions: ['validate-schema', 'format'] },
            unknown: { handler: 'unknown', actions: [] },
        };

        const { handler, actions } = handlerMap[fileType] || handlerMap.unknown;

        return {
            ...state,
            processed: fileType !== 'unknown',
            handler,
            actions,
        };
    }, { name: 'Process File' });

    graph.addNode('output', async (state: WorkflowState) => {
        const input = getInput<FileProcessingInput>(state);
        return {
            ...state,
            completed: true,
            output: {
                path: input.path,
                type: state.fileType,
                handler: state.handler,
                actions: state.actions || [],
                success: state.processed === true,
            },
        };
    }, { name: 'Output Results' });

    graph.setEntryPoint('detect_type');
    graph.addEdge('detect_type', 'process_file');
    graph.addEdge('process_file', 'output');

    const workflow = graph.build(
        'file-processing-v1',
        'File Processing Pipeline',
        'Process files based on type with specialized handlers'
    );

    engine.registerWorkflow(workflow);
    return workflow;
}

// ============================================================================
// Usage Example
// ============================================================================

/**
 * Example: Running a workflow
 * 
 * ```typescript
 * import { WorkflowEngine } from './orchestrator/WorkflowEngine';
 * import { createCodeReviewWorkflow } from './examples/ExampleWorkflows';
 * 
 * const engine = new WorkflowEngine({ persistDir: '.tormentnexus/workflows' });
 * const workflow = createCodeReviewWorkflow(engine);
 * 
 * const result = await engine.start(workflow.id, {
 *   input: {
 *     code: `
 *       async function processData(input) {
 *         const result = await fetch(input.url);
 *         console.log(result);
 *         return result.json();
 *       }
 *     `
 *   }
 * });
 * 
 * console.log(result.state.report);
 * ```
 */

export default {
    createCodeReviewWorkflow,
    createResearchWorkflow,
    createFileProcessingWorkflow,
};
