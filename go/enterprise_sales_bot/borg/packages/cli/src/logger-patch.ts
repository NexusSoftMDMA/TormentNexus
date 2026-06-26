import { EventEmitter } from 'events';

export const LogEmitter = new EventEmitter();

// Patches the global console object to prepend timestamps
const originalLog = console.log;
const originalWarn = console.warn;
const originalError = console.error;
const originalInfo = console.info;

function getTimestamp() {
    return new Date().toISOString();
}

function emitLog(level: string, args: any[]) {
    const message = args.map(a =>
        typeof a === 'object' ? JSON.stringify(a) : String(a)
    ).join(' ');

    LogEmitter.emit('log', {
        id: Math.random().toString(36).substr(2, 9),
        timestamp: getTimestamp(),
        level,
        message
    });
}

console.log = function (...args: any[]) {
    emitLog('info', args);
    // process.stdout is hooked by Ink, so we might want to avoid direct stdout writes
    // depending on how Ink is configured. 
    // Ideally, for TUI, we STOP logging to stdout directly to avoid breaking the UI,
    // and ONLY send to the LogEmitter.
    // BUT, if Ink crashes, we lose logs.
    // For now, let's keep originalLog but maybe Ink handles it.
    // Ink usually intercepts stdout.
    // originalLog.apply(console, [`[${getTimestamp()}]`, ...args]);
};

console.warn = function (...args: any[]) {
    emitLog('warn', args);
    // originalWarn.apply(console, [`[${getTimestamp()}]`, ...args]);
};

console.error = function (...args: any[]) {
    emitLog('error', args);
    // originalError.apply(console, [`[${getTimestamp()}]`, ...args]);
};

console.info = function (...args: any[]) {
    emitLog('info', args);
    // originalInfo.apply(console, [`[${getTimestamp()}]`, ...args]);
};

