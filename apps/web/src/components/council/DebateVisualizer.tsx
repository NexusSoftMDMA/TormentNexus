
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Avatar, AvatarFallback, AvatarImage } from "@tormentnexus/ui";
import { Check, X, Shield, Users, Brain, Activity, Gavel } from "lucide-react";

interface TranscriptEntry {
    speaker: string;
    text: string;
    vote?: boolean;
    confidence?: number;
    round?: number;
}

interface DebateConfig {
    rounds: number;
    status: 'active' | 'concluded';
    result?: string;
    agreement?: number;
}

interface Props {
    topic: string;
    transcripts: TranscriptEntry[];
    config: DebateConfig;
}

export function DebateVisualizer({ topic, transcripts, config }: Props) {
    const getRoleIcon = (speaker: string) => {
        if (speaker === "Architect") return <Brain className="h-4 w-4 text-blue-400" />;
        if (speaker === "Security Expert" || speaker === "Critic") return <Shield className="h-4 w-4 text-red-400" />;
        if (speaker === "QA Lead" || speaker === "Product") return <Users className="h-4 w-4 text-green-400" />;
        return <Activity className="h-4 w-4 text-gray-400" />;
    };

    const getRoleColor = (speaker: string) => {
        if (speaker === "Architect") return "bg-blue-950/30 border-blue-500/20";
        if (speaker === "Security Expert" || speaker === "Critic") return "bg-red-950/30 border-red-500/20";
        if (speaker === "QA Lead" || speaker === "Product") return "bg-green-950/30 border-green-500/20";
        return "bg-slate-900/30 border-slate-500/20";
    };

    return (
        <Card className="h-[600px] flex flex-col border-indigo-500/20 bg-slate-950/50">
            <CardHeader className="bg-slate-900/50 border-b border-white/5 pb-3">
                <div className="flex justify-between items-start">
                    <div>
                        <Badge variant="outline" className="mb-2 text-indigo-400 border-indigo-500/30">
                            Debate Session
                        </Badge>
                        <CardTitle className="text-lg font-medium leading-tight text-slate-200">
                            {topic}
                        </CardTitle>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                        {config.status === 'active' ? (
                            <Badge className="bg-green-500/20 text-green-400 animate-pulse border-none">
                                LIVE
                            </Badge>
                        ) : (
                            <Badge variant="secondary">Concluded</Badge>
                        )}
                        {config.agreement !== undefined && (
                            <div className="flex flex-col items-end gap-1">
                                <span className="text-[10px] text-slate-500 uppercase font-bold tracking-tighter">Consensus</span>
                                <div className="w-24 h-1.5 bg-slate-800 rounded-full overflow-hidden border border-white/5">
                                    <div
                                        className={`h-full transition-all duration-1000 ${
                                            config.agreement > 0.8 ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]' :
                                            config.agreement > 0.5 ? 'bg-amber-500' : 'bg-red-500'
                                        }`}
                                        style={{ width: `${config.agreement * 100}%` }}
                                    />
                                </div>
                                <span className="text-[10px] font-mono text-slate-400">{(config.agreement * 100).toFixed(0)}%</span>
                            </div>
                        )}
                    </div>
                </div>
            </CardHeader>
            <CardContent className="flex-1 p-0 overflow-hidden relative">
                <ScrollArea className="h-full p-4">
                    <div className="space-y-6">
                        {transcripts.length === 0 && (
                            <div className="text-center text-muted-foreground py-12 italic">
                                Waiting for opening statements...
                            </div>
                        )}
                        {transcripts.map((entry, idx) => (
                            <div key={idx} className={`flex gap-3 ${entry.speaker === 'Meta-Architect' ? 'justify-center' : ''}`}>
                                {entry.speaker !== 'Meta-Architect' && (
                                    <Avatar className="h-8 w-8 border border-white/10">
                                        <AvatarFallback className="bg-slate-800 text-xs text-muted-foreground">
                                            {entry.speaker[0]}
                                        </AvatarFallback>
                                    </Avatar>
                                )}

                                <div className={`flex flex-col max-w-[85%] ${entry.speaker === 'Meta-Architect' ? 'items-center w-full' : ''}`}>
                                    <div className="flex items-center gap-2 mb-1">
                                        <span className="text-xs font-semibold text-slate-400">
                                            {entry.speaker}
                                        </span>
                                        {getRoleIcon(entry.speaker)}
                                        {entry.round && (
                                            <span className="text-[10px] text-slate-600 bg-slate-900 px-1 rounded">
                                                Round {entry.round}
                                            </span>
                                        )}
                                    </div>

                                    <div className={`p-3 rounded-lg text-sm text-slate-300 border ${getRoleColor(entry.speaker)}`}>
                                        {entry.text}
                                    </div>

                                    {/* Vote Indicator */}
                                    {entry.vote !== undefined && (
                                        <div className="flex items-center gap-2 mt-1 ml-1">
                                            {entry.vote ? (
                                                <Badge variant="outline" className="text-[10px] h-4 bg-green-500/10 text-green-400 border-green-500/20 gap-1 px-1">
                                                    <Check className="h-2 w-2" /> Approve
                                                </Badge>
                                            ) : (
                                                <Badge variant="outline" className="text-[10px] h-4 bg-red-500/10 text-red-400 border-red-500/20 gap-1 px-1">
                                                    <X className="h-2 w-2" /> Dissent
                                                </Badge>
                                            )}
                                            {entry.confidence && (
                                                <span className="text-[10px] text-slate-600">
                                                    Conf: {(entry.confidence * 100).toFixed(0)}%
                                                </span>
                                            )}
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                </ScrollArea>

                {/* Overlay Result if Done */}
                {config.result && (
                    <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-slate-950 via-slate-950/95 to-transparent p-6pt-12">
                        <div className="bg-indigo-950/40 border border-indigo-500/30 p-4 rounded-lg mx-4 mb-4">
                            <h4 className="font-semibold text-indigo-300 mb-1 flex items-center gap-2">
                                <Gavel className="h-4 w-4" /> Final Verdict
                            </h4>
                            <p className="text-sm text-indigo-100/80">{config.result}</p>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
