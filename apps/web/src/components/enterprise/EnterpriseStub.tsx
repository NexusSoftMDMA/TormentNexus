import React from 'react';

export const EnterpriseStub = () => (
  <div className="p-8 rounded-xl bg-slate-950/40 border border-dashed border-slate-800 text-center">
    <h3 className="text-xl font-bold text-white mb-2">Enterprise Module Required</h3>
    <p className="text-slate-400 text-sm mb-4">
      This feature (SSO/RBAC) requires a commercial Enterprise license.
    </p>
    <button className="px-4 py-2 bg-purple-600 text-white rounded text-sm font-medium">
      Upgrade to Enterprise
    </button>
  </div>
);
