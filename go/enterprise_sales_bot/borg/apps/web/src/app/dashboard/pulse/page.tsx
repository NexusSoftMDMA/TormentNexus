
import { SystemPulse } from '@/components/pulse/SystemPulse';

export default function PulsePage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex flex-col gap-2">
                <h1 className="text-3xl font-bold tracking-tight">The Pulse</h1>
                <p className="text-muted-foreground">
                    Real-time observability of the Collective.
                </p>
            </div>

            <SystemPulse />
        </div>
    );
}
