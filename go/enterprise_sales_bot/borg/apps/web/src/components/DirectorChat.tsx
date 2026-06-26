"use client";

import { useState, useEffect } from "react";
import { trpc } from "../utils/trpc";

function getPasteToSubmitDelayMs(value: unknown): number {
    if (typeof value !== 'object' || value === null) {
        return 0;
    }

    const delay = (value as { pasteToSubmitDelayMs?: unknown }).pasteToSubmitDelayMs;
    return typeof delay === 'number' ? delay : 0;
}

export function DirectorChat() {
    const [input, setInput] = useState("");
    const [messages, setMessages] = useState<{ role: 'user' | 'agent', content: string }[]>([
        { role: 'agent', content: 'Hello! I am the Director. What task shall I perform?' }
    ]);
    const [submitTimer, setSubmitTimer] = useState<NodeJS.Timeout | null>(null);

    // Fetch config for auto-submit delay
    const configQuery = trpc.directorConfig.get.useQuery();
    const pasteToSubmitDelayMs = getPasteToSubmitDelayMs(configQuery.data);

    const chatMutation = trpc.director.chat.useMutation({
        onSuccess: (data) => {
            const content = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
            setMessages(prev => [...prev, { role: 'agent', content }]);
        },
        onError: (error) => {
            setMessages(prev => [...prev, { role: 'agent', content: `Error: ${error.message}` }]);
        }
    });

    const submitMessage = (msg: string) => {
        if (!msg.trim() || chatMutation.isPending) return;
        setMessages(prev => [...prev, { role: 'user', content: msg }]);
        setInput("");
        chatMutation.mutate({ message: msg });
        if (submitTimer) clearTimeout(submitTimer);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        submitMessage(input);
    };


    // Allow external injection (e.g. from Extension)
    useEffect(() => {
        // @ts-ignore
        window.injectDirectorMessage = (text: string, autoSubmit: boolean = false) => {
            console.log("[DirectorChat] Received external injection:", text);
            setInput(text);
            if (autoSubmit) {
                // Short delay to allow state update
                setTimeout(() => submitMessage(text), 100);
            }
        };
        return () => {
            // @ts-ignore
            delete window.injectDirectorMessage;
        }
    }, [chatMutation.isPending]); // Re-bind if mutation state changes? No need really.

    const handlePaste = (e: React.ClipboardEvent) => {
        const text = e.clipboardData.getData('text');
        // If config allows and delay > 0
        const delay = pasteToSubmitDelayMs;

        if (delay > 0 && text.trim().length > 0) {
            // We need to wait for state update or pass text directly. 
            // Setting input here might race with the timer if we use 'input' state in timeout.
            // Better to just set a timer that submits the COMBINED text (current input + paste).
            // But React handlePaste default behavior inserts text *after*. 
            // Let's rely on event default insertion? No, controlled input.

            // Actually, usually users just want the pasted content + auto submit.
            // Let's prevent default, insert text, sets input, then schedule submit.
            e.preventDefault();
            const newVal = input + text; // Simplified (appends to end)
            setInput(newVal);

            if (submitTimer) clearTimeout(submitTimer);
            const timer = setTimeout(() => {
                submitMessage(newVal);
            }, delay);
            setSubmitTimer(timer);
        }
    };

    // Clear timer if user types manually to cancel auto-submit
    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setInput(e.target.value);
        if (submitTimer) {
            clearTimeout(submitTimer);
            setSubmitTimer(null);
        }
    };

    const EMERGENCY_MODE = false; // Emergency over

    return (
        <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 shadow-xl w-full">
            <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
                💬 Director Chat
            </h2>

            {EMERGENCY_MODE && (
                <div className="bg-red-900/50 text-red-200 text-xs p-2 rounded border border-red-500/50 animate-pulse mb-4">
                    🛑 <b>EMERGENCY STOP ACTIVE</b>: Auto-typing disabled to break loop. Please restart Director process.
                </div>
            )}

            <div className="bg-zinc-950 border border-zinc-800 rounded-md p-4 h-64 overflow-y-auto mb-4 flex flex-col gap-3">
                {messages.map((msg, i) => (
                    <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                        <div className={`max-w-[80%] rounded-lg p-3 text-sm ${msg.role === 'user'
                            ? 'bg-blue-600 text-white'
                            : 'bg-zinc-800 text-zinc-200'
                            }`}>
                            {msg.content}
                        </div>
                    </div>
                ))}
                {chatMutation.isPending && (
                    <div className="flex justify-start">
                        <div className="bg-zinc-800 text-zinc-400 rounded-lg p-3 text-sm animate-pulse">
                            Thinking...
                        </div>
                    </div>
                )}
            </div>

            <form onSubmit={handleSubmit} className="flex gap-2">
                <input
                    type="text"
                    value={input}
                    onChange={handleChange}
                    onPaste={handlePaste}
                    placeholder="Tell the Director what to do..."
                    className="flex-1 bg-zinc-800 border border-zinc-700 rounded p-2 text-white placeholder-zinc-500 focus:outline-none focus:border-blue-500"
                />
                <button
                    type="submit"
                    disabled={chatMutation.isPending}
                    className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded disabled:opacity-50"
                >
                    Send
                </button>
            </form>
            {submitTimer && <div className="text-xs text-blue-400 mt-1 animate-pulse">Auto-submitting in {(pasteToSubmitDelayMs || 1000) / 1000}s... (type to cancel)</div>}
        </div >
    );
}
