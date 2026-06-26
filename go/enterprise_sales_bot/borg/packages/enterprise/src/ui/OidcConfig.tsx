"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@tormentnexus/ui';
import { Badge } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { Input } from '@tormentnexus/ui';
import { Switch } from '@tormentnexus/ui';
import { Label } from '@tormentnexus/ui';
import { Shield, Plus, Trash2, Key, Settings } from 'lucide-react';

export function OidcConfig() {
  const [providers, setProviders] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [newProvider, setNewProvider] = useState({
    name: '',
    issuerUrl: '',
    clientId: '',
    clientSecret: '',
    scopes: ['openid', 'profile', 'email'],
    enabled: true
  });

  useEffect(() => {
    fetchProviders();
  }, []);

  const fetchProviders = () => {
    setLoading(true);
    fetch('/api/oidc/providers')
      .then(res => res.json())
      .then(data => {
        setProviders(data.providers || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch providers:', err);
        setLoading(false);
      });
  };

  const handleAddProvider = async () => {
    try {
      const res = await fetch('/api/oidc/providers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newProvider)
      });
      if (res.ok) {
        setShowAdd(false);
        setNewProvider({
          name: '',
          issuerUrl: '',
          clientId: '',
          clientSecret: '',
          scopes: ['openid', 'profile', 'email'],
          enabled: true
        });
        fetchProviders();
      }
    } catch (err) {
      console.error('Failed to add provider:', err);
    }
  };

  const handleDeleteProvider = async (id: string) => {
    try {
      const res = await fetch(`/api/oidc/providers/${id}`, { method: 'DELETE' });
      if (res.ok) fetchProviders();
    } catch (err) {
      console.error('Failed to delete provider:', err);
    }
  };

  const handleToggleEnabled = async (id: string, enabled: boolean) => {
    try {
      await fetch(`/api/oidc/providers/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled })
      });
      fetchProviders();
    } catch (err) {
      console.error('Failed to toggle provider:', err);
    }
  };

  return (
    <div className="space-y-6">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-lg text-slate-50 font-bold flex items-center gap-2">
              <Shield className="h-5 w-5 text-purple-400" />
              SSO & Identity Providers
            </CardTitle>
            <CardDescription className="text-slate-400">
              Configure OIDC and SAML providers for enterprise single sign-on.
            </CardDescription>
          </div>
          <Button onClick={() => setShowAdd(!showAdd)} variant="outline" className="border-slate-700">
            <Plus className="h-4 w-4 mr-2" />
            Add Provider
          </Button>
        </CardHeader>
        <CardContent>
          {showAdd && (
            <div className="p-4 mb-6 rounded bg-slate-950 border border-slate-800 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Provider Name</Label>
                  <Input 
                    placeholder="GitHub / Google / Okta" 
                    value={newProvider.name}
                    onChange={e => setNewProvider({...newProvider, name: e.target.value})}
                    className="bg-slate-900 border-slate-800"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Issuer URL</Label>
                  <Input 
                    placeholder="https://auth.example.com" 
                    value={newProvider.issuerUrl}
                    onChange={e => setNewProvider({...newProvider, issuerUrl: e.target.value})}
                    className="bg-slate-900 border-slate-800"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Client ID</Label>
                  <Input 
                    placeholder="client_id_123" 
                    value={newProvider.clientId}
                    onChange={e => setNewProvider({...newProvider, clientId: e.target.value})}
                    className="bg-slate-900 border-slate-800"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Client Secret</Label>
                  <Input 
                    type="password"
                    placeholder="••••••••" 
                    value={newProvider.clientSecret}
                    onChange={e => setNewProvider({...newProvider, clientSecret: e.target.value})}
                    className="bg-slate-900 border-slate-800"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" onClick={() => setShowAdd(false)}>Cancel</Button>
                <Button onClick={handleAddProvider} className="bg-purple-600 hover:bg-purple-500">Save Provider</Button>
              </div>
            </div>
          )}

          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">Provider</TableHead>
                <TableHead className="text-slate-400">Client ID</TableHead>
                <TableHead className="text-slate-400">Status</TableHead>
                <TableHead className="text-slate-400 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-10 text-slate-500">Loading providers...</TableCell>
                </TableRow>
              ) : providers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-10 text-slate-500">No identity providers configured.</TableCell>
                </TableRow>
              ) : providers.map((p) => (
                <TableRow key={p.id} className="border-slate-800 hover:bg-white/5">
                  <TableCell>
                    <div className="flex flex-col">
                      <span className="font-bold text-slate-50">{p.name}</span>
                      <span className="text-[10px] text-slate-500 font-mono">{p.issuerUrl}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-400 font-mono text-xs">{p.clientId}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Switch 
                        checked={p.enabled} 
                        onCheckedChange={(checked) => handleToggleEnabled(p.id, checked)}
                      />
                      <Badge variant="outline" className={p.enabled ? 'text-emerald-500 border-emerald-500/20' : 'text-slate-500 border-slate-800'}>
                        {p.enabled ? 'ENABLED' : 'DISABLED'}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-400">
                        <Settings className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => handleDeleteProvider(p.id)} className="h-8 w-8 text-slate-400 hover:text-red-400">
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

      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-lg text-slate-50 font-bold flex items-center gap-2">
            <Key className="h-5 w-5 text-blue-400" />
            Security Policies
          </CardTitle>
          <CardDescription className="text-slate-400">
            Global authentication settings and session management.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between p-4 rounded bg-slate-950 border border-slate-800">
            <div className="space-y-0.5">
              <Label className="text-slate-200">Force SSO Login</Label>
              <p className="text-xs text-slate-500">Require users to authenticate via an external provider (disables local tokens).</p>
            </div>
            <Switch disabled />
          </div>
          <div className="flex items-center justify-between p-4 rounded bg-slate-950 border border-slate-800">
            <div className="space-y-0.5">
              <Label className="text-slate-200">Session Timeout</Label>
              <p className="text-xs text-slate-500">Duration of inactivity before a session is revoked (default: 1 hour).</p>
            </div>
            <Input className="w-24 bg-slate-900 border-slate-800 text-right" defaultValue="3600" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
