import {
    detectLocalExecutionEnvironment,
    type DetectedExecutionShell,
    type LocalExecutionEnvironment,
    type ShellFamily,
} from './execution-environment.js';

export type SessionExecutionProfile = 'auto' | 'powershell' | 'posix' | 'compatibility';

export interface SessionExecutionPolicy {
    requestedProfile: SessionExecutionProfile;
    effectiveProfile: Exclude<SessionExecutionProfile, 'auto'> | 'fallback';
    shellId: string | null;
    shellLabel: string | null;
    shellFamily: ShellFamily | null;
    shellPath: string | null;
    supportsPowerShell: boolean;
    supportsPosixShell: boolean;
    reason: string;
}

function findFirstVerifiedShell(
    environment: LocalExecutionEnvironment,
    predicate: (shell: DetectedExecutionShell) => boolean,
): DetectedExecutionShell | null {
    return environment.shells.find((shell) => shell.verified && predicate(shell)) ?? null;
}

function getPreferredVerifiedShell(environment: LocalExecutionEnvironment): DetectedExecutionShell | null {
    if (environment.summary.preferredShellId) {
        const preferred = environment.shells.find(
            (shell) => shell.id === environment.summary.preferredShellId && shell.verified,
        );
        if (preferred) {
            return preferred;
        }
    }

    return environment.shells.find((shell) => shell.verified) ?? null;
}

function toPolicy(
    environment: LocalExecutionEnvironment,
    requestedProfile: SessionExecutionProfile,
    effectiveProfile: SessionExecutionPolicy['effectiveProfile'],
    shell: DetectedExecutionShell | null,
    reason: string,
): SessionExecutionPolicy {
    return {
        requestedProfile,
        effectiveProfile,
        shellId: shell?.id ?? null,
        shellLabel: shell?.name ?? null,
        shellFamily: shell?.family ?? null,
        shellPath: shell?.resolvedPath ?? null,
        supportsPowerShell: environment.summary.supportsPowerShell,
        supportsPosixShell: environment.summary.supportsPosixShell,
        reason,
    };
}

export function selectSessionExecutionPolicy(
    environment: LocalExecutionEnvironment,
    requestedProfile: SessionExecutionProfile = 'auto',
): SessionExecutionPolicy {
    const preferred = getPreferredVerifiedShell(environment);
    const powerShell = findFirstVerifiedShell(environment, (shell) => shell.family === 'powershell');
    const posixShell = findFirstVerifiedShell(environment, (shell) => shell.family === 'posix' || shell.family === 'wsl');
    const compatibilityShell = environment.os === 'win32'
        ? findFirstVerifiedShell(environment, (shell) => shell.id === 'cmd')
        : preferred;

    switch (requestedProfile) {
        case 'powershell': {
            if (powerShell) {
                return toPolicy(
                    environment,
                    requestedProfile,
                    'powershell',
                    powerShell,
                    `${powerShell.name} selected because the session explicitly requested a PowerShell-native execution profile.`,
                );
            }

            return toPolicy(
                environment,
                requestedProfile,
                'fallback',
                preferred,
                preferred
                    ? `A PowerShell shell was requested, but none verified; falling back to ${preferred.name}.`
                    : 'A PowerShell shell was requested, but TormentNexus could not verify any shell on this host.',
            );
        }

        case 'posix': {
            if (posixShell) {
                return toPolicy(
                    environment,
                    requestedProfile,
                    'posix',
                    posixShell,
                    `${posixShell.name} selected because the session requested POSIX-style pipelines or Unix-first tooling.`,
                );
            }

            return toPolicy(
                environment,
                requestedProfile,
                'fallback',
                preferred,
                preferred
                    ? `A POSIX shell was requested, but none verified; falling back to ${preferred.name}.`
                    : 'A POSIX shell was requested, but TormentNexus could not verify any shell on this host.',
            );
        }

        case 'compatibility': {
            return toPolicy(
                environment,
                requestedProfile,
                compatibilityShell?.id === 'cmd' ? 'compatibility' : 'fallback',
                compatibilityShell,
                compatibilityShell
                    ? `${compatibilityShell.name} selected for the most conservative compatibility posture on this host.`
                    : 'Compatibility mode was requested, but TormentNexus could not verify any shell on this host.',
            );
        }

        case 'auto':
        default: {
            if (environment.os === 'win32' && powerShell) {
                return toPolicy(
                    environment,
                    requestedProfile,
                    'powershell',
                    powerShell,
                    `${powerShell.name} selected automatically as TormentNexus's preferred Windows execution shell for general harness supervision.`,
                );
            }

            if (posixShell) {
                return toPolicy(
                    environment,
                    requestedProfile,
                    'posix',
                    posixShell,
                    `${posixShell.name} selected automatically because no verified PowerShell shell was preferred for this host.`,
                );
            }

            return toPolicy(
                environment,
                requestedProfile,
                'fallback',
                preferred,
                preferred
                    ? `${preferred.name} selected automatically as the only verified shell TormentNexus can trust on this host.`
                    : 'Auto execution profile could not verify a runnable shell on this host.',
            );
        }
    }
}

export function buildExecutionPolicyEnv(policy: SessionExecutionPolicy | null): Record<string, string> {
    if (!policy) {
        return {};
    }

    const env: Record<string, string> = {
        TORMENTNEXUS_EXECUTION_PROFILE_REQUESTED: policy.requestedProfile,
        TORMENTNEXUS_EXECUTION_PROFILE_EFFECTIVE: policy.effectiveProfile,
        TORMENTNEXUS_EXECUTION_SHELL_ID: policy.shellId ?? '',
        TORMENTNEXUS_EXECUTION_SHELL_LABEL: policy.shellLabel ?? '',
        TORMENTNEXUS_EXECUTION_SHELL_FAMILY: policy.shellFamily ?? '',
        TORMENTNEXUS_EXECUTION_SHELL_PATH: policy.shellPath ?? '',
        TORMENTNEXUS_EXECUTION_POLICY_REASON: policy.reason,
        TORMENTNEXUS_SUPPORTS_POWERSHELL: policy.supportsPowerShell ? '1' : '0',
        TORMENTNEXUS_SUPPORTS_POSIX_SHELL: policy.supportsPosixShell ? '1' : '0',
    };

    if (policy.shellPath) {
        env.SHELL = policy.shellPath;
        env.npm_config_script_shell = policy.shellPath;
    }

    if (policy.shellFamily === 'cmd' && policy.shellPath) {
        env.COMSPEC = policy.shellPath;
    }

    if (policy.shellFamily === 'powershell' && policy.shellPath) {
        env.PSModulePath = process.env.PSModulePath ?? '';
    }

    return env;
}

export async function detectSessionExecutionPolicy(
    requestedProfile: SessionExecutionProfile = 'auto',
): Promise<SessionExecutionPolicy> {
    const environment = await detectLocalExecutionEnvironment();
    return selectSessionExecutionPolicy(environment, requestedProfile);
}