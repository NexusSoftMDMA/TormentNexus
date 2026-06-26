"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Globe, Plus, Trash2, RefreshCw, Zap } from 'lucide-react';

export function NodeManager() {
  const [nodes, setNodes] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [newNode, setNewNode] = useState({ name: '', url: '', token: '' });

  useEffect(() => {
    fetchNodes();
  }, []);

  const fetchNodes = () => {
    setLoading(true);
    fetch('/api/council-nodes')
      .then(res => res.json())
      .then(data => {
        setNodes(data.nodes || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch nodes:', err);
        setLoading(false);
      });
  };

  const handleAddNode = async () => {
    if (!newNode.name || !newNode.url) return;
    try {
      const res = await fetch('/api/council-nodes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newNode)
      });
      if (res.ok) {
        setNewNode({ name: '', url: '', token: '' });
        fetchNodes();
      }
    } catch (err) {
      console.error('Failed to add node:', err);
    }
  };

  const handleDeleteNode = async (id: string) => {
    try {
      const res = await fetch(`/api/council-nodes/${id}`, { method: 'DELETE' });
      if (res.ok) fetchNodes();
    } catch (err) {
      console.error('Failed to delete node:', err);
    }
  };

  const handlePingNode = async (id: string) => {
    try {
      await fetch(`/api/council-nodes/${id}/ping`, { method: 'POST' });
      fetchNodes();
    } catch (err) {
      console.error('Failed to ping node:', err);
    }
  };

  return (
    <div className="space-y-6">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-lg text-slate-50 font-bold flex items-center gap-2">
            <Globe className="h-5 w-5 text-blue-400" />
            Distributed Council Nodes
          </CardTitle>
          <CardDescription className="text-slate-400">
            Register and manage remote TORMENTNEXUS instances to form a distributed supervisor network.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <Input 
              placeholder="Node Name (e.g. EU-West)" 
              value={newNode.name}
              onChange={e => setNewNode({...newNode, name: e.target.value})}
              className="bg-slate-950 border-slate-800"
            />
            <Input 
              placeholder="URL (https://...)" 
              value={newNode.url}
              onChange={e => setNewNode({...newNode, url: e.target.value})}
              className="bg-slate-950 border-slate-800"
            />
            <Input 
              placeholder="Auth Token (Optional)" 
              value={newNode.token}
              onChange={e => setNewNode({...newNode, token: e.target.value})}
              className="bg-slate-950 border-slate-800"
            />
            <Button onClick={handleAddNode} className="bg-blue-600 hover:bg-blue-500">
              <Plus className="h-4 w-4 mr-2" />
              Add Node
            </Button>
          </div>

          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">Node Name</TableHead>
                <TableHead className="text-slate-400">Endpoint</TableHead>
                <TableHead className="text-slate-400">Status</TableHead>
                <TableHead className="text-slate-400 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-10 text-slate-500">Loading nodes...</TableCell>
                </TableRow>
              ) : nodes.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-10 text-slate-500">No remote nodes registered.</TableCell>
                </TableRow>
              ) : nodes.map((node) => (
                <TableRow key={node.id} className="border-slate-800 hover:bg-white/5">
                  <TableCell className="font-bold text-slate-50">{node.name}</TableCell>
                  <TableCell className="text-slate-400 font-mono text-xs">{node.url}</TableCell>
                  <TableCell>
                    <Badge className={
                      node.status === 'online' ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' : 
                      node.status === 'offline' ? 'bg-red-500/10 text-red-500 border-red-500/20' :
                      'bg-slate-500/10 text-slate-500 border-slate-500/20'
                    }>
                      {node.status.toUpperCase()}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" onClick={() => handlePingNode(node.id)} className="h-8 w-8 text-slate-400 hover:text-white">
                        <RefreshCw className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => handleDeleteNode(node.id)} className="h-8 w-8 text-slate-400 hover:text-red-400">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="bg-slate-950 border-slate-800 border-dashed">
          <CardContent className="py-10 flex flex-col items-center justify-center text-center space-y-4">
            <div className="p-3 rounded-full bg-blue-500/10 text-blue-500">
              <Zap className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-slate-50 font-bold">Auto-Discovery</h3>
              <p className="text-sm text-slate-500 max-w-xs">
                Automatically find and join existing TORMENTNEXUS supervisor mesh networks on your local subnet.
              </p>
            </div>
            <Button variant="outline" className="border-slate-800 text-slate-400" disabled>
              Scan Subnet
            </Button>
          </CardContent>
        </Card>
        
        <Card className="bg-slate-950 border-slate-800 border-dashed">
          <CardContent className="py-10 flex flex-col items-center justify-center text-center space-y-4">
            <div className="p-3 rounded-full bg-purple-500/10 text-purple-500">
              <Globe className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-slate-50 font-bold">Peer-to-Peer Relay</h3>
              <p className="text-sm text-slate-500 max-w-xs">
                Enable P2P relaying for nodes behind restrictive NAT or firewalls.
              </p>
            </div>
            <Button variant="outline" className="border-slate-800 text-slate-400" disabled>
              Configure Relay
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
