'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function OpencodeAutopilotPage() {
  return (
    <div className="p-6 space-y-6">
      <h1 className="text-3xl font-bold">Opencode Autopilot Dashboard</h1>
      <p className="text-muted-foreground">Autonomous coding agent management.</p>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Agent Fleet</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2">
              <li className="flex justify-between"><span>Architect</span> <span className="text-green-500">Idle</span></li>
              <li className="flex justify-between"><span>Implementer</span> <span className="text-green-500">Idle</span></li>
              <li className="flex justify-between"><span>Reviewer</span> <span className="text-gray-400">Offline</span></li>
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Task Queue</CardTitle>
          </CardHeader>
          <CardContent>
            <p>No active tasks.</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
