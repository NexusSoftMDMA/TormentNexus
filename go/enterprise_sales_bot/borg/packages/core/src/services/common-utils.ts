
/**
 * Environment variables to inherit by default, if an environment is not explicitly given.
 */
export const DEFAULT_INHERITED_ENV_VARS =
    process.platform === "win32"
        ? [
            "APPDATA",
            "HOMEDRIVE",
            "HOMEPATH",
            "LOCALAPPDATA",
            "PATH",
            "PROCESSOR_ARCHITECTURE",
            "SYSTEMDRIVE",
            "SYSTEMROOT",
            "TEMP",
            "USERNAME",
            "USERPROFILE",
            "PROGRAMFILES",
        ]
        : /* list inspired by the default env inheritance of sudo */
        [
            "HOME",
            "LOGNAME",
            "PATH",
            "SHELL",
            "TERM",
            "USER",
            // SSL/Certificate variables
            "NODE_EXTRA_CA_CERTS",
            "NODE_TLS_REJECT_UNAUTHORIZED",
            "SSL_CERT_FILE",
            "CERT_FILE",
            "REQUESTS_CA_BUNDLE",
            "REQUESTS_CERT_FILE",
            "CURL_CA_BUNDLE",
            "PIP_CERT",
            "UV_CERT",
            "PYTHONHTTPSVERIFY",
            // Proxy variables
            "HTTP_PROXY",
            "HTTPS_PROXY",
            "NO_PROXY",
            "http_proxy",
            "https_proxy",
            "no_proxy",
        ];

export type IOType = "overlapped" | "pipe" | "ignore" | "inherit";

/**
 * Returns a default environment object including only environment variables deemed safe to inherit.
 */
export function getDefaultEnvironment(): Record<string, string> {
    const env: Record<string, string> = {};

    for (const key of DEFAULT_INHERITED_ENV_VARS) {
        const value = process.env[key];
        if (value === undefined) {
            continue;
        }

        if (value.startsWith("()")) {
            continue;
        }

        env[key] = value;
    }

    return env;
}

export function resolveEnvVariables(
    envObject: Record<string, unknown>,
): Record<string, unknown> {
    const resolved: Record<string, unknown> = {};

    for (const [key, value] of Object.entries(envObject)) {
        if (
            typeof value === "string" &&
            value.startsWith("${") &&
            value.endsWith("}")
        ) {
            const varName = value.slice(2, -1);
            if (process.env[varName]) {
                resolved[key] = process.env[varName];
            } else {
                resolved[key] = value;
                console.warn(
                    `Environment variable not found: ${varName}, keeping original value: ${value}`,
                );
            }
        } else {
            resolved[key] = value;
        }
    }

    return resolved;
}

/**
 * Sanitize a string to be safe for use in identifiers/paths.
 * Allows alphanumeric, underscore, and hyphen.
 */
export function sanitizeName(name: string): string {
    return name.replace(/[^a-zA-Z0-9_-]/g, "");
}
