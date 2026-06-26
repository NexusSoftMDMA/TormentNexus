"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { Database, HardDrive, Cloud, ShieldCheck, Save } from 'lucide-react';

export function DataResidencyConfig() {
  const [policies, setPolicies] = useState<any[]>([]);
  const [providers, setProviders] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [pRes, prRes] = await Promise.all([
        fetch('/api/data-residency/policies'),
        fetch('/api/data-residency/providers')
      ]);
      const pData = await pRes.json();
      const prData = await prRes.json();
      setPolicies(pData.policies || []);
      setProviders(prData.providers || []);
    } catch (err) {
      console.error('Failed to fetch residency data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateProvider = async (dataType: string, provider: string) => {
    try {
      await fetch(`/api/data-residency/policies/${dataType}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ provider })
      });
      fetchData();
    } catch (err) {
      console.error('Update failed:', err);
    }
  };

  const getProviderIcon = (provider: string) => {
    switch (provider) {
      case 'local': return <HardDrive className="h-4 w-4" />;
      case 's3': 
      case 'azure-blob': return <Cloud className="h-4 w-4" />;
      default: return <Database className="h-4 w-4" />;
    }
  };

  return (
    <div className="space-y-6">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-lg text-slate-50 font-bold flex items-center gap-2">
            <Database className="h-5 w-5 text-amber-400" />
            Data Residency & Storage
          </CardTitle>
          <CardDescription className="text-slate-400">
            Configure storage locations and compliance rules for different types of enterprise data.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">Data Type</TableHead>
                <TableHead className="text-slate-400">Storage Provider</TableHead>
                <TableHead className="text-slate-400">Security</TableHead>
                <TableHead className="text-slate-400 text-right">Settings</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center py-10 text-slate-500">Loading residency policies...</TableCell>
                </TableRow>
              ) : policies.map((policy) => (
                <TableRow key={policy.dataType} className="border-slate-800 hover:bg-white/5">
                  <TableCell className="font-bold text-slate-50 capitalize">{policy.dataType}</TableCell>
                  <TableCell>
                    <Select 
                      defaultValue={policy.provider} 
                      onValueChange={(val) => handleUpdateProvider(policy.dataType, val)}
                    >
                      <SelectTrigger className="w-[160px] h-8 bg-slate-950 border-slate-800 text-xs">
                        <div className="flex items-center gap-2">
                          {getProviderIcon(policy.provider)}
                          <SelectValue />
                        </div>
                      </SelectTrigger>
                      <SelectContent className="bg-slate-900 border-slate-800">
                        {providers.map(p => (
                          <SelectItem key={p} value={p} className="text-xs uppercase">
                            {p}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {policy.encryption && (
                        <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20 text-[10px]">
                          <ShieldCheck className="h-3 w-3 mr-1" /> ENCRYPTED
                        </Badge>
                      )}
                      {policy.retentionDays && (
                        <span className="text-[10px] text-slate-500">{policy.retentionDays}d Retention</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm" className="h-8 text-xs text-slate-400">Configure</Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-bold text-slate-50">Storage Encryption</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-xs text-slate-400">All data stored in cloud providers is AES-256 encrypted by default using system-managed keys.</p>
            <Button variant="outline" size="sm" className="w-full border-slate-800 text-xs">Manage Keys</Button>
          </CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-bold text-slate-50">Cross-Region Sync</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-xs text-slate-400">Enable automatic synchronization of memories across distributed nodes for high availability.</p>
            <Button variant="outline" size="sm" className="w-full border-slate-800 text-xs">Enable Sync</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
