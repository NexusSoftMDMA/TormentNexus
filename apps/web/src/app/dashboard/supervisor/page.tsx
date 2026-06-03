'use client';

import { trpc } from "@/utils/trpc";
import { useState } from 'react';
import { Card } from '@tormentnexus/ui';
import { Button } from '@tormentnexus/ui';
import { Input } from '@tormentnexus/ui';
import { Textarea } from '@tormentnexus/ui';

export default function SupervisorPage() {
    const [goal, setGoal] = useState('');
    const [plan, setPlan] = useState<any[] | null>(null);
    const [executionLog, setExecutionLog] = useState<string>('');
    const [isExecuting, setIsExecuting] = useState(false);

    const decomposeMutation = trpc.supervisor.decompose.useMutation();
    const superviseMutation = trpc.supervisor.supervise.useMutation();

    const handleDecompose = async () => {
        if (!goal) return;
        try {
            const result = await decomposeMutation.mutateAsync({ goal });
            setPlan(result);
        } catch (e: any) {
            setExecutionLog(prev => prev + `\n[Error] Decomposition failed: ${e.message}`);
        }
    };

    const handleExecute = async () => {
        if (!goal) return;
        setIsExecuting(true);
        setExecutionLog(prev => prev + `\n[System] Starting supervision of goal: "${goal}"...`);
        try {
            const result = await superviseMutation.mutateAsync({ goal });
            setExecutionLog(prev => prev + `\n${result}`);
        } catch (e: any) {
            setExecutionLog(prev => prev + `\n[Error] Execution failed: ${e.message}`);
        } finally {
            setIsExecuting(false);
        }
    };

    return (
        <div className="p-6 space-y-6 h-full flex flex-col">
            <div className="flex flex-col gap-2">
                <h1 className="text-3xl font-bold tracking-tight">👮 Supervisor</h1>
                <p className="text-muted-foreground">
                    Hierarchical task delegation and sub-agent orchestration.
                </p>
            </div>

            <Card className="p-6 space-y-4">
                <h2 className="text-xl font-semibold">Mission Control</h2>
                <div className="flex gap-4">
                    <Input
                        placeholder="Enter a high-level goal (e.g., 'Research and summarize the latest React 19 features')"
                        value={goal}
                        onChange={(e) => setGoal(e.target.value)}
                        className="flex-1"
                    />
                    <Button onClick={handleDecompose} disabled={!goal || decomposeMutation.isPending}>
                        {decomposeMutation.isPending ? 'Planning...' : 'Plan'}
                    </Button>
                    <Button onClick={handleExecute} disabled={!goal || isExecuting} variant="default">
                        {isExecuting ? 'Supervising...' : 'Execute'}
                    </Button>
                </div>
            </Card>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 flex-1 min-h-0">
                {/* Plan View */}
                <Card className="p-6 flex flex-col gap-4 overflow-hidden">
                    <h3 className="font-semibold border-b pb-2">Proposed Plan</h3>
                    <div className="flex-1 overflow-y-auto space-y-4">
                        {!plan && <div className="text-muted-foreground italic">No plan generated yet.</div>}
                        {plan?.map((task: any) => (
                            <div key={task.id} className="border rounded p-3 bg-muted/20">
                                <div className="flex justify-between items-start mb-2">
                                    <span className="font-mono text-xs bg-primary/10 text-primary px-2 py-1 rounded uppercase">
                                        {task.assignedTo}
                                    </span>
                                    <span className="text-xs text-muted-foreground">{task.status}</span>
                                </div>
                                <p className="text-sm">{task.description}</p>
                            </div>
                        ))}
                    </div>
                </Card>

                {/* Execution Log */}
                <Card className="p-6 flex flex-col gap-4 overflow-hidden">
                    <h3 className="font-semibold border-b pb-2">Execution Log</h3>
                    <div className="flex-1 overflow-y-auto bg-black text-green-400 font-mono text-sm p-4 rounded-md whitespace-pre-wrap">
                        {executionLog || "// Ready for orders..."}
                    </div>
                </Card>
            </div>
        </div>
    );
}
