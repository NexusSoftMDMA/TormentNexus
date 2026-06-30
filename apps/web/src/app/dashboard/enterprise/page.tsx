"use client";

import { useState, useEffect, useCallback } from "react";
import { Shield, Key, Users, FileText, Loader2, RotateCcw, CheckCircle, XCircle, RefreshCw, Save } from "lucide-react";

interface LicenseInfo {
  valid: boolean;
  licensedTo?: string;
  expiresAt?: string;
  features?: string[];
  maxNodes?: number;
  tier?: string;
  ssoSettings?: Record<string, string>;
}

interface RBACRole {
  name: string;
  permissions: string[];
  description?: string;
}

export default function EnterprisePage() {
  const [license, setLicense] = useState<LicenseInfo | null>(null);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
  const [roles, setRoles] = useState<RBACRole[]>([]);
  const [loading, setLoading] = useState(true);

  // SSO form state
  const [providerUrl, setProviderUrl] = useState("");
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [ssoSaving, setSsoSaving] = useState(false);
  const [ssoStatus, setSsoStatus] = useState<string | null>(null);

  // RBAC roles state
  const [editingRoles, setEditingRoles] = useState<RBACRole[]>([]);
  const [rolesSaving, setRolesSaving] = useState(false);
  const [rolesStatus, setRolesStatus] = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [licenseRes, auditRes, rolesRes] = await Promise.all([
        fetch("/api/go/api/enterprise/license").catch(() => null),
        fetch("/api/go/api/enterprise/audit?limit=20").catch(() => null),
        fetch("/api/go/api/enterprise/roles").catch(() => null),
      ]);
      if (licenseRes?.ok) {
        const d = await licenseRes.json();
        const licData = d.data ?? d;
        setLicense(licData);
        if (licData.ssoSettings) {
          setProviderUrl(licData.ssoSettings.providerUrl || "");
          setClientId(licData.ssoSettings.clientId || "");
          setClientSecret(licData.ssoSettings.clientSecret || "");
        }
      }
      if (auditRes?.ok) {
        const d = await auditRes.json();
        setAuditLogs(d.data ?? []);
      }
      if (rolesRes?.ok) {
        const d = await rolesRes.json();
        const rList = d.data ?? [];
        setRoles(rList);
        setEditingRoles(JSON.parse(JSON.stringify(rList)));
      }
    } catch {
      // Best-effort
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const saveSSO = async () => {
    setSsoSaving(true);
    setSsoStatus(null);
    try {
      const res = await fetch("/api/go/api/enterprise/sso/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ providerUrl, clientId, clientSecret }),
      });
      if (res.ok) {
        setSsoStatus("SSO configuration saved successfully!");
      } else {
        setSsoStatus("Failed to save SSO configuration.");
      }
    } catch (e: any) {
      setSsoStatus(`Error: ${e.message}`);
    }
    setSsoSaving(false);
  };

  const handleRoleDescChange = (index: number, val: string) => {
    const updated = [...editingRoles];
    updated[index].description = val;
    setEditingRoles(updated);
  };

  const handleRolePermsChange = (index: number, val: string) => {
    const updated = [...editingRoles];
    updated[index].permissions = val.split(",").map(p => p.trim()).filter(Boolean);
    setEditingRoles(updated);
  };

  const saveRoles = async () => {
    setRolesSaving(true);
    setRolesStatus(null);
    try {
      const res = await fetch("/api/go/api/enterprise/roles/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(editingRoles),
      });
      if (res.ok) {
        setRolesStatus("RBAC roles saved successfully!");
      } else {
        setRolesStatus("Failed to save RBAC roles.");
      }
    } catch (e: any) {
      setRolesStatus(`Error: ${e.message}`);
    }
    setRolesSaving(false);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="w-5 h-5 text-amber-400" />
          <div>
            <h1 className="text-lg font-semibold text-white">Enterprise Security</h1>
            <p className="text-xs text-zinc-500 mt-0.5">
              License validation, RBAC permissions, and audit logging
            </p>
          </div>
        </div>
        <button
          onClick={fetchAll}
          disabled={loading}
          className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50 flex items-center gap-1.5"
          title="Refresh enterprise security data"
        >
          {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
          Refresh
        </button>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* License Card */}
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
          <div className="flex items-center gap-2 mb-3">
            <Key className="w-4 h-4 text-amber-400" />
            <h2 className="text-sm font-medium text-white">License Status</h2>
          </div>
          {license ? (
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2">
                {license.valid ? (
                  <CheckCircle className="w-4 h-4 text-emerald-400" />
                ) : (
                  <XCircle className="w-4 h-4 text-red-400" />
                )}
                <span className={license.valid ? "text-emerald-400" : "text-red-400"}>
                  {license.valid ? "Valid License" : "No License Found"}
                </span>
              </div>
              {license.licensedTo && (
                <div className="text-zinc-400">
                  Licensed to: <span className="text-zinc-300">{license.licensedTo}</span>
                </div>
              )}
              {license.tier && (
                <div className="text-zinc-400">
                  Tier: <span className="text-zinc-300">{license.tier}</span>
                </div>
              )}
              {license.expiresAt && (
                <div className="text-zinc-400">
                  Expires: <span className="text-zinc-300">{license.expiresAt}</span>
                </div>
              )}
              {license.maxNodes && (
                <div className="text-zinc-400">
                  Max nodes: <span className="text-zinc-300">{license.maxNodes}</span>
                </div>
              )}
              {license.features && license.features.length > 0 && (
                <div>
                  <span className="text-zinc-500 text-xs font-semibold">Enabled Features</span>
                  <div className="flex gap-1 mt-1 flex-wrap">
                    {license.features.map((f) => (
                      <span key={f} className="px-1.5 py-0.5 bg-zinc-800 rounded text-2xs text-zinc-400">{f}</span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="text-zinc-600 text-sm italic">
              {loading ? "Checking license..." : "No license information available"}
            </div>
          )}
        </div>

        {/* SSO Settings Card */}
        <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4 space-y-3">
          <div className="flex items-center gap-2">
            <Shield className="w-4 h-4 text-emerald-400" />
            <h2 className="text-sm font-medium text-white">SSO Authentication Settings</h2>
          </div>

          <div className="space-y-2">
            <div>
              <label className="text-2xs text-zinc-500 block mb-1">Provider IDP URL</label>
              <input
                value={providerUrl}
                onChange={e => setProviderUrl(e.target.value)}
                placeholder="e.g. https://id.nexus.auth/oauth2"
                className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white focus:border-amber-500 outline-none"
              />
            </div>
            <div>
              <label className="text-2xs text-zinc-500 block mb-1">Client ID</label>
              <input
                value={clientId}
                onChange={e => setClientId(e.target.value)}
                placeholder="Client ID token"
                className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white focus:border-amber-500 outline-none"
              />
            </div>
            <div>
              <label className="text-2xs text-zinc-500 block mb-1">Client Secret</label>
              <input
                type="password"
                value={clientSecret}
                onChange={e => setClientSecret(e.target.value)}
                placeholder="••••••••••••••••"
                className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white focus:border-amber-500 outline-none"
              />
            </div>
          </div>

          <div className="flex items-center justify-between pt-2">
            <span className="text-2xs text-amber-500">{ssoStatus}</span>
            <button
              onClick={saveSSO}
              disabled={ssoSaving}
              className="px-3 py-1.5 bg-amber-500 text-black font-semibold rounded hover:bg-amber-400 text-xs flex items-center gap-1.5 disabled:opacity-50"
            >
              {ssoSaving ? <Loader2 className="w-3 h-3 animate-spin" /> : <Save className="w-3 h-3" />}
              Save SSO
            </button>
          </div>
        </div>
      </div>

      {/* RBAC Roles Editor */}
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <Users className="w-4 h-4 text-blue-400" />
            <h2 className="text-sm font-medium text-white">RBAC Roles & Permissions Configurator</h2>
          </div>
          <button
            onClick={saveRoles}
            disabled={rolesSaving}
            className="px-3 py-1 bg-blue-500 text-black font-semibold rounded hover:bg-blue-400 text-xs flex items-center gap-1"
          >
            {rolesSaving ? <Loader2 className="w-3 h-3 animate-spin" /> : <Save className="w-3 h-3" />}
            Save Roles
          </button>
        </div>

        {rolesStatus && (
          <div className="text-2xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 p-2 rounded mb-3">
            {rolesStatus}
          </div>
        )}

        {editingRoles.length === 0 ? (
          <div className="text-zinc-600 text-sm italic">
            {loading ? "Loading roles..." : "No roles configured"}
          </div>
        ) : (
          <div className="space-y-3">
            {editingRoles.map((role, idx) => (
              <div key={role.name} className="bg-zinc-950/50 rounded p-3 border border-zinc-800 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold text-zinc-300 uppercase tracking-wider">{role.name}</span>
                </div>
                <div className="grid gap-2 md:grid-cols-2">
                  <div>
                    <label className="text-2xs text-zinc-600 block mb-1">Description</label>
                    <input
                      value={role.description || ""}
                      onChange={e => handleRoleDescChange(idx, e.target.value)}
                      className="w-full bg-zinc-900 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 focus:border-blue-500 outline-none"
                    />
                  </div>
                  <div>
                    <label className="text-2xs text-zinc-600 block mb-1">Permissions (comma-separated)</label>
                    <input
                      value={role.permissions.join(", ")}
                      onChange={e => handleRolePermsChange(idx, e.target.value)}
                      className="w-full bg-zinc-900 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-300 focus:border-blue-500 outline-none"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Audit Logs */}
      <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-3">
          <FileText className="w-4 h-4 text-purple-400" />
          <h2 className="text-sm font-medium text-white">System Security Audit Log (Last 20 Actions)</h2>
        </div>
        {auditLogs.length === 0 ? (
          <div className="text-zinc-600 text-sm italic">
            {loading ? "Loading audit logs..." : "No audit entries recorded"}
          </div>
        ) : (
          <div className="space-y-1 max-h-60 overflow-y-auto pr-2">
            {auditLogs.map((log: any, i: number) => (
              <div key={i} className="text-xs flex gap-2 py-1 border-b border-zinc-800/50 last:border-0 font-mono text-zinc-400">
                <span className="text-zinc-600 shrink-0 w-20">
                  {log.timestamp?.slice(11, 19) || log.timestamp?.slice(0, 10) || "?"}
                </span>
                <span className="text-purple-400 shrink-0 w-24 uppercase tracking-wider text-2xs">{log.action?.slice(0, 20) || "?"}</span>
                <span className="text-zinc-300 truncate">{log.detail || JSON.stringify(log)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
