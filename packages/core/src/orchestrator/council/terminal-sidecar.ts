import * as pty from 'node-pty';
import { createServer } from 'net';

/**
 * Terminal Sidecar - Integrated into TormentNexus Core
 * 
 * This is a standalone process that manages a PTY session.
 * It allows the main Orchestrator to restart without killing the CLI process.
 */

const [command, argsJson, portStr, cwd] = process.argv.slice(2);
const args = JSON.parse(argsJson);
const port = parseInt(portStr);

console.log(`[Sidecar] Starting PTY: ${command} ${args.join(' ')} on port ${port}`);

const ptyProcess = pty.spawn(command, args, {
  name: 'xterm-color',
  cols: 80,
  rows: 30,
  cwd: cwd || process.cwd(),
  env: process.env as Record<string, string>
});

const server = createServer((socket) => {
  console.log('[Sidecar] Client connected');

  // Send PTY output to client
  const ptyDisposable = ptyProcess.onData((data) => {
    socket.write(data);
  });

  // Heartbeat to Orchestrator
  const heartbeat = setInterval(() => {
    const telemetry = {
      type: 'HEARTBEAT',
      timestamp: Date.now(),
      pid: ptyProcess.pid,
    };
    socket.write(`TORMENTNEXUS_TELEMETRY:${JSON.stringify(telemetry)}`);
  }, 5000);

  // Send client input to PTY
  socket.on('data', (data) => {
    const text = data.toString();
    
    // Check for control messages
    if (text.startsWith('TORMENTNEXUS_CTRL:')) {
      try {
        const ctrl = JSON.parse(text.substring(10));
        if (ctrl.type === 'SET_ENV') {
          console.log(`[Sidecar] Injecting env: ${ctrl.key}=${ctrl.value.substring(0, 4)}...`);
          const isWin = process.platform === 'win32';
          const cmd = isWin 
            ? `$env:${ctrl.key} = '${ctrl.value}';` 
            : `export ${ctrl.key}='${ctrl.value}';`;
          
          ptyProcess.write(`${cmd}\n`);
        }
        return;
      } catch (e) {
        console.error('[Sidecar] Failed to parse control message:', e);
      }
    }

    ptyProcess.write(text);
  });

  socket.on('close', () => {
    console.log('[Sidecar] Client disconnected');
    clearInterval(heartbeat);
    ptyDisposable.dispose();
  });

  socket.on('error', (err) => {
    console.error('[Sidecar] Socket error:', err);
  });
});

server.listen(port, '127.0.0.1', () => {
  console.log(`[Sidecar] Listening on 127.0.0.1:${port}`);
});

ptyProcess.onExit(({ exitCode, signal }) => {
  console.log(`[Sidecar] PTY exited with code ${exitCode}, signal ${signal}`);
  process.exit(exitCode);
});

// Handle graceful shutdown
process.on('SIGTERM', () => {
  ptyProcess.kill();
  process.exit(0);
});
