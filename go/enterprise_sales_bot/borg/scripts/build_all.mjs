#!/usr/bin/env node

/**
 * Cross-platform TormentNexus build orchestrator.
 *
 * Why this exists:
 * - Root Turbo builds only cover packages included in `pnpm-workspace.yaml`.
 * - Some extension deliverables live outside the root workspace (`apps/tormentnexus-extension`).
 * - The JetBrains plugin uses Gradle, so it needs a native build step.
 * - The richer browser extension has separate Chromium and Firefox modes that would
 *   otherwise overwrite the same `dist/` directory.
 */

import { spawnSync } from "node:child_process";
import { cpSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const args = new Set(process.argv.slice(2));
const workspaceOnly = args.has("--workspace-only");
const extensionsOnly = args.has("--extensions-only");

if (workspaceOnly && extensionsOnly) {
  console.error("[build] Choose either --workspace-only or --extensions-only, not both.");
  process.exit(1);
}

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const pnpmCommand = process.platform === "win32" ? "pnpm.cmd" : "pnpm";

function printStep(message) {
  console.log(`\n[build] ${message}`);
}

function formatFailure(result) {
  return [
    result.error ? `error=${String(result.error)}` : null,
    typeof result.status === "number" ? `status=${result.status}` : null,
    result.signal ? `signal=${result.signal}` : null,
  ]
    .filter(Boolean)
    .join(" ");
}

function fail(message, result) {
  const suffix = result ? ` (${formatFailure(result)})` : "";
  throw new Error(`${message}${suffix}`);
}

function run(command, commandArgs, options = {}) {
  const result = spawnSync(command, commandArgs, {
    cwd: repoRoot,
    encoding: "utf-8",
    stdio: "inherit",
    shell: false,
    env: process.env,
    ...options,
  });

  return result;
}

function runPnpm(commandArgs, options = {}) {
  const direct = run(pnpmCommand, commandArgs, options);

  if (!direct.error) {
    return direct;
  }

  return run(`pnpm ${commandArgs.join(" ")}`, [], {
    ...options,
    shell: true,
  });
}

function getProcessList() {
  if (process.platform === "win32") {
    const result = spawnSync(
      "powershell.exe",
      [
        "-NoProfile",
        "-Command",
        "Get-CimInstance Win32_Process | Select-Object ProcessId, Name, CommandLine | ConvertTo-Json -Compress",
      ],
      {
        cwd: repoRoot,
        encoding: "utf-8",
        stdio: "pipe",
        shell: false,
        env: process.env,
      },
    );

    if (result.error || (result.status ?? 1) !== 0 || !result.stdout.trim()) {
      return [];
    }

    try {
      const parsed = JSON.parse(result.stdout);
      const rows = Array.isArray(parsed) ? parsed : [parsed];
      return rows.map((row) => ({
        pid: Number(row.ProcessId),
        name: String(row.Name ?? ""),
        commandLine: String(row.CommandLine ?? ""),
      }));
    } catch {
      return [];
    }
  }

  const result = spawnSync("ps", ["-ax", "-o", "pid=,command="], {
    cwd: repoRoot,
    encoding: "utf-8",
    stdio: "pipe",
    shell: false,
    env: process.env,
  });

  if (result.error || (result.status ?? 1) !== 0 || !result.stdout.trim()) {
    return [];
  }

  return result.stdout
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const match = line.match(/^(\d+)\s+(.*)$/);
      if (!match) {
        return null;
      }

      return {
        pid: Number(match[1]),
        name: "",
        commandLine: match[2],
      };
    })
    .filter(Boolean);
}

function hasActiveNextBuild(appDir) {
  const appDirLower = appDir.toLowerCase();
  const appPathFragment = appDirLower.replace(/\\/g, "/");

  return getProcessList().some((proc) => {
    const command = proc.commandLine.toLowerCase().replace(/\\/g, "/");
    return command.includes("next build")
      && (command.includes(appPathFragment) || command.includes("apps/web"));
  });
}

function clearStaleNextBuildLock(appRelativeDir) {
  const appDir = path.join(repoRoot, ...appRelativeDir.split("/"));
  const lockPath = path.join(appDir, ".next", "lock");

  if (!existsSync(lockPath)) {
    return;
  }

  if (hasActiveNextBuild(appDir)) {
    printStep(`Detected an active Next.js build for ${appRelativeDir}; leaving ${path.relative(repoRoot, lockPath)} in place.`);
    return;
  }

  rmSync(lockPath, { force: true });
  printStep(`Removed stale Next.js lock at ${path.relative(repoRoot, lockPath)}.`);
}

function copyDirectory(sourceDir, targetDir) {
  rmSync(targetDir, { recursive: true, force: true });
  mkdirSync(path.dirname(targetDir), { recursive: true });
  cpSync(sourceDir, targetDir, { recursive: true, force: true });
}

function directoryHasMergeMarkers(rootDir) {
  if (!existsSync(rootDir)) {
    return false;
  }

  const entries = readdirSync(rootDir, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(rootDir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === "node_modules" || entry.name === ".git" || entry.name === ".turbo" || entry.name === "dist") {
        continue;
      }
      if (directoryHasMergeMarkers(fullPath)) {
        return true;
      }
      continue;
    }

    if (!/\.(?:[cm]?[jt]s|tsx|jsx|json)$/i.test(entry.name)) {
      continue;
    }

    const content = readFileSync(fullPath, "utf-8");
    if (content.includes("<<<<<<< HEAD") || content.includes(">>>>>>> upstream/main")) {
      return true;
    }
  }

  return false;
}

function getWorkspacePackageName(packageRoot, fallbackName) {
  const packageJsonPath = path.join(packageRoot, "package.json");
  if (!existsSync(packageJsonPath)) {
    return fallbackName;
  }

  try {
    const parsed = JSON.parse(readFileSync(packageJsonPath, "utf-8"));
    if (typeof parsed.name === "string" && parsed.name.trim()) {
      return parsed.name.trim();
    }
  } catch {
    // Fall back to the caller-provided logical label when package.json is malformed.
  }

  return fallbackName;
}

function runWorkspaceBuild() {
  printStep("Running Turbo workspace build (includes VS Code and browser-extension package workspaces)...");
  clearStaleNextBuildLock("apps/web");

  const turboArgs = [
    "exec",
    "turbo",
    "run",
    "build",
    "--filter=!@repo/*",
  ];

  const claudeMemRoot = path.join(repoRoot, "packages", "tormentnexus");
  const claudeMemPackageName = getWorkspacePackageName(claudeMemRoot, "tormentnexus");
  const shouldRequireTormentNexus = process.env.TORMENTNEXUS_REQUIRE_TORMENTNEXUS_BUILD === "true";
  const claudeMemHasMergeMarkers = directoryHasMergeMarkers(path.join(claudeMemRoot, "src"))
    || directoryHasMergeMarkers(path.join(claudeMemRoot, "scripts"));

  if (claudeMemHasMergeMarkers && !shouldRequireTormentNexus) {
    const exclusionTargets = Array.from(new Set(["tormentnexus", claudeMemPackageName]));
    printStep(`Detected unresolved merge markers in packages/tormentnexus; excluding ${exclusionTargets.join(", ")} from the workspace build so TormentNexus can still start.`);
    for (const target of exclusionTargets) {
      turboArgs.push(`--filter=!${target}`);
    }
  }

  // Exclude tormentnexus-extension from workspace turbo build � its nested pnpm
  // workspace needs a separate install step. Non-blocking for dashboard startup.
  turboArgs.push(`--filter=!tormentnexus-extension`);

  const result = runPnpm(
    turboArgs,
    {
      cwd: repoRoot,
      env: {
        ...process.env,
        TURBO_DAEMON: "false",
      },
    },
  );

  if ((result.status ?? 1) !== 0) {
    fail("Workspace build failed", result);
  }
}

function runBrowserExtensionBuilds() {
  const extensionRoot = path.join(repoRoot, "apps", "tormentnexus-extension");
  const distDir = path.join(extensionRoot, "dist");
  const chromiumDistDir = path.join(extensionRoot, "dist-chromium");
  const firefoxDistDir = path.join(extensionRoot, "dist-firefox");
  const snapshotRoot = path.join(extensionRoot, ".build-artifacts");
  const chromiumSnapshotDir = path.join(snapshotRoot, "dist-chromium-snapshot");

  if (!existsSync(extensionRoot)) {
    printStep("Skipping browser-extension aggregate build because `apps/tormentnexus-extension` is not present.");
    return;
  }

  mkdirSync(snapshotRoot, { recursive: true });

  printStep("Installing TormentNexus browser-extension workspace dependencies...");
  const installResult = runPnpm(["install", "--frozen-lockfile"], {
    cwd: extensionRoot,
    env: {
      ...process.env,
      CI: process.env.CI ?? "true",
    },
  });

  if ((installResult.status ?? 1) !== 0) {
    printStep("Browser-extension install failed; retrying with --ignore-scripts to bypass flaky lifecycle hooks.");

    const fallbackInstallResult = runPnpm(["install", "--frozen-lockfile", "--ignore-scripts"], {
      cwd: extensionRoot,
      env: {
        ...process.env,
        CI: process.env.CI ?? "true",
      },
    });

    if ((fallbackInstallResult.status ?? 1) !== 0) {
      fail("Browser-extension dependency install failed", fallbackInstallResult);
    }

    printStep("Browser-extension dependencies installed via --ignore-scripts fallback.");
  }

  // Update .env file to set CLI_CEB_DEV=false for production builds
  const envPath = path.join(extensionRoot, ".env");
  if (existsSync(envPath)) {
    try {
      let envContent = readFileSync(envPath, "utf-8");
      envContent = envContent.replace(/CLI_CEB_DEV=true/g, "CLI_CEB_DEV=false");
      writeFileSync(envPath, envContent, "utf-8");
      printStep("Enforced CLI_CEB_DEV=false in browser extension .env file");
    } catch (err) {
      console.error("[build] Failed to write to .env:", err.message);
    }
  }

  printStep("Building TormentNexus browser extension for Chromium/Chrome/Edge...");
  const chromiumBuild = runPnpm(["run", "base-build"], {
    cwd: extensionRoot,
    env: {
      ...process.env,
      NODE_ENV: "production",
      CLI_CEB_DEV: "false",
      CLI_CEB_FIREFOX: "false",
    },
  });

  if ((chromiumBuild.status ?? 1) !== 0) {
    fail("Chromium browser-extension build failed", chromiumBuild);
  }

  if (!existsSync(distDir)) {
    fail(`Expected browser-extension output at ${distDir}`);
  }

  copyDirectory(distDir, chromiumDistDir);
  copyDirectory(distDir, chromiumSnapshotDir);

  printStep("Building TormentNexus browser extension for Firefox...");
  const firefoxBuild = runPnpm(["run", "base-build"], {
    cwd: extensionRoot,
    env: {
      ...process.env,
      NODE_ENV: "production",
      CLI_CEB_DEV: "false",
      CLI_CEB_FIREFOX: "true",
    },
  });

  if ((firefoxBuild.status ?? 1) !== 0) {
    fail("Firefox browser-extension build failed", firefoxBuild);
  }

  if (!existsSync(distDir)) {
    fail(`Expected Firefox browser-extension output at ${distDir}`);
  }

  copyDirectory(distDir, firefoxDistDir);

  // Restore the default `dist/` to the Chromium build so callers that already
  // expect `pnpm build` -> Chromium continue to work, while keeping a preserved
  // Firefox artifact beside it.
  copyDirectory(chromiumSnapshotDir, distDir);
  rmSync(snapshotRoot, { recursive: true, force: true });
}

function detectGradleCommand(jetbrainsRoot) {
  const wrapperCandidates = process.platform === "win32"
    ? [path.join(jetbrainsRoot, "gradlew.bat")]
    : [path.join(jetbrainsRoot, "gradlew")];

  const wrapper = wrapperCandidates.find((candidate) => existsSync(candidate));
  if (wrapper) {
    return { command: wrapper, args: ["buildPlugin"] };
  }

  const fallbackCandidates = process.platform === "win32"
    ? ["gradle", "gradle.bat"]
    : ["gradle"];

  for (const fallbackCommand of fallbackCandidates) {
    const probe = spawnSync(fallbackCommand, ["--version"], {
      cwd: jetbrainsRoot,
      stdio: "ignore",
      shell: false,
      env: process.env,
    });

    if (!probe.error && (probe.status ?? 1) === 0) {
      return { command: fallbackCommand, args: ["buildPlugin"] };
    }
  }

  return null;
}

function runJetBrainsBuild() {
  const jetbrainsRoot = path.join(repoRoot, "packages", "jetbrains");

  if (!existsSync(jetbrainsRoot)) {
    printStep("Skipping JetBrains plugin build because `packages/jetbrains` is not present.");
    return;
  }

  const gradle = detectGradleCommand(jetbrainsRoot);
  if (!gradle) {
    const strictJetBrainsBuild = process.env.TORMENTNEXUS_REQUIRE_JETBRAINS_BUILD === "true";
    const message = "Skipping JetBrains plugin build because Gradle is not available. Install Gradle or add a Gradle wrapper under `packages/jetbrains`, or set TORMENTNEXUS_REQUIRE_JETBRAINS_BUILD=true to fail instead.";

    if (strictJetBrainsBuild) {
      fail(message);
    }

    printStep(message);
    return;
  }

  printStep("Building JetBrains plugin artifact...");
  const result = run(gradle.command, gradle.args, {
    cwd: jetbrainsRoot,
    env: {
      ...process.env,
      CI: process.env.CI ?? "true",
    },
  });

  if ((result.status ?? 1) !== 0) {
    fail("JetBrains plugin build failed", result);
  }
}

async function main() {
  if (!extensionsOnly) {
    runWorkspaceBuild();
  }

  if (!workspaceOnly) {
    runBrowserExtensionBuilds();
    runJetBrainsBuild();
  }

  printStep("Build completed successfully ✅");
}

main().catch((error) => {
  console.error(`\n[build] ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
});
