'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from "./ui/sheet";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "./ui/dialog";
import { ScrollArea } from "./ui/scroll-area";
import { Users, Plus, Trash2, Terminal, GitBranch, Activity, RefreshCw, MessageSquare } from 'lucide-react';
import { toast } from "sonner";
import { trpc } from '@/utils/trpc';
import { useRouter } from 'next/navigation';

export function SquadsPanel() {
    const router = useRouter();
    const [spawnOpen, setSpawnOpen] = useState(false);
    const [newBranch, setNewBranch] = useState('');
    const [newGoal, setNewGoal] = useState('');

    const utils = trpc.useUtils();
    const membersQuery = trpc.squad.list.useQuery(undefined, {
        refetchInterval: 5000 // Refresh every 5s
    });

    const spawnMutation = trpc.squad.spawn.useMutation({
        onSuccess: () => {
            toast.success("Squad member spawned successfully!");
            setSpawnOpen(false);
            setNewBranch('');
            setNewGoal('');
            utils.squad.list.invalidate();
        },
        onError: (err: any) => {
            toast.error(`Failed to spawn member: ${err.message}`);
        }
    });



    const [chatMember, setChatMember] = useState<any>(null);
    const [brainMember, setBrainMember] = useState<any>(null);
    const [chatMessage, setChatMessage] = useState('');

    const chatMutation = trpc.squad.chat.useMutation({
        onSuccess: () => {
            toast.success("Message sent to agent.");
            setChatMessage('');
        },
        onError: (err: any) => {
            toast.error(`Failed to send message: ${err.message}`);
        }
    });

    const handleSendChat = () => {
        if (!chatMember || !chatMessage.trim()) return;
        chatMutation.mutate({ branch: chatMember.branch, message: chatMessage });
    };

    const killMutation = trpc.squad.kill.useMutation({
        onSuccess: () => {
            toast.success("Squad member terminated.");
            utils.squad.list.invalidate();
        },
        onError: (err: any) => {
            toast.error(`Failed to kill member: ${err.message}`);
        }
    });

    const handleSpawn = () => {
        if (!newBranch || !newGoal) {
            toast.error("Branch and Goal are required");
            return;
        }
        spawnMutation.mutate({ branch: newBranch, goal: newGoal });
    };



    return (
        <div className="container mx-auto p-6 max-w-6xl space-y-8">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        <Users className="w-8 h-8" />
                        Squads
                    </h1>
                    <p className="text-muted-foreground mt-1">
                        Manage autonomous agents running in parallel Git Worktrees.
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="icon" onClick={() => membersQuery.refetch()}>
                        <RefreshCw className={`w-4 h-4 ${membersQuery.isFetching ? 'animate-spin' : ''}`} />
                    </Button>
                    <Dialog open={spawnOpen} onOpenChange={setSpawnOpen}>
                        <DialogTrigger asChild>
                            <Button>
                                <Plus className="w-4 h-4 mr-2" />
                                Spawn Member
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>Spawn New Squad Member</DialogTitle>
                                <DialogDescription>
                                    This will create a new Git Worktree and launch a Director agent to work on the specified task.
                                </DialogDescription>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                                <div className="grid gap-2">
                                    <label htmlFor="branch" className="text-sm font-medium">Branch Name</label>
                                    <Input
                                        id="branch"
                                        placeholder="feat/my-feature"
                                        value={newBranch}
                                        onChange={(e) => setNewBranch(e.target.value)}
                                    />
                                </div>
                                <div className="grid gap-2">
                                    <label htmlFor="goal" className="text-sm font-medium">Goal / Task</label>
                                    <Input
                                        id="goal"
                                        placeholder="Refactor the login component..."
                                        value={newGoal}
                                        onChange={(e) => setNewGoal(e.target.value)}
                                    />
                                </div>
                            </div>
                            <DialogFooter>
                                <Button onClick={handleSpawn} disabled={spawnMutation.isPending}>
                                    {spawnMutation.isPending ? 'Spawning...' : 'Spawn'}
                                </Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {membersQuery.isLoading ? (
                    <div className="col-span-full py-12 text-center text-muted-foreground">Loading members...</div>
                ) : membersQuery.data?.length === 0 ? (
                    <div className="col-span-full py-12 text-center border-2 border-dashed rounded-lg">
                        <Users className="w-12 h-12 mx-auto text-muted-foreground/50 mb-4" />
                        <h3 className="text-lg font-medium">No Active Squads</h3>
                        <p className="text-muted-foreground">Spawn a member to start parallel work.</p>
                    </div>
                ) : (
                    membersQuery.data?.map((member: any) => (
                        <Card key={member.id} className="relative overflow-hidden">
                            <CardHeader className="pb-3">
                                <div className="flex justify-between items-start">
                                    <CardTitle className="text-lg font-mono truncate mr-2" title={member.branch}>
                                        {member.branch}
                                    </CardTitle>
                                    <Badge variant={member.status === 'busy' ? "default" : "secondary"}>
                                        {member.status}
                                    </Badge>
                                </div>
                                <CardDescription className="font-mono text-xs text-muted-foreground truncate" title={member.id}>
                                    ID: {member.id}
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-4">
                                    <div className="flex items-center text-sm text-muted-foreground">
                                        <GitBranch className="w-4 h-4 mr-2" />
                                        Worktree Active
                                    </div>
                                    <div className="flex items-center text-sm text-muted-foreground">
                                        <Activity className="w-4 h-4 mr-2" />
                                        Director: {member.active ? 'Running' : 'Idle'}
                                    </div>
                                    {member.brain && (
                                        <div className="text-xs bg-muted/50 p-2 rounded border border-dashed hover:bg-muted cursor-help" onClick={() => setBrainMember(member)} title="Click to view thought process">
                                            <div className="font-semibold mb-1">🧠 Brain Activity:</div>
                                            <div>Step: {member.brain.step || 0} / {member.brain.totalSteps || '?'}</div>
                                            <div className="truncate opacity-70">Goal: {member.brain.goal || 'No active goal'}</div>
                                        </div>
                                    )}

                                    <div className="flex justify-end pt-2 gap-2">
                                        <Button
                                            variant="secondary"
                                            size="sm"
                                            onClick={() => setBrainMember(member)}
                                        >
                                            <Activity className="w-4 h-4 mr-2" />
                                            Brain
                                        </Button>
                                        <Button
                                            variant="secondary"
                                            size="sm"
                                            onClick={() => setChatMember(member)}
                                        >
                                            <MessageSquare className="w-4 h-4 mr-2" />
                                            Chat
                                        </Button>
                                        <Button
                                            variant="destructive"
                                            size="sm"
                                            onClick={() => {
                                                if (confirm(`Are you sure you want to kill ${member.branch}? This will delete the worktree.`)) {
                                                    killMutation.mutate({ branch: member.branch });
                                                }
                                            }}
                                            disabled={killMutation.isPending}
                                        >
                                            <Trash2 className="w-4 h-4 mr-2" />
                                            Kill
                                        </Button>
                                    </div>
                                </div>
                            </CardContent>
                            <div className="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
                                <Terminal className="w-24 h-24" />
                            </div>
                        </Card>
                    ))
                )}
            </div>

            <Sheet open={!!chatMember} onOpenChange={(open) => !open && setChatMember(null)}>
                <SheetContent side="right" className="min-w-[400px]">
                    <SheetHeader>
                        <SheetTitle>Chat with {chatMember?.branch}</SheetTitle>
                        <SheetDescription>
                            Send instructions or inject context into the running agent.
                        </SheetDescription>
                    </SheetHeader>
                    <div className="py-6 space-y-4 flex flex-col h-full max-h-[calc(100vh-120px)]">
                        <div className="flex-1 border rounded-md p-4 bg-muted/30 text-xs font-mono text-muted-foreground overflow-y-auto">
                            <p className="mb-2 opacity-70">Agent output is currently visible in the Director logs or server console.</p>
                            <div className="p-2 border border-dashed rounded bg-background/50">
                                History injection enabled. Instructions sent here will be added to the agent's active context loop.
                            </div>
                        </div>
                        <div className="space-y-2 mt-auto">
                            <label className="text-sm font-medium">Message</label>
                            <textarea
                                className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 resize-none"
                                value={chatMessage}
                                onChange={(e) => setChatMessage(e.target.value)}
                                placeholder="Enter command or context..."
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' && !e.shiftKey) {
                                        e.preventDefault();
                                        handleSendChat();
                                    }
                                }}
                            />
                            <div className="text-xs text-muted-foreground text-right">
                                Press Enter to send, Shift+Enter for new line
                            </div>
                        </div>
                    </div>
                    <SheetFooter>
                        <Button onClick={handleSendChat} disabled={chatMutation.isPending} className="w-full">
                            {chatMutation.isPending ? 'Sending...' : 'Send Instruction'}
                        </Button>
                    </SheetFooter>
                </SheetContent>
            </Sheet>

            <Sheet open={!!brainMember} onOpenChange={(open) => !open && setBrainMember(null)}>
                <SheetContent side="left" className="min-w-[500px] sm:min-w-[600px]">
                    <SheetHeader>
                        <SheetTitle>Brain Activity: {brainMember?.branch}</SheetTitle>
                        <SheetDescription>
                            Real-time thought process and execution history.
                        </SheetDescription>
                    </SheetHeader>
                    {brainMember?.brain ? (
                        <ScrollArea className="h-[calc(100vh-120px)] mt-4 pr-4">
                            <div className="space-y-6">
                                <div className="space-y-2">
                                    <h3 className="text-sm font-semibold flex items-center gap-2">
                                        <Activity className="w-4 h-4 text-blue-500" />
                                        Current Status
                                    </h3>
                                    <div className="grid grid-cols-2 gap-2 text-sm">
                                        <div className="bg-muted p-2 rounded">
                                            <span className="text-muted-foreground block text-xs">State</span>
                                            <span className="font-mono">{brainMember.brain.status}</span>
                                        </div>
                                        <div className="bg-muted p-2 rounded">
                                            <span className="text-muted-foreground block text-xs">Progress</span>
                                            <span className="font-mono">{brainMember.brain.step || 0} / {brainMember.brain.totalSteps || '?'}</span>
                                        </div>
                                        <div className="col-span-2 bg-muted p-2 rounded">
                                            <span className="text-muted-foreground block text-xs">Active Goal</span>
                                            <div className="font-medium">{brainMember.brain.goal || 'No active goal'}</div>
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-2">
                                    <h3 className="text-sm font-semibold flex items-center gap-2">
                                        <Terminal className="w-4 h-4 text-green-500" />
                                        Recent Thought Trace
                                    </h3>
                                    <div className="bg-black/90 text-green-400 p-4 rounded-md font-mono text-xs overflow-x-auto whitespace-pre-wrap">
                                        {brainMember.brain.lastHistory && brainMember.brain.lastHistory.length > 0
                                            ? brainMember.brain.lastHistory.join('\n\n')
                                            : <span className="text-muted-foreground opacity-50">// No recent history available</span>}
                                    </div>
                                </div>
                            </div>
                        </ScrollArea>
                    ) : (
                        <div className="py-12 text-center text-muted-foreground">
                            Brain data unavailable (Director might be sleeping).
                        </div>
                    )}
                </SheetContent>
            </Sheet>
        </div>
    );
}
