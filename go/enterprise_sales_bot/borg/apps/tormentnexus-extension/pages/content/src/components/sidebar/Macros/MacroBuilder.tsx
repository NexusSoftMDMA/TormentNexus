import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useMacroStore, type Macro } from '@src/stores';
import { Icon, Typography, Button, Input, Textarea } from '../ui';
import { useToastStore } from '@src/stores';
import {
  ReactFlow,
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  addEdge,
  Background,
  Controls,
  Connection,
  Edge,
  Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { ToolNode } from './nodes/ToolNode';
import { ConditionNode } from './nodes/ConditionNode';

const nodeTypes = {
  tool: ToolNode,
  condition: ConditionNode,
};

interface MacroBuilderProps {
  existingMacro?: Macro | null;
  onClose: () => void;
}

const BuilderCanvas = ({
  existingMacro,
  onClose,
  name, setName,
  description, setDescription,
  handleSaveMacro
}: any) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(existingMacro?.nodes || []);
  const [edges, setEdges, onEdgesChange] = useEdgesState(existingMacro?.edges || []);
  const [availableTools, setAvailableTools] = useState<any[]>([]);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  useEffect(() => {
    const tools = (window as any).availableTools || [];
    setAvailableTools(tools);
  }, []);

  const onConnect = useCallback(
    (params: Connection | Edge) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  const onAddNode = (type: 'tool' | 'condition') => {
    const newNode: Node = {
      id: crypto.randomUUID(),
      type,
      position: { x: 250, y: 100 + nodes.length * 100 },
      data: type === 'tool' ? { toolName: '', args: {} } : { expression: 'true' },
    };
    setNodes((nds) => nds.concat(newNode));
  };

  const onNodeClick = (_: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
  };

  const onPaneClick = () => {
    setSelectedNode(null);
  };

  const updateNodeData = (id: string, dataUpdates: any) => {
    setNodes((nds) =>
      nds.map((n) => {
        if (n.id === id) {
          n.data = { ...n.data, ...dataUpdates };
        }
        return n;
      })
    );
    if (selectedNode?.id === id) {
      setSelectedNode((prev: any) => ({ ...prev, data: { ...prev.data, ...dataUpdates } }));
    }
  };

  const saveWorkflow = () => {
    if (!name.trim()) return alert('Macro name is required');
    if (nodes.length === 0) return alert('Macro must have at least one node');

    handleSaveMacro(name, description, nodes, edges);
  };

  const deleteSelectedNode = () => {
    if (!selectedNode) return;
    setNodes(nds => nds.filter(n => n.id !== selectedNode.id));
    setEdges(eds => eds.filter(e => e.source !== selectedNode.id && e.target !== selectedNode.id));
    setSelectedNode(null);
  };

  return (
    <div className="flex flex-col h-full bg-slate-50 dark:bg-slate-900 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shrink-0 z-10">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={onClose}>
            <Icon name="arrow-left" size="sm" />
          </Button>
          <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100">
            {existingMacro ? 'Edit Macro Graph' : 'New Macro Graph'}
          </Typography>
        </div>
        <div className="flex items-center gap-2">
          <Button onClick={saveWorkflow} size="sm" className="bg-primary-600 hover:bg-primary-700 text-white">
            <Icon name="save" size="xs" className="mr-1" />
            Save
          </Button>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden relative">
        {/* Canvas Area */}
        <div className="flex-1 h-full" ref={reactFlowWrapper}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onPaneClick={onPaneClick}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background />
            <Controls />
          </ReactFlow>
        </div>

        {/* Sidebar Configuration */}
        <div className="w-80 h-full border-l border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 overflow-y-auto flex flex-col shrink-0">
          <div className="p-4 border-b border-slate-200 dark:border-slate-700 space-y-3">
            <div>
              <label className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-1 block">Name</label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Weekly Report Generator"
                className="w-full text-sm"
              />
            </div>
            <div>
              <label className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-1 block">Description</label>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe what this macro does..."
                className="w-full h-16 text-sm"
              />
            </div>
            <div className="flex gap-2 pt-2">
              <Button size="sm" variant="outline" className="flex-1" onClick={() => onAddNode('tool')}>
                <Icon name="wrench" size="xs" className="mr-1" /> Add Tool
              </Button>
              <Button size="sm" variant="outline" className="flex-1" onClick={() => onAddNode('condition')}>
                <Icon name="git-merge" size="xs" className="mr-1" /> Add Condition
              </Button>
            </div>
          </div>

          <div className="p-4 flex-1">
            <Typography variant="subtitle" className="font-semibold text-slate-800 dark:text-slate-200 mb-4">
              Node Configuration
            </Typography>

            {!selectedNode && (
              <div className="text-sm text-slate-500 text-center mt-10">
                Select a node on the canvas to configure it.
              </div>
            )}

            {selectedNode && selectedNode.type === 'tool' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-blue-600 uppercase tracking-wider bg-blue-50 px-2 py-1 rounded">Tool Node</span>
                  <Button variant="ghost" size="icon" onClick={deleteSelectedNode} className="h-6 w-6 text-red-500">
                    <Icon name="trash-2" size="xs" />
                  </Button>
                </div>

                <div>
                  <label className="text-xs font-medium text-slate-500 mb-1 block">Select Tool</label>
                  <select
                    className="w-full text-sm border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-900 p-2"
                    value={selectedNode.data?.toolName as string || ''}
                    onChange={(e) => updateNodeData(selectedNode.id, { toolName: e.target.value })}
                  >
                    <option value="">Select a tool...</option>
                    {availableTools.map(t => (
                      <option key={t.name} value={t.name}>{t.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-500 mb-1 block">Arguments (JSON)</label>
                  <Textarea
                    value={selectedNode.data?.args ? JSON.stringify(selectedNode.data.args, null, 2) : '{}'}
                    onChange={(e) => {
                      try {
                        const args = JSON.parse(e.target.value);
                        updateNodeData(selectedNode.id, { args });
                      } catch (err) {
                        // ignore parsing error while typing
                      }
                    }}
                    placeholder={'{ "arg": "value" }'}
                    className="font-mono text-xs h-32"
                  />
                </div>
              </div>
            )}

            {selectedNode && selectedNode.type === 'condition' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-amber-600 uppercase tracking-wider bg-amber-50 px-2 py-1 rounded">Condition Node</span>
                  <Button variant="ghost" size="icon" onClick={deleteSelectedNode} className="h-6 w-6 text-red-500">
                    <Icon name="trash-2" size="xs" />
                  </Button>
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-500 mb-1 block">JavaScript Expression</label>
                  <Input
                    value={selectedNode.data?.expression as string || ''}
                    onChange={(e) => updateNodeData(selectedNode.id, { expression: e.target.value })}
                    placeholder="lastResult.status === 'success'"
                    className="font-mono text-sm"
                  />
                  <p className="text-[10px] text-slate-400 mt-1">
                    Available: <code>lastResult</code>, <code>allResults</code>, <code>env</code>
                  </p>
                </div>
              </div>
            )}

            {/* Note: Variable and Delay nodes omitted for brevity, could be added similarly */}
          </div>
        </div>
      </div>
    </div>
  );
};

const MacroBuilder: React.FC<MacroBuilderProps> = ({ existingMacro, onClose }) => {
  const { addMacro, updateMacro } = useMacroStore();
  const { addToast } = useToastStore();
  const [name, setName] = useState(existingMacro?.name || '');
  const [description, setDescription] = useState(existingMacro?.description || '');

  const handleSaveMacro = (nameStr: string, descStr: string, nodes: Node[], edges: Edge[]) => {
    const macroData = {
      name: nameStr,
      description: descStr,
      nodes,
      edges
    };

    if (existingMacro) {
      updateMacro(existingMacro.id, macroData);
      addToast({ title: 'Saved', message: 'Macro flow updated', type: 'success' });
    } else {
      addMacro(macroData);
      addToast({ title: 'Saved', message: 'Macro flow created', type: 'success' });
    }
    onClose();
  };

  return (
    <ReactFlowProvider>
      <BuilderCanvas
        existingMacro={existingMacro}
        onClose={onClose}
        name={name}
        setName={setName}
        description={description}
        setDescription={setDescription}
        handleSaveMacro={handleSaveMacro}
      />
    </ReactFlowProvider>
  );
};

export default MacroBuilder;
