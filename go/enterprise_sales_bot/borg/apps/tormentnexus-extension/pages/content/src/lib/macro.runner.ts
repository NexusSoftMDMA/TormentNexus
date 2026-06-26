import { type Macro } from '../stores';
import type { Edge, Node } from '@xyflow/react';

export class MacroRunner {
  private sendMessage: (toolName: string, args: any) => Promise<any>;
  private onLog: (message: string, type: 'info' | 'error' | 'success') => void;

  constructor(
    sendMessage: (toolName: string, args: any) => Promise<any>,
    onLog: (message: string, type: 'info' | 'error' | 'success') => void
  ) {
    this.sendMessage = sendMessage;
    this.onLog = onLog;
  }

  async run(macro: Macro) {
    this.onLog(`Starting macro: ${macro.name}`, 'info');

    // Build adjacency list for edges
    const incomingEdges = new Map<string, Edge[]>();
    const outgoingEdges = new Map<string, Edge[]>();

    for (const node of macro.nodes) {
      incomingEdges.set(node.id, []);
      outgoingEdges.set(node.id, []);
    }

    for (const edge of macro.edges) {
      if (!incomingEdges.has(edge.target)) incomingEdges.set(edge.target, []);
      if (!outgoingEdges.has(edge.source)) outgoingEdges.set(edge.source, []);
      incomingEdges.get(edge.target)!.push(edge);
      outgoingEdges.get(edge.source)!.push(edge);
    }

    // Identify start nodes (nodes with 0 incoming edges)
    // If a workflow has cycles, indegree 0 might not exist or might miss nodes.
    // For simplicity, we start with indegree 0.
    const startNodes = macro.nodes.filter(n => incomingEdges.get(n.id)!.length === 0);

    // Fallback: if no empty incoming edges, randomly pick the first node.
    const queue: Node[] = startNodes.length > 0 ? [...startNodes] : (macro.nodes.length > 0 ? [macro.nodes[0]] : []);

    const results: any[] = [];
    let lastResult: any = null;
    const env: Record<string, any> = {};
    const visited = new Set<string>();

    const maxSteps = 1000; // Safety limit
    let stepCount = 0;

    while (queue.length > 0) {
      if (stepCount++ > maxSteps) {
        throw new Error('Macro execution exceeded maximum step limit (potential infinite loop)');
      }

      const node = queue.shift()!;
      if (visited.has(node.id)) continue; // Keep it simple: no cycles, run once per node
      visited.add(node.id);

      this.onLog(`Executing node ${node.id} (${node.type || 'default'})`, 'info');

      try {
        const data = (node.data || {}) as Record<string, any>;
        let nextHandle: string | null = null; // Used by condition nodes to route next target

        if (node.type === 'tool') {
          if (!data.toolName) throw new Error('Tool name missing in tool block');

          this.onLog(`Running tool: ${data.toolName}`, 'info');
          const processedArgs = this.processArgs(data.args, lastResult, results, env);

          lastResult = await this.sendMessage(data.toolName, processedArgs);
          results.push(lastResult);

          this.onLog(`Tool finished: ${data.toolName}`, 'success');

        } else if (node.type === 'delay') {
          const delay = data.delayMs || 1000;
          this.onLog(`Waiting ${delay}ms...`, 'info');
          await new Promise(r => setTimeout(r, delay));

        } else if (node.type === 'set_variable') {
          const name = data.variableName;
          const valueExpr = data.variableValue || '';

          if (name) {
            let value: any = valueExpr;

            if (valueExpr === 'lastResult') {
              value = lastResult;
            } else if (typeof valueExpr === 'string' && valueExpr.startsWith('lastResult.')) {
              const path = valueExpr.replace('lastResult.', '');
              value = path.split('.').reduce((o: any, k: string) => (o || {})[k], lastResult);
            } else if (typeof valueExpr === 'string' && valueExpr.startsWith('env.')) {
              const path = valueExpr.replace('env.', '');
              value = path.split('.').reduce((o: any, k: string) => (o || {})[k], env);
            } else if (valueExpr === 'allResults') {
              value = results;
            } else if (typeof valueExpr === 'string' && !isNaN(Number(valueExpr)) && valueExpr.trim() !== '') {
              value = Number(valueExpr);
            }

            env[name] = value;
            this.onLog(`Variable set: ${name} = ${JSON.stringify(value)}`, 'info');
          }

        } else if (node.type === 'condition') {
          const condition = data.expression || 'false';
          let conditionResult = false;

          try {
            conditionResult = this.evaluateCondition(condition, { lastResult, allResults: results, env });
            this.onLog(`Condition evaluated to: ${conditionResult}`, 'info');
          } catch (e) {
            this.onLog(`Condition evaluation error: ${e}`, 'error');
            conditionResult = false;
          }

          // Condition nodes map true/false to specific target handles on their output edges.
          // By convention, condition node has sourceHandle='true' and sourceHandle='false'.
          nextHandle = conditionResult ? 'true' : 'false';
        }

        // Push subsequent nodes into the execution queue
        const outEdges = outgoingEdges.get(node.id) || [];
        for (const edge of outEdges) {
          // If we evaluated a Condition, only follow the edge matching the result handle
          if (nextHandle && edge.sourceHandle && edge.sourceHandle !== nextHandle) {
            continue;
          }

          const targetNode = macro.nodes.find(n => n.id === edge.target);
          if (targetNode && !visited.has(targetNode.id)) {
            queue.push(targetNode);
          }
        }

      } catch (error) {
        this.onLog(`Node execution failed: ${error}`, 'error');
        throw error;
      }
    }

    this.onLog(`Macro ${macro.name} completed`, 'success');
    return lastResult;
  }

  private processArgs(args: any, lastResult: any, allResults: any[], env: any): any {
    if (!args) return {};
    let processed = JSON.parse(JSON.stringify(args));

    const walk = (obj: any) => {
      for (const key in obj) {
        if (typeof obj[key] === 'string') {
          const val = obj[key];
          // Check for variable substitution: {{variableName}}
          if (val.match(/^\{\{[\w\d\.\[\]]+\}\}$/)) {
            // Exact match substitution (preserves type)
            const varName = val.slice(2, -2);
            if (varName === 'lastResult') obj[key] = lastResult;
            else if (env[varName] !== undefined) obj[key] = env[varName];
            else if (varName.startsWith('lastResult.')) {
              const path = varName.replace('lastResult.', '');
              obj[key] = path.split('.').reduce((o: any, k: string) => (o || {})[k], lastResult);
            }
          } else if (val.includes('{{')) {
            // String interpolation
            obj[key] = val.replace(/\{\{([\w\d\.\[\]]+)\}\}/g, (_: string, varName: string) => {
              if (varName === 'lastResult') return JSON.stringify(lastResult);
              if (env[varName] !== undefined) return String(env[varName]);
              if (varName.startsWith('lastResult.')) {
                const path = varName.replace('lastResult.', '');
                const res = path.split('.').reduce((o: any, k: string) => (o || {})[k], lastResult);
                return res !== undefined ? String(res) : '';
              }
              return '';
            });
          }
        } else if (typeof obj[key] === 'object' && obj[key] !== null) {
          walk(obj[key]);
        }
      }
    };

    walk(processed);
    return processed;
  }

  private evaluateCondition(expression: string, context: any): boolean {
    const expr = expression.trim();
    if (expr === 'true') return true;
    if (expr === 'false') return false;

    // Matches: path operator value
    const match = expr.match(/^([\w\d\.\[\]]+)\s*(===|==|!==|!=|>|<|>=|<=)\s*(.+)$/);

    if (!match) return false;

    const [_, path, op, rightStr] = match;

    // Resolve left side from context
    let leftVal = context;
    const parts = path.split('.');
    for (const part of parts) {
      if (leftVal === undefined || leftVal === null) break;
      leftVal = leftVal[part];
    }

    // Parse right side
    let rightVal: any = rightStr.trim();
    if ((rightVal.startsWith("'") && rightVal.endsWith("'")) || (rightVal.startsWith('"') && rightVal.endsWith('"'))) {
      rightVal = rightVal.slice(1, -1);
    } else if (rightVal === 'true') {
      rightVal = true;
    } else if (rightVal === 'false') {
      rightVal = false;
    } else if (!isNaN(Number(rightVal)) && rightVal !== '') {
      rightVal = Number(rightVal);
    }

    switch (op) {
      case '==': return leftVal == rightVal;
      case '===': return leftVal === rightVal;
      case '!=': return leftVal != rightVal;
      case '!==': return leftVal !== rightVal;
      case '>': return leftVal > rightVal;
      case '<': return leftVal < rightVal;
      case '>=': return leftVal >= rightVal;
      case '<=': return leftVal <= rightVal;
    }

    return false;
  }
}
