import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Shield, Lock } from "lucide-react";
import { trpc } from '../utils/trpc'; // Context-provided tRPC

export const SecurityWidget = () => {
    // @ts-ignore
    const rulesQuery = trpc.policy.getRules.useQuery(undefined, { refetchInterval: 5000 });
    // @ts-ignore
    const autonomyQuery = trpc.autonomy.getLevel.useQuery(undefined, { refetchInterval: 5000 });

    const isLocked = rulesQuery.data?.[0]?.reason === 'SYSTEM LOCKDOWN';
    const autonomy = autonomyQuery.data || 'unknown';

    return (
        <Card className={isLocked ? "bg-red-50 dark:bg-red-950/30 border-red-500" : ""}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Security Status</CardTitle>
                <Shield className={`h-4 w-4 ${isLocked ? "text-red-500" : "text-green-500"}`} />
            </CardHeader>
            <CardContent>
                <div className={`text-2xl font-bold ${isLocked ? "text-red-500" : "text-green-500"}`}>
                    {isLocked ? "LOCKDOWN" : "SECURE"}
                </div>
                <div className="flex items-center gap-2 mt-2">
                    <span className="text-xs text-muted-foreground uppercase">{autonomy} Autonomy</span>
                    {isLocked && <Lock className="h-3 w-3 text-red-500" />}
                </div>
            </CardContent>
        </Card>
    );
};
