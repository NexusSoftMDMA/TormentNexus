#!/usr/bin/env node
// Printed once after `npm install -g chorus-codes`.
// Skipped when CHORUS_SKIP_POSTINSTALL=1 (used by CI / dev linking).

if (process.env.CHORUS_SKIP_POSTINSTALL === "1") process.exit(0);

const lines = [
  "",
  "🎉 Chorus installed!",
  "",
  "Two more commands to get going:",
  "",
  "  ➡️  chorus init    — register MCP with your editors, seed templates, detect CLIs",
  "  ➡️  chorus start   — bring up the daemon + cockpit UI on http://127.0.0.1:5050",
  "",
  "Docs: https://chorus.codes",
  "",
];

console.log(lines.join("\n"));
