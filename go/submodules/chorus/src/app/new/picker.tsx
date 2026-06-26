"use client";

import { useState } from "react";

interface PickerProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  wide?: boolean;
  children: React.ReactNode;
}

export function Picker({ icon, label, value, wide, children }: PickerProps) {
  const [open, setOpen] = useState(false);
  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-2 rounded-md border border-border bg-card px-3 py-1.5 text-sm transition hover:border-muted-foreground/40"
      >
        <span className="text-muted-foreground">{icon}</span>
        <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
          {label}
        </span>
        <span className="font-medium">{value}</span>
      </button>
      {open && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setOpen(false)}
          />
          <div
            className={`absolute left-0 top-full z-20 mt-1 rounded-md border border-border bg-popover p-1.5 shadow-xl ${wide ? "w-80" : "w-56"}`}
            onClick={() => setOpen(false)}
          >
            {children}
          </div>
        </>
      )}
    </div>
  );
}
