import type { Command } from 'commander';
import fs from 'fs';
import os from 'os';
import path from 'path';
import {
  clearDaemonInfo,
  DEFAULT_COCKPIT_PORT,
  DEFAULT_DAEMON_PORT,
  readDaemonInfo,
} from '../../lib/daemon-discovery.js';
import {
  findPidsOnPort,
  isPortInUse,
  killAndVerify,
} from '../port-utils.js';
import { c, header, sym } from '../ui.js';

/**
 * Two-stage shutdown per managed process: SIGTERM, wait up to 1.5s,
 * escalate to SIGKILL if still alive. Don't unlink the pidfile until
 * the process is confirmed dead — otherwise an orphan that ignores
 * SIGTERM keeps running while we forget about it (the bug behind the
 * "stale next-server serving 500s" incident on 2026-05-03).
 *
 * v0.8: PID lookup goes through `~/.chorus/daemon.json` first, falling
 * back to the legacy daemon.pid / web.pid files for compat with
 * v0.7-and-earlier installs. Belt-and-braces port sweep targets the
 * recorded ports (or defaults if no daemon.json).
 */
export function registerStopCommand(program: Command): void {
  program
    .command('stop')
    .description('Stop the Chorus daemon and cockpit')
    .action(async () => {
      try {
        const chorusDir = path.join(os.homedir(), '.chorus');
        const daemonPidFile = path.join(chorusDir, 'daemon.pid');
        const webPidFile = path.join(chorusDir, 'web.pid');

        const info = readDaemonInfo();
        const daemonPort = info?.daemonPort ?? DEFAULT_DAEMON_PORT;
        const cockpitPort = info?.cockpitPort ?? DEFAULT_COCKPIT_PORT;

        const daemonPidfileExists = fs.existsSync(daemonPidFile);
        const webPidfileExists = fs.existsSync(webPidFile);
        const daemonPortInUse = await isPortInUse(daemonPort);
        const cockpitPortInUse = await isPortInUse(cockpitPort);

        if (
          !info &&
          !daemonPidfileExists &&
          !webPidfileExists &&
          !daemonPortInUse &&
          !cockpitPortInUse
        ) {
          console.log('');
          console.log(header(sym.info, 'Chorus is not running', 'nothing to stop'));
          console.log('');
          return;
        }

        console.log('');
        console.log(header(sym.pointer, 'Stopping Chorus...'));
        console.log('');

        // Prefer daemon.json (v0.8), fall back to pidfiles (v0.7).
        if (info) {
          await stopByPid('Daemon', info.daemonPid);
          if (info.cockpitPid) {
            await stopByPid('Cockpit', info.cockpitPid);
          }
        }
        await stopByPidFile('Daemon', daemonPidFile);
        await stopByPidFile('Cockpit', webPidFile);

        // Port-based sweep — kills any chorus-owned listener that
        // escaped the pidfile path.
        await sweepPort(daemonPort, 'Daemon');
        await sweepPort(cockpitPort, 'Cockpit');

        clearDaemonInfo();

        console.log('');
      } catch (error) {
        console.error(`${sym.err} ${c.red('Error stopping chorus:')}`, error);
        process.exit(1);
      }
    });
}

async function stopByPid(label: string, pid: number): Promise<void> {
  if (!Number.isFinite(pid) || pid <= 0) return;
  const dead = await killAndVerify(pid, label);
  if (dead) {
    console.log(`  ${sym.ok} ${label.padEnd(7)} ${c.dim(`(PID ${pid})`)}`);
  }
}

async function stopByPidFile(label: string, pidFile: string): Promise<void> {
  if (!fs.existsSync(pidFile)) return;
  const pid = parseInt(fs.readFileSync(pidFile, 'utf-8'), 10);
  if (!Number.isFinite(pid) || pid <= 0) {
    fs.unlinkSync(pidFile);
    return;
  }
  const dead = await killAndVerify(pid, label);
  if (dead) {
    console.log(`  ${sym.ok} ${label.padEnd(7)} ${c.dim(`(PID ${pid})`)}`);
    try {
      fs.unlinkSync(pidFile);
    } catch {
      /* already gone */
    }
  }
}

async function sweepPort(port: number, label: string): Promise<void> {
  const pids = findPidsOnPort(port);
  for (const pid of pids) {
    const dead = await killAndVerify(pid, `${label} orphan`);
    if (dead) {
      console.log(
        `  ${sym.ok} ${label.padEnd(7)} ${c.dim(`(orphan PID ${pid} on :${port})`)}`,
      );
    }
  }
}
