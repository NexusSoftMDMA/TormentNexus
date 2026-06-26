import { create } from 'zustand';

export interface DebugPacket {
  id: string;
  timestamp: number;
  type: 'request' | 'response' | 'error';
  direction: 'outbound' | 'inbound';
  method: string;
  payload: any;
  toolName?: string;
  durationMs?: number;
}

interface DebuggerState {
  packets: DebugPacket[];
  isRecording: boolean;
  addPacket: (packet: Omit<DebugPacket, 'id' | 'timestamp'>) => void;
  clearPackets: () => void;
  toggleRecording: () => void;
}

export const useDebuggerStore = create<DebuggerState>((set) => ({
  packets: [],
  isRecording: true,
  addPacket: (packet) =>
    set((state) => {
      if (!state.isRecording) return state;
      const newPacket: DebugPacket = {
        ...packet,
        id: Math.random().toString(36).substring(2, 9),
        timestamp: Date.now(),
      };
      // Keep last 100 packets
      return { packets: [newPacket, ...state.packets].slice(0, 100) };
    }),
  clearPackets: () => set({ packets: [] }),
  toggleRecording: () => set((state) => ({ isRecording: !state.isRecording })),
}));
