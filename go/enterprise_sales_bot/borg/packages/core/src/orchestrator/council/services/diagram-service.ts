import type { TaskPlan, SubTask } from './types.js';

class DiagramService {
  /**
   * Generates a Mermaid flowchart representing a TaskPlan and its dependencies.
   */
  generateSwarmMermaid(plan: TaskPlan): string {
    let mermaid = 'graph TD\n';
    
    // Add nodes
    for (const task of plan.subtasks) {
      const toolLabel = task.preferredCLI ? ` [${task.preferredCLI}]` : '';
      const escapedTitle = task.title.replace(/"/g, "'");
      mermaid += `  ${task.id}["${escapedTitle}${toolLabel}"]\n`;
      
      // Color coding based on status
      if (task.status === 'completed') {
        mermaid += `  style ${task.id} fill:#d4edda,stroke:#28a745\n`;
      } else if (task.status === 'in_progress') {
        mermaid += `  style ${task.id} fill:#fff3cd,stroke:#ffc107\n`;
      } else if (task.status === 'failed') {
        mermaid += `  style ${task.id} fill:#f8d7da,stroke:#dc3545\n`;
      }
    }

    // Add edges
    for (const task of plan.subtasks) {
      for (const depId of task.dependencies) {
        mermaid += `  ${depId} --> ${task.id}\n`;
      }
    }

    return mermaid;
  }

  /**
   * Generates a system architecture diagram based on the current CLI registry and active sessions.
   */
  generateSystemMermaid(availableTools: string[], activeSessions: Array<{id: string, cliType: string}>): string {
    let mermaid = 'graph LR\n';
    
    mermaid += '  subgraph Council["Supreme Council"]\n';
    mermaid += '    Orchestrator((Orchestrator))\n';
    mermaid += '    Analytics[(Analytics DB)]\n';
    mermaid += '    Checkpoints[(Checkpoints)]\n';
    mermaid += '  end\n\n';

    mermaid += '  subgraph Tools["AI Tool Fleet"]\n';
    for (const tool of availableTools) {
      mermaid += `    ${tool}[${tool}]\n`;
    }
    mermaid += '  end\n\n';

    mermaid += '  subgraph Runtime["Active Swarm Sessions"]\n';
    for (const session of activeSessions) {
      mermaid += `    ${session.id.replace(/-/g, '_')}{${session.cliType}}\n`;
    }
    mermaid += '  end\n\n';

    // Connections
    mermaid += '  Orchestrator --> Analytics\n';
    mermaid += '  Orchestrator --> Checkpoints\n';
    
    for (const session of activeSessions) {
      mermaid += `  Orchestrator -.-> ${session.id.replace(/-/g, '_')}\n`;
      mermaid += `  ${session.id.replace(/-/g, '_')} --- ${session.cliType}\n`;
    }

    return mermaid;
  }

  /**
   * Parses a simple Mermaid graph definition into a TaskPlan structure.
   * Supports 'id["Title"]' nodes and 'id1 --> id2' dependencies.
   */
  parseMermaidToPlan(mermaid: string): Partial<TaskPlan> {
    const lines = mermaid.split('\n');
    const subtasks: SubTask[] = [];
    const idMap = new Map<string, SubTask>();

    for (const line of lines) {
      const trimmed = line.trim();
      
      // Parse nodes: id["Title"] or id["Title [tool]"]
      const nodeMatch = trimmed.match(/^(\w+)\["([^\]"]+)(?:\s+\[(\w+)\])?"\]$/);
      if (nodeMatch) {
        const [_, id, title, tool] = nodeMatch;
        const task: SubTask = {
          id,
          title,
          description: title, // Default description to title
          dependencies: [],
          status: 'pending',
          preferredCLI: tool as any,
        };
        idMap.set(id, task);
        subtasks.push(task);
        continue;
      }

      // Parse edges: id1 --> id2
      const edgeMatch = trimmed.match(/^(\w+)\s+-->\s+(\w+)$/);
      if (edgeMatch) {
        const [_, fromId, toId] = edgeMatch;
        const toTask = idMap.get(toId);
        if (toTask) {
          toTask.dependencies.push(fromId);
        }
      }
    }

    return {
      subtasks,
      reasoning: 'Generated from Mermaid diagram architecture.',
    };
  }
}

export const diagramService = new DiagramService();
