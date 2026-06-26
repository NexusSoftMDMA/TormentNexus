import { CouncilDebateWidget } from "./CouncilDebateWidget";

export function CouncilDashboard() {
  return (
    <div className="flex-1 overflow-y-auto p-8 h-full">
      <div className="mx-auto max-w-6xl space-y-8 h-full flex flex-col">
        <div className="flex items-center justify-between flex-shrink-0">
          <div>
            <h2 className="text-lg font-bold text-white uppercase tracking-widest">Autopilot Council</h2>
            <p className="text-sm text-white/60">Live debate stream from the Multi-Agent Consensus Engine.</p>
          </div>
        </div>

        <div className="flex-1 min-h-0">
          <CouncilDebateWidget />
        </div>
      </div>
    </div>
  );
}
