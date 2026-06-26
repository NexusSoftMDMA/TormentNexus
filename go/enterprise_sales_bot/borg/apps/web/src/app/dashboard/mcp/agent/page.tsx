"use client";

import { useState, useRef, useEffect } from 'react';
import { Card, CardContent } from "@tormentnexus/ui";
import { Button } from "@tormentnexus/ui";
import { Loader2, Send, Bot, User, Terminal } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

export default function AgentPlayground() {
    const [messages, setMessages] = useState<Array<{ role: 'user' | 'assistant', content: string, tools?: any[] }>>([
        { role: 'assistant', content: "Hello! I'm your TormentNexus Agent. I have access to all your connected tools. How can I help you today?" }
    ]);
    const [input, setInput] = useState('');
    const scrollRef = useRef<HTMLDivElement>(null);

    const chatMutation = trpc.agent.chat.useMutation({
        onSuccess: (data) => {
            setMessages(prev => [...prev, {
                role: 'assistant',
                content: data.response,
                tools: data.tool_calls
            }]);
        },
        onError: (err) => {
            toast.error(`Agent error: ${err.message}`);
            setMessages(prev => [...prev, { role: 'assistant', content: `Error: ${err.message}` }]);
        }
    });

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages]);

    const handleSend = (e: React.FormEvent) => {
        e.preventDefault();
        if (!input.trim() || chatMutation.isPending) return;

        const userMsg = input;
        setMessages(prev => [...prev, { role: 'user', content: userMsg }]);
        setInput('');

        chatMutation.mutate({ message: userMsg });
    };

    return (
        <div className="p-8 space-y-8 h-full flex flex-col">
            <div className="flex justify-between items-center shrink-0">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight text-white">Agent Playground</h1>
                    <p className="text-zinc-500">
                        Chat with an agent capable of using your tools
                    </p>
                </div>
            </div>

            <Card className="bg-zinc-900 border-zinc-800 flex-1 flex flex-col overflow-hidden">
                {/* Chat Area */}
                <CardContent className="flex-1 overflow-y-auto p-4 space-y-4" ref={scrollRef}>
                    {messages.map((msg, idx) => (
                        <div key={idx} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                            <div className={`max-w-[80%] rounded-lg p-3 ${msg.role === 'user'
                                    ? 'bg-blue-600 text-white'
                                    : 'bg-zinc-800 text-zinc-100'
                                }`}>
                                <div className="flex items-center gap-2 mb-1 opacity-50 text-xs uppercase font-bold tracking-wider">
                                    {msg.role === 'user' ? <User className="h-3 w-3" /> : <Bot className="h-3 w-3" />}
                                    {msg.role}
                                </div>
                                <div className="whitespace-pre-wrap">{msg.content}</div>

                                {/* Tool Call Visualization */}
                                {msg.tools && msg.tools.length > 0 && (
                                    <div className="mt-3 space-y-2">
                                        {msg.tools.map((toolCall: any, i: number) => (
                                            <div key={i} className="bg-black/30 rounded p-2 text-xs font-mono border border-white/10">
                                                <div className="flex items-center gap-1 text-purple-400 mb-1">
                                                    <Terminal className="h-3 w-3" />
                                                    {toolCall.name}
                                                </div>
                                                <div className="text-zinc-400 truncate">
                                                    {JSON.stringify(toolCall.args)}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                    {chatMutation.isPending && (
                        <div className="flex justify-start">
                            <div className="bg-zinc-800 rounded-lg p-4 flex items-center gap-2">
                                <Loader2 className="h-4 w-4 animate-spin text-zinc-400" />
                                <span className="text-zinc-400 text-sm">Thinking...</span>
                            </div>
                        </div>
                    )}
                </CardContent>

                {/* Input Area */}
                <div className="p-4 border-t border-zinc-800 bg-zinc-950">
                    <form onSubmit={handleSend} className="flex gap-2">
                        <input
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            className="flex-1 bg-zinc-900 border border-zinc-800 rounded-md p-3 text-white focus:ring-1 focus:ring-blue-500 outline-none"
                            placeholder="Ask the agent to do something..."
                            disabled={chatMutation.isPending}
                        />
                        <Button
                            type="submit"
                            disabled={!input.trim() || chatMutation.isPending}
                            className="bg-blue-600 hover:bg-blue-500 min-w-[100px]"
                        >
                            <Send className="h-4 w-4" />
                        </Button>
                    </form>
                </div>
            </Card>
        </div>
    );
}
