"use client";

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@tormentnexus/ui';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@tormentnexus/ui';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@tormentnexus/ui';
import { Badge } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';

export type UserRole = 'admin' | 'developer' | 'operator' | 'viewer';

export function RbacManager() {
  const [roles, setRoles] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [userLoading, setUserLoading] = useState(true);

  useEffect(() => {
    // Fetch roles
    fetch('/api/rbac/roles')
      .then(res => res.json())
      .then(data => {
        setRoles(data.roles || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch roles:', err);
        setLoading(false);
      });

    // Fetch users
    fetchUsers();
  }, []);

  const fetchUsers = () => {
    setUserLoading(true);
    fetch('/api/rbac/users')
      .then(res => res.json())
      .then(data => {
        setUsers(data.users || []);
        setUserLoading(false);
      })
      .catch(err => {
        console.error('Failed to fetch users:', err);
        setUserLoading(false);
      });
  };

  const handleUpdateRole = async (userId: string, role: string) => {
    try {
      const res = await fetch(`/api/rbac/users/${userId}/role`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role })
      });
      if (res.ok) {
        fetchUsers();
      } else {
        const error = await res.json();
        alert(`Failed to update role: ${error.error}`);
      }
    } catch (err) {
      console.error('Failed to update role:', err);
    }
  };

  return (
    <div className="space-y-6">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-lg text-slate-50 font-bold">System Roles</CardTitle>
          <CardDescription className="text-slate-400">
            Current role definitions and assigned permissions.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">Role</TableHead>
                <TableHead className="text-slate-400">Description</TableHead>
                <TableHead className="text-slate-400">Permissions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={3} className="text-center py-10 text-slate-500">Loading roles...</TableCell>
                </TableRow>
              ) : roles.map((role) => (
                <TableRow key={role.role} className="border-slate-800 hover:bg-white/5">
                  <TableCell className="font-bold text-slate-50 uppercase tracking-tight">
                    {role.role}
                  </TableCell>
                  <TableCell className="text-slate-400 text-sm">
                    {role.description}
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {role.permissions.map((p: string) => (
                        <Badge key={p} variant="outline" className="text-[10px] bg-slate-950 border-slate-700 text-slate-300">
                          {p}
                        </Badge>
                      ))}
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
          <CardTitle className="text-lg text-slate-50 font-bold">User Access Management</CardTitle>
          <CardDescription className="text-slate-400">
            Assign roles to individual users or API keys.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader className="border-slate-800">
              <TableRow className="hover:bg-transparent border-slate-800">
                <TableHead className="text-slate-400">User ID</TableHead>
                <TableHead className="text-slate-400">Current Role</TableHead>
                <TableHead className="text-slate-400">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {userLoading ? (
                <TableRow>
                  <TableCell colSpan={3} className="text-center py-10 text-slate-500">Loading users...</TableCell>
                </TableRow>
              ) : users.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={3} className="text-center py-10 text-slate-500">
                    No users found. Login via OIDC to register users.
                  </TableCell>
                </TableRow>
              ) : users.map((user) => (
                <TableRow key={user.userId} className="border-slate-800 hover:bg-white/5">
                  <TableCell className="font-mono text-xs text-slate-300">
                    {user.userId}
                  </TableCell>
                  <TableCell>
                    <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20 uppercase">
                      {user.role}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Select 
                      defaultValue={user.role} 
                      onValueChange={(val) => handleUpdateRole(user.userId, val)}
                    >
                      <SelectTrigger className="w-[140px] h-8 bg-slate-950 border-slate-800 text-xs">
                        <SelectValue placeholder="Change role" />
                      </SelectTrigger>
                      <SelectContent className="bg-slate-900 border-slate-800">
                        {roles.map(r => (
                          <SelectItem key={r.role} value={r.role} className="text-xs uppercase">
                            {r.role}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          
          <div className="mt-6 flex items-center gap-4 p-4 rounded bg-slate-950 border border-slate-800 opacity-50">
            <div className="flex-1 space-y-1">
              <div className="text-sm font-medium text-slate-300">Global Admin Token</div>
              <div className="text-xs text-slate-500 font-mono italic">SUPER_AI_TOKEN</div>
            </div>
            <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">ADMIN</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
