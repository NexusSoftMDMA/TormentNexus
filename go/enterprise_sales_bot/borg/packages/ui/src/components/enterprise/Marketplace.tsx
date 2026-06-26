"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '../ui/card';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs';
import { ShoppingBag, Search, Download, ExternalLink, Bot, Zap, Package } from 'lucide-react';

export function Marketplace() {
  const [packages, setPackages] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [installing, setInstalling] = useState<string | null>(null);

  useEffect(() => {
    fetchPackages();
  }, []);

  const fetchPackages = () => {
    setLoading(true);
    fetch('/api/state')
      .then(res => res.json())
      .then(data => {
        setPackages(data.marketplace || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch marketplace:', err);
        setLoading(false);
      });
  };

  const handleInstall = async (name: string) => {
    setInstalling(name);
    try {
      const res = await fetch('/api/marketplace/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name })
      });
      if (res.ok) {
        const data = await res.json();
        alert(`Successfully installed: ${name}`);
      } else {
        const err = await res.json();
        alert(`Installation failed: ${err.error}`);
      }
    } catch (err) {
      console.error('Install error:', err);
    } finally {
      setInstalling(null);
    }
  };

  const filteredPackages = packages.filter(p => 
    p.name.toLowerCase().includes(search.toLowerCase()) || 
    p.description.toLowerCase().includes(search.toLowerCase())
  );

  const getIcon = (type: string) => {
    switch (type) {
      case 'agent': return <Bot className="h-4 w-4" />;
      case 'workflow': return <Zap className="h-4 w-4" />;
      case 'skill': return <Package className="h-4 w-4" />;
      default: return <ShoppingBag className="h-4 w-4" />;
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'agent': return 'bg-purple-500/10 text-purple-500 border-purple-500/20';
      case 'workflow': return 'bg-blue-500/10 text-blue-500 border-blue-500/20';
      case 'skill': return 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20';
      default: return 'bg-slate-500/10 text-slate-500 border-slate-500/20';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-50 flex items-center gap-2">
            <ShoppingBag className="h-6 w-6 text-emerald-400" />
            TORMENTNEXUS Marketplace
          </h2>
          <p className="text-slate-400 text-sm">Discover and install autonomous agents, skills, and enterprise workflows.</p>
        </div>
        <div className="relative w-full md:w-72">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
          <Input 
            placeholder="Search ecosystem..." 
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-9 bg-slate-950 border-slate-800"
          />
        </div>
      </div>

      <Tabs defaultValue="all" className="space-y-6">
        <TabsList className="bg-slate-900 border-slate-800">
          <TabsTrigger value="all">All Packages</TabsTrigger>
          <TabsTrigger value="agent">Agents</TabsTrigger>
          <TabsTrigger value="workflow">Workflows</TabsTrigger>
          <TabsTrigger value="skill">Skills</TabsTrigger>
        </TabsList>

        <TabsContent value="all" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {loading ? (
            <div className="col-span-full py-20 text-center text-slate-500">Loading marketplace...</div>
          ) : filteredPackages.length === 0 ? (
            <div className="col-span-full py-20 text-center text-slate-500">No matching packages found.</div>
          ) : filteredPackages.map(pkg => (
            <Card key={pkg.name} className="bg-slate-900 border-slate-800 flex flex-col">
              <CardHeader>
                <div className="flex justify-between items-start mb-2">
                  <Badge variant="outline" className={`flex items-center gap-1.5 uppercase text-[10px] ${getTypeColor(pkg.type)}`}>
                    {getIcon(pkg.type)}
                    {pkg.type}
                  </Badge>
                  {pkg.provider === 'official' && (
                    <Badge className="bg-blue-500/20 text-blue-400 text-[10px] border-none">OFFICIAL</Badge>
                  )}
                </div>
                <CardTitle className="text-slate-50">{pkg.name}</CardTitle>
                <CardDescription className="text-slate-400 text-xs line-clamp-2 h-8">
                  {pkg.description}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1">
                {pkg.metadata && (
                  <div className="flex flex-wrap gap-2">
                    {pkg.metadata.stepsCount && (
                      <span className="text-[10px] text-slate-500">{pkg.metadata.stepsCount} steps</span>
                    )}
                    {pkg.metadata.author && (
                      <span className="text-[10px] text-slate-500">by {pkg.metadata.author}</span>
                    )}
                  </div>
                )}
              </CardContent>
              <CardFooter className="border-t border-slate-800 pt-4 flex gap-2">
                <Button 
                  onClick={() => handleInstall(pkg.name)} 
                  disabled={installing === pkg.name}
                  className="flex-1 bg-emerald-600 hover:bg-emerald-500 h-8 text-xs"
                >
                  {installing === pkg.name ? 'Installing...' : 'Install'}
                  <Download className="h-3 w-3 ml-2" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-slate-500">
                  <ExternalLink className="h-3 w-3" />
                </Button>
              </CardFooter>
            </Card>
          ))}
        </TabsContent>
        
        {/* Type specific contents would follow similar pattern with filter */}
      </Tabs>
    </div>
  );
}
