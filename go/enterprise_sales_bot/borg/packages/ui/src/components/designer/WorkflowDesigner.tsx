"use client";

import React, { useState, useCallback, useEffect } from 'react';
import ReactFlow, { 
  addEdge, 
  Background, 
  Controls, 
  MiniMap,
  applyEdgeChanges,
  applyNodeChanges,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type OnNodesChange,
  type OnEdgesChange,
  type OnConnect
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { Save, Plus, Play, Trash2, Settings2, Layout } from 'lucide-react';

const initialNodes: Node[] = [
  { 
    id: 'start', 
    data: { label: 'Start Trigger' }, 
    position: { x: 250, y: 0 },
    type: 'input',
    style: { background: '#1e293b', color: '#fff', border: '1px solid #334155' }
  }
];

export function WorkflowDesigner() {
  const [workflows, setWorkflows] = useState<any[]>([]);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [name, setName] = useState('New Workflow');

  useEffect(() => {
    fetchWorkflows();
  }, []);

  const fetchWorkflows = async () => {
    try {
      const res = await fetch('/api/workflows');
      const data = await res.json();
      setWorkflows(data.workflows || []);
    } catch (err) {
      console.error('Failed to fetch workflows:', err);
    }
  };

  const loadWorkflow = async (id: string) => {
    try {
      const res = await fetch(`/api/workflows/${id}`);
      const data = await res.json();
      if (data.success) {
        setSelectedWorkflowId(id);
        setName(data.workflow.name);
        if (data.workflow.uiConfig) {
          setNodes(data.workflow.uiConfig.nodes || initialNodes);
          setEdges(data.workflow.uiConfig.edges || []);
        } else {
          setNodes(initialNodes);
          setEdges([]);
        }
      }
    } catch (err) {
      console.error('Failed to load workflow:', err);
    }
  };

  const onConnect: OnConnect = useCallback(
    (connection) => setEdges((eds) => addEdge(connection, eds)),
    [setEdges]
  );

  const handleSave = async () => {
    const payload = {
      name,
      uiConfig: { nodes, edges },
      status: 'active'
    };

    try {
      let res;
      if (selectedWorkflowId) {
        res = await fetch(`/api/workflows/${selectedWorkflowId}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
      } else {
        res = await fetch('/api/workflows', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
      }
      
      const data = await res.json();
      if (data.success) {
        alert('Workflow saved successfully!');
        if (!selectedWorkflowId) setSelectedWorkflowId(data.workflow.id);
        fetchWorkflows();
      }
    } catch (err) {
      console.error('Failed to save workflow:', err);
    }
  };

  const addNode = (type: string) => {
    const id = `${type}-${Date.now()}`;
    const newNode: Node = {
      id,
      data: { label: `${type.toUpperCase()}: New Node` },
      position: { x: Math.random() * 400, y: Math.random() * 400 },
      style: { 
        background: type === 'agent' ? '#4c1d95' : type === 'tool' ? '#064e3b' : '#1e293b', 
        color: '#fff', 
        border: '1px solid #334155' 
      }
    };
    setNodes((nds) => nds.concat(newNode));
  };

  return (
    <div className="flex flex-col h-[700px] w-full">
      <div className="flex items-center justify-between p-4 bg-slate-900 border-b border-slate-800 rounded-t-lg">
        <div className="flex items-center gap-4">
          <Select onValueChange={loadWorkflow} value={selectedWorkflowId || undefined}>
            <SelectTrigger className="w-[200px] bg-slate-950 border-slate-800">
              <SelectValue placeholder="Select Workflow" />
            </SelectTrigger>
            <SelectContent className="bg-slate-900 border-slate-800 text-slate-50">
              {workflows.map(wf => (
                <SelectItem key={wf.id} value={wf.id}>{wf.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Input 
            value={name} 
            onChange={e => setName(e.target.value)} 
            className="w-[200px] bg-slate-950 border-slate-800"
          />
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => addNode('agent')} className="border-slate-700 text-xs">
            <Plus className="h-3 w-3 mr-1" /> Agent
          </Button>
          <Button variant="outline" size="sm" onClick={() => addNode('tool')} className="border-slate-700 text-xs">
            <Plus className="h-3 w-3 mr-1" /> Tool
          </Button>
          <Button variant="outline" size="sm" onClick={() => addNode('council')} className="border-slate-700 text-xs">
            <Plus className="h-3 w-3 mr-1" /> Council
          </Button>
          <div className="w-px h-4 bg-slate-800 mx-2" />
          <Button size="sm" onClick={handleSave} className="bg-emerald-600 hover:bg-emerald-500 h-8 text-xs">
            <Save className="h-3 w-3 mr-1" /> Save
          </Button>
          <Button size="sm" variant="secondary" className="h-8 text-xs">
            <Play className="h-3 w-3 mr-1" /> Execute
          </Button>
        </div>
      </div>

      <div className="flex-1 bg-slate-950 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
        >
          <Background color="#1e293b" gap={20} />
          <Controls className="bg-slate-900 border-slate-800 fill-slate-400" />
          <MiniMap className="bg-slate-900 border-slate-800" nodeColor="#334155" />
        </ReactFlow>
        
        {/* Help Overlay */}
        <div className="absolute bottom-4 right-4 p-3 bg-slate-900/80 backdrop-blur border border-slate-800 rounded text-[10px] text-slate-500 pointer-events-none">
          <p>Drag to move • Connect ports to link • Delete key to remove</p>
        </div>
      </div>
    </div>
  );
}
