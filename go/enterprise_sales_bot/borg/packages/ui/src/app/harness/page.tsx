"use client";

import { ArchitectDashboard } from "@/components/harness/ArchitectDashboard";

export default function HarnessPage() {
  return (
    <div className="max-w-[1600px] mx-auto p-6">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-50">Coding Harness</h1>
          <p className="text-slate-400 mt-2">
            High-fidelity AI architect, visual repo maps, and autonomous verification.
          </p>
        </div>
      </div>
      <ArchitectDashboard />
    </div>
  );
}
