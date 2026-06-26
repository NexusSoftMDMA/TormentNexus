'use client';

import React, { useState, useEffect } from 'react';

// Minimal types for SpeechRecognition to avoid 'any'
interface SpeechRecognition extends EventTarget {
    continuous: boolean;
    interimResults: boolean;
    lang: string;
    start(): void;
    stop(): void;
    onstart: ((this: SpeechRecognition, ev: Event) => any) | null;
    onend: ((this: SpeechRecognition, ev: Event) => any) | null;
    onresult: ((this: SpeechRecognition, ev: SpeechRecognitionEvent) => any) | null;
    onerror: ((this: SpeechRecognition, ev: SpeechRecognitionErrorEvent) => any) | null;
}

interface SpeechRecognitionEvent extends Event {
    results: SpeechRecognitionResultList;
}

interface SpeechRecognitionResultList {
    [index: number]: SpeechRecognitionResult;
}

interface SpeechRecognitionResult {
    [index: number]: SpeechRecognitionAlternative;
}

interface SpeechRecognitionAlternative {
    transcript: string;
}

interface SpeechRecognitionErrorEvent extends Event {
    error: string;
}

interface VoiceInputProps {
    onTranscript: (text: string) => void;
    isProcessing?: boolean;
}

export default function VoiceInput({ onTranscript, isProcessing }: VoiceInputProps) {
    const [isListening, setIsListening] = useState(false);
    const recognitionRef = React.useRef<SpeechRecognition | null>(null);

    useEffect(() => {
        // Init Speech Recognition
        if (typeof window !== 'undefined') {
            // @ts-expect-error - SpeechRecognition is experimental
            const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
            if (SpeechRecognition) {
                const reco = new SpeechRecognition();
                reco.continuous = false;
                reco.interimResults = false;
                reco.lang = 'en-US';

                reco.onstart = () => setIsListening(true);
                reco.onend = () => setIsListening(false);
                reco.onresult = (event: SpeechRecognitionEvent) => {
                    const text = event.results[0][0].transcript;
                    console.log('Voice Transcript:', text);
                    onTranscript(text);
                };
                reco.onerror = (event: SpeechRecognitionErrorEvent) => {
                    console.error('Speech Error:', event.error);
                    setIsListening(false);
                };

                recognitionRef.current = reco;
            }
        }
    }, [onTranscript]);

    const toggleListening = () => {
        if (!recognitionRef.current) {
            alert("Speech Recognition not supported in this browser.");
            return;
        }

        if (isListening) {
            recognitionRef.current.stop();
        } else {
            recognitionRef.current.start();
        }
    };

    return (
        <button
            onClick={toggleListening}
            disabled={isProcessing}
            className={`p-2 rounded-full transition-all duration-300 flex items-center justify-center ${isListening
                ? 'bg-red-500/20 text-red-500 ring-2 ring-red-500 animate-pulse'
                : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-white'
                }`}
            title="Toggle Voice Command"
        >
            {isListening ? (
                <span className="w-5 h-5">⏹</span>
            ) : (
                <span className="w-5 h-5">🎤</span>
            )}
        </button>
    );
}
