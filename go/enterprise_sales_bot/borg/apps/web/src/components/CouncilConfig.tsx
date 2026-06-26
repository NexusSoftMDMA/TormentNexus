'use client';
import { useState, useEffect } from 'react';
import { trpc } from '@/utils/trpc';

export default function CouncilConfig() {
    const configQuery = trpc.directorConfig.get.useQuery(undefined, { refetchInterval: 10000 });
    const updateMutation = trpc.directorConfig.update.useMutation({
        onSuccess: () => configQuery.refetch()
    });

    const [council, setCouncil] = useState<any>({ personas: [], contextFiles: [] });
    const [isExpanded, setIsExpanded] = useState(false);

    useEffect(() => {
        const data = configQuery.data as any;
        if (data?.council) {
            setCouncil(data.council);
        }
    }, [configQuery.data]);

    const handleUpdate = (newCouncil: any) => {
        setCouncil(newCouncil);
        updateMutation.mutate({ council: newCouncil } as any);
    };

    const addPersona = () => {
        handleUpdate({
            ...council,
            personas: [...(council.personas || []), 'New Persona']
        });
    };

    const removePersona = (idx: number) => {
        const newPersonas = [...(council.personas || [])];
        newPersonas.splice(idx, 1);
        handleUpdate({ ...council, personas: newPersonas });
    };

    const updatePersona = (idx: number, val: string) => {
        const newPersonas = [...(council.personas || [])];
        newPersonas[idx] = val;
        setCouncil({ ...council, personas: newPersonas }); // Optimistic
    };

    const savePersona = () => {
        updateMutation.mutate({ council } as any);
    };

    if (configQuery.isPending) return null;

    return (
        <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="w-full flex items-center justify-between p-4 bg-gray-900 hover:bg-gray-800 transition-colors"
            >
                <div className="flex items-center gap-2">
                    <span className="text-xl">🏛️</span>
                    <h2 className="text-lg font-bold text-gray-200">Council Configuration</h2>
                </div>
                <span className="text-gray-500">{isExpanded ? '▼' : '▶'}</span>
            </button>

            {isExpanded && (
                <div className="p-4 border-t border-gray-800 space-y-6 animate-in slide-in-from-top-2">

                    {/* Personas */}
                    <div className="space-y-3">
                        <div className="flex justify-between items-center">
                            <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wider">Personas</h3>
                            <button onClick={addPersona} className="text-xs bg-blue-900/30 text-blue-400 px-2 py-1 rounded border border-blue-900 hover:bg-blue-900/50">
                                + Add System Input
                            </button>
                        </div>
                        <div className="space-y-2">
                            {(council.personas || []).map((p: string, i: number) => (
                                <div key={i} className="flex gap-2">
                                    <input
                                        className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-1 text-sm text-gray-200 focus:border-blue-500 outline-none"
                                        value={p}
                                        onChange={(e) => updatePersona(i, e.target.value)}
                                        onBlur={savePersona}
                                    />
                                    <button onClick={() => removePersona(i)} className="text-red-500 hover:text-red-400 px-2">×</button>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Context Context */}
                    <div className="space-y-3">
                        <h3 className="text-sm font-medium text-gray-400 uppercase tracking-wider">Context Sources</h3>
                        <div className="text-xs text-gray-500 mb-2">Files the Council reads for decision making.</div>
                        <div className="space-y-2">
                            {(council.contextFiles || []).map((f: string, i: number) => (
                                <div key={i} className="flex gap-2 items-center bg-gray-800/50 px-3 py-2 rounded text-sm text-gray-300 font-mono">
                                    <span className="flex-1 truncate">{f}</span>
                                </div>
                            ))}
                            {/* Future: File Picker */}
                        </div>
                    </div>

                </div>
            )}
        </div>
    );
}
