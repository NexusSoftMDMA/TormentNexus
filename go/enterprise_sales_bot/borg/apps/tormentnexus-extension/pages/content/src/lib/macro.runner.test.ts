import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MacroRunner } from './macro.runner';
import type { Macro } from '../stores';
import type { Node, Edge } from '@xyflow/react';

describe('MacroRunner (Graph)', () => {
    let sendMessageMock: any;
    let onLogMock: any;
    let runner: MacroRunner;

    beforeEach(() => {
        sendMessageMock = vi.fn().mockResolvedValue({ status: 'success' });
        onLogMock = vi.fn();
        runner = new MacroRunner(sendMessageMock, onLogMock);
    });

    it('executes a linear graph of tool nodes correctly', async () => {
        const nodes: Node[] = [
            { id: 'node1', type: 'tool', data: { toolName: 'toolA', args: { a: 1 } }, position: { x: 0, y: 0 } },
            { id: 'node2', type: 'tool', data: { toolName: 'toolB', args: { b: 2 } }, position: { x: 0, y: 0 } },
        ];
        const edges: Edge[] = [
            { id: 'e1-2', source: 'node1', target: 'node2' }
        ];

        const macro: Macro = { id: 'm1', name: 'Test', description: '', nodes, edges, createdAt: 0, updatedAt: 0 };

        await runner.run(macro);

        expect(sendMessageMock).toHaveBeenCalledTimes(2);
        expect(sendMessageMock).toHaveBeenNthCalledWith(1, 'toolA', { a: 1 });
        expect(sendMessageMock).toHaveBeenNthCalledWith(2, 'toolB', { b: 2 });
    });

    it('evaluates conditions and follows correct branches', async () => {
        // node1 (tool) -> node2 (condition) --(true)--> node3 (toolX)
        //                                   --(false)-> node4 (toolY)

        sendMessageMock.mockResolvedValueOnce({ value: 10 }); // toolA returns 10

        const nodes: Node[] = [
            { id: 'n1', type: 'tool', data: { toolName: 'toolA', args: {} }, position: { x: 0, y: 0 } },
            { id: 'n2', type: 'condition', data: { expression: 'lastResult.value === 10' }, position: { x: 0, y: 0 } },
            { id: 'n3', type: 'tool', data: { toolName: 'toolTrue', args: {} }, position: { x: 0, y: 0 } },
            { id: 'n4', type: 'tool', data: { toolName: 'toolFalse', args: {} }, position: { x: 0, y: 0 } },
        ];

        const edges: Edge[] = [
            { id: 'e1-2', source: 'n1', target: 'n2' },
            { id: 'e2-3', source: 'n2', target: 'n3', sourceHandle: 'true' },
            { id: 'e2-4', source: 'n2', target: 'n4', sourceHandle: 'false' },
        ];

        const macro: Macro = { id: 'm2', name: 'Condition Test', description: '', nodes, edges, createdAt: 0, updatedAt: 0 };

        await runner.run(macro);

        expect(sendMessageMock).toHaveBeenCalledTimes(2);
        expect(sendMessageMock).toHaveBeenNthCalledWith(1, 'toolA', {});
        expect(sendMessageMock).toHaveBeenNthCalledWith(2, 'toolTrue', {});
    });

    it('substitutes variables in tool arguments correctly', async () => {
        sendMessageMock.mockResolvedValueOnce({ id: 123 });

        const nodes: Node[] = [
            { id: 'n1', type: 'tool', data: { toolName: 'tool1', args: {} }, position: { x: 0, y: 0 } },
            { id: 'n2', type: 'set_variable', data: { variableName: 'myVar', variableValue: 'lastResult.id' }, position: { x: 0, y: 0 } },
            { id: 'n3', type: 'tool', data: { toolName: 'tool2', args: { targetId: '{{myVar}}', fullRes: '{{lastResult}}' } }, position: { x: 0, y: 0 } },
        ];

        const edges: Edge[] = [
            { id: 'e1-2', source: 'n1', target: 'n2' },
            { id: 'e2-3', source: 'n2', target: 'n3' }
        ];

        const macro: Macro = { id: 'm3', name: 'Var Test', description: '', nodes, edges, createdAt: 0, updatedAt: 0 };

        await runner.run(macro);

        expect(sendMessageMock).toHaveBeenCalledTimes(2);
        expect(sendMessageMock).toHaveBeenNthCalledWith(2, 'tool2', {
            targetId: 123,
            fullRes: { id: 123 }
        });
    });
});
