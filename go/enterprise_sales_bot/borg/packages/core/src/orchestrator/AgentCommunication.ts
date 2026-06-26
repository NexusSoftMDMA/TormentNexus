
export interface AgentMessage {
    id: string;
    from: string;
    to: string;
    type: 'TASK_REQUEST' | 'TASK_RESPONSE' | 'STATUS_UPDATE' | 'HANDOFF';
    payload: any;
    timestamp: number;
}

export interface AgentCapability {
    name: string;
    description: string;
    inputSchema: any;
}

export interface AgentInterface {
    id: string;
    getCapabilities(): Promise<AgentCapability[]>;
    sendMessage(msg: AgentMessage): Promise<void>;
    onMessage(handler: (msg: AgentMessage) => void): void;
}
