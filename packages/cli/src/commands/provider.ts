/**
 * `tormentnexus provider` - AI provider management
 *
 * Configure and manage AI model providers, API keys, OAuth logins,
 * quota tracking, and automatic model fallback chains.
 *
 * @example
 *   tormentnexus provider list         # List providers with quota status
 *   tormentnexus provider add openai   # Add a new provider
 *   tormentnexus provider quota        # Show all quota/billing info
 */

import type { Command } from 'commander';

export function registerProviderCommand(program: Command): void {
  const provider = program
    .command('provider')
    .alias('prov')
    .description('Providers — manage AI providers, API keys, OAuth, quotas, and model fallback');

  provider
    .command('list')
    .description('List all configured providers with status and quota usage')
    .option('--json', 'Output as JSON')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;

      // Try to get provider data from API
      let providers: any[] = [];
      try {
        const res = await fetch('http://127.0.0.1:4100/trpc/settings.get', {
          signal: AbortSignal.timeout(3000),
        });
        if (res.ok) {
          const json = await res.json();
          const settings = json?.result?.data;
          providers = settings?.providers ?? [];
        }
      } catch {}

      if (opts.json) {
        console.log(JSON.stringify({ providers }, null, 2));
        return;
      }

      const table = new Table({
        head: ['Provider', 'Status', 'Auth', 'Default Model', 'Tasks', 'Quota', 'Resets'],
        style: { head: ['cyan'] },
      });

      console.log(chalk.bold.cyan('\n  AI Providers\n'));

      if (providers.length === 0) {
        // Check environment variables for known API keys
        const envProviders: string[] = [];
        const envMap: Record<string, string> = {
          OPENAI_API_KEY: 'openai',
          ANTHROPIC_API_KEY: 'anthropic',
          GOOGLE_API_KEY: 'google',
          GEMINI_API_KEY: 'google',
          XAI_API_KEY: 'xai',
          DEEPSEEK_API_KEY: 'deepseek',
          MISTRAL_API_KEY: 'mistral',
          OPENROUTER_API_KEY: 'openrouter',
        };

        // Enrich with Go sidecar provider catalog
        let goProviders: any[] = [];
        try {
          const gRes = await fetch('http://127.0.0.1:4300/api/providers/catalog', { signal: AbortSignal.timeout(3000) });
          if (gRes.ok) goProviders = (await gRes.json()).data ?? [];
        } catch {}
        const goMap = new Map(goProviders.map((p: any) => [p.provider, p]));

        for (const [env, provider] of Object.entries(envMap)) {
          if (process.env[env]) {
            envProviders.push(provider);
            const goInfo = goMap.get(provider);
            const defaultModel = goInfo?.defaultModel ?? '-';
            const tasks = goInfo?.preferredTasks?.join(', ') ?? '-';
            table.push([
              provider,
              chalk.green('● Available'),
              chalk.dim('env: ' + env),
              defaultModel,
              tasks,
              '-',
              '-',
            ]);
          }
        }
        if (envProviders.length > 0) {
          console.log(table.toString());
          console.log(chalk.dim(`\n  ${envProviders.length} provider(s) detected from environment variables`));
          console.log(chalk.dim(`  Use ${chalk.cyan('tormentnexus provider add <name>')} to configure explicitly\n`));
        } else {
          console.log(chalk.dim('  No providers configured. Use `tormentnexus provider add` to add one.\n'));
          console.log(chalk.dim('  Supported: openai, anthropic, google, xai, deepseek, mistral,'));
          console.log(chalk.dim('             openrouter, copilot, antigravity, local\n'));
        }
      } else {
        for (const p of providers) {
          table.push([
            p.name ?? p,
            p.connected ? chalk.green('● Connected') : chalk.yellow('○ Disconnected'),
            p.authType ?? 'api-key',
            p.models?.length ?? '-',
            p.quotaUsed ?? '-',
            p.quotaLimit ?? '-',
            p.quotaResets ?? '-',
          ]);
        }
        console.log(table.toString() + '\n');
      }
    });

  provider
    .command('add <name>')
    .description('Add and configure a new AI provider')
    .option('-k, --api-key <key>', 'API key (or set via env var)')
    .option('--oauth', 'Use OAuth login flow (for subscription services)')
    .option('--base-url <url>', 'Custom API base URL')
    .option('--models <models...>', 'Enabled models for this provider')
    .addHelpText('after', `
Supported providers:
  openai       OpenAI API (GPT-5, Codex, etc.)
  anthropic    Anthropic API (Claude Opus, Sonnet)
  google       Google AI API (Gemini Pro, Flash)
  xai          xAI API (Grok)
  deepseek     DeepSeek API
  mistral      Mistral API
  openrouter   OpenRouter (multi-provider gateway)
  copilot      GitHub Copilot (via OAuth/PAT)
  antigravity  Antigravity (via Google OAuth)
  local        Local models (Ollama, LM Studio, etc.)

OAuth-capable subscription services:
  $ tormentnexus provider add anthropic --oauth    # Claude Max/Pro subscription
  $ tormentnexus provider add google --oauth       # Google AI Plus subscription
  $ tormentnexus provider add copilot --oauth      # Copilot Premium Plus
  $ tormentnexus provider add openai --oauth       # ChatGPT Plus subscription
    `)
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;

      if (opts.oauth) {
        console.log(chalk.yellow(`  Starting OAuth flow for ${name}...`));
        console.log(chalk.dim('  (OAuth login not yet implemented)'));
        return;
      }

      if (!opts.apiKey) {
        console.log(chalk.yellow(`  Set API key via environment variable or --api-key flag`));
        console.log(chalk.dim(`  Example: export ${name.toUpperCase()}_API_KEY=sk-...`));
        return;
      }

      // Write provider config to ~/.tormentnexus/config.jsonc
      const { readFileSync, writeFileSync, mkdirSync } = await import('fs');
      const { resolve, dirname } = await import('path');
      const { homedir } = await import('os');
      const configDir = process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), '.tormentnexus');
      const configPath = resolve(configDir, 'config.jsonc');

      let config: Record<string, any> = {};
      try {
        const raw = readFileSync(configPath, 'utf8');
        config = JSON.parse(raw.replace(/\/\/.*$/gm, '').replace(/\/\*[\s\S]*?\*\//g, ''));
      } catch {}

      if (!config.providers) config.providers = {};
      config.providers[name] = {
        apiKey: opts.apiKey,
        baseUrl: opts.baseUrl,
        models: opts.models,
        addedAt: new Date().toISOString(),
      };

      mkdirSync(dirname(configPath), { recursive: true });
      writeFileSync(configPath, JSON.stringify(config, null, 2) + '\n');
      console.log(chalk.green(`  ✓ Provider '${name}' configured`));
      if (opts.baseUrl) console.log(chalk.dim(`    Base URL: ${opts.baseUrl}`));
      if (opts.models?.length) console.log(chalk.dim(`    Models: ${opts.models.join(', ')}`));
    });

  provider
    .command('remove <name>')
    .description('Remove a configured provider')
    .option('-f, --force', 'Skip confirmation')
    .action(async (name, opts) => {
      const chalk = (await import('chalk')).default;
      const { readFileSync, writeFileSync } = await import('fs');
      const { resolve, dirname } = await import('path');
      const { homedir } = await import('os');

      const configDir = process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), '.tormentnexus');
      const configPath = resolve(configDir, 'config.jsonc');

      let config: Record<string, any> = {};
      try {
        const raw = readFileSync(configPath, 'utf8');
        config = JSON.parse(raw.replace(/\/\/.*$/gm, '').replace(/\/\*[\s\S]*?\*\//g, ''));
      } catch {}

      if (config.providers?.[name]) {
        delete config.providers[name];
        writeFileSync(configPath, JSON.stringify(config, null, 2) + '\n');
        console.log(chalk.green(`  ✓ Provider '${name}' removed from config`));
      } else {
        console.log(chalk.yellow(`  Provider '${name}' not found in config`));
        console.log(chalk.dim(`    It may be detected from environment variables only`));
      }
    });

  provider
    .command('quota')
    .description('Show quota and billing information for all providers')
    .option('--json', 'Output as JSON')
    .option('--provider <name>', 'Show quota for specific provider')
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      const Table = (await import('cli-table3')).default;

      console.log(chalk.bold.cyan('\n  Provider Quotas & Billing\n'));

      // Check known providers
      const providers = [
        { name: 'OpenAI', envKey: 'OPENAI_API_KEY' },
        { name: 'Anthropic', envKey: 'ANTHROPIC_API_KEY' },
        { name: 'Google', envKey: 'GOOGLE_API_KEY' },
        { name: 'xAI', envKey: 'XAI_API_KEY' },
        { name: 'DeepSeek', envKey: 'DEEPSEEK_API_KEY' },
        { name: 'Mistral', envKey: 'MISTRAL_API_KEY' },
        { name: 'OpenRouter', envKey: 'OPENROUTER_API_KEY' },
      ];

      if (opts.provider) {
        const filtered = providers.filter(p => p.name.toLowerCase() === opts.provider.toLowerCase());
        if (filtered.length === 0) {
          console.log(chalk.red(`  Unknown provider: ${opts.provider}`));
          return;
        }
      }

      const table = new Table({
        head: ['Provider', 'API Key', 'Status'],
        style: { head: ['cyan'] },
      });

      for (const p of providers) {
        if (opts.provider && p.name.toLowerCase() !== opts.provider.toLowerCase()) continue;
        const key = process.env[p.envKey];
        if (key) {
          const masked = key.substring(0, 8) + '...' + key.substring(key.length - 4);
          table.push([p.name, chalk.dim(masked), chalk.green('● Active')]);
        } else {
          table.push([p.name, chalk.dim('not set'), chalk.dim('○ Inactive')]);
        }
      }

      console.log(table.toString());
      console.log(chalk.dim('\n  Quota tracking requires billing dashboard integration.'));
      console.log(chalk.dim('  Use tormentnexus provider test <name> to verify connectivity.\n'));
    });

  provider
    .command('test <name>')
    .description('Test connectivity and authentication for a provider')
    .action(async (name) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.yellow(`  Testing provider: ${name}...`));

      // Map provider names to env vars and test endpoints
      const providerTests: Record<string, { envKey: string; url: string; header: (key: string) => [string, string] }> = {
        openai: { envKey: 'OPENAI_API_KEY', url: 'https://api.openai.com/v1/models', header: (k) => ['Authorization', `Bearer ${k}`] },
        anthropic: { envKey: 'ANTHROPIC_API_KEY', url: 'https://api.anthropic.com/v1/models', header: (k) => ['x-api-key', k] },
        google: { envKey: 'GOOGLE_API_KEY', url: 'https://generativelanguage.googleapis.com/v1beta/models?key=', header: () => ['', ''] },
        gemini: { envKey: 'GEMINI_API_KEY', url: 'https://generativelanguage.googleapis.com/v1beta/models?key=', header: () => ['', ''] },
        xai: { envKey: 'XAI_API_KEY', url: 'https://api.x.ai/v1/models', header: (k) => ['Authorization', `Bearer ${k}`] },
        deepseek: { envKey: 'DEEPSEEK_API_KEY', url: 'https://api.deepseek.com/v1/models', header: (k) => ['Authorization', `Bearer ${k}`] },
        mistral: { envKey: 'MISTRAL_API_KEY', url: 'https://api.mistral.ai/v1/models', header: (k) => ['Authorization', `Bearer ${k}`] },
        openrouter: { envKey: 'OPENROUTER_API_KEY', url: 'https://openrouter.ai/api/v1/models', header: (k) => ['Authorization', `Bearer ${k}`] },
      };

      const test = providerTests[name.toLowerCase()];
      if (!test) {
        console.log(chalk.red(`  ✗ Unknown provider: ${name}`));
        console.log(chalk.dim(`    Supported: ${Object.keys(providerTests).join(', ')}`));
        return;
      }

      const apiKey = process.env[test.envKey];
      if (!apiKey) {
        console.log(chalk.red(`  ✗ No API key found (${test.envKey})`));
        console.log(chalk.dim(`    Set it: export ${test.envKey}=sk-...`));
        return;
      }

      try {
        const start = Date.now();
        let url = test.url;
        const [headerName, headerValue] = test.header(apiKey);
        const headers: Record<string, string> = {};
        if (headerName) headers[headerName] = headerValue;
        if (name.toLowerCase() === 'anthropic') headers['anthropic-version'] = '2023-06-01';

        // Google/Gemini use key in URL
        if (name.toLowerCase() === 'google' || name.toLowerCase() === 'gemini') {
          url += apiKey;
        }

        const res = await fetch(url, { headers, signal: AbortSignal.timeout(10000) });
        const latency = Date.now() - start;

        if (res.ok) {
          const data = await res.json();
          const models = data?.data ?? data?.models ?? [];
          console.log(chalk.green(`  ✓ ${name} authenticated (${latency}ms)`));
          console.log(chalk.dim(`    Models available: ${Array.isArray(models) ? models.length : 'unknown'}`));
          if (Array.isArray(models) && models.length > 0) {
            const sampleNames = models.slice(0, 5).map((m: any) => m.id ?? m.name).join(', ');
            console.log(chalk.dim(`    Sample: ${sampleNames}`));
          }
        } else {
          const body = await res.text().catch(() => '');
          console.log(chalk.red(`  ✗ ${name} returned ${res.status} (${latency}ms)`));
          console.log(chalk.dim(`    ${body.substring(0, 100)}`));
        }
      } catch (e: any) {
        console.log(chalk.red(`  ✗ ${name} error: ${e.message}`));
      }
    });

  provider
    .command('fallback')
    .description('Manage the automatic model fallback chain')
    .option('--show', 'Show current fallback chain')
    .option('--set <models...>', 'Set fallback chain (ordered list)')
    .option('--strategy <strategy>', 'Fallback strategy: priority, cost-optimized, quota-aware, round-robin', 'quota-aware')
    .addHelpText('after', `
The fallback chain determines which model to use when the primary model's
quota is exhausted. Models are tried in order.

Examples:
  $ tormentnexus provider fallback --show
  $ tormentnexus provider fallback --set claude-opus-4 gpt-5.2 gemini-3-pro grok-4
  $ tormentnexus provider fallback --strategy cost-optimized
    `)
    .action(async (opts) => {
      const chalk = (await import('chalk')).default;
      console.log(chalk.bold.cyan('\n  Model Fallback Chain\n'));

      if (opts.set) {
        // Write fallback chain to config
        const { readFileSync, writeFileSync, mkdirSync } = await import('fs');
        const { resolve, dirname } = await import('path');
        const { homedir } = await import('os');
        const configDir = process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), '.tormentnexus');
        const configPath = resolve(configDir, 'config.jsonc');

        let config: Record<string, any> = {};
        try {
          const raw = readFileSync(configPath, 'utf8');
          config = JSON.parse(raw.replace(/\/\/.*$/gm, '').replace(/\/\*[\s\S]*?\*\//g, ''));
        } catch {}

        if (!config.fallback) config.fallback = {};
        config.fallback.chain = opts.set;
        config.fallback.strategy = opts.strategy || 'quota-aware';

        mkdirSync(dirname(configPath), { recursive: true });
        writeFileSync(configPath, JSON.stringify(config, null, 2) + '\n');
        console.log(chalk.green(`  ✓ Fallback chain set: ${opts.set.join(' → ')}`));
        console.log(chalk.dim(`  Strategy: ${config.fallback.strategy}`));
        console.log('');
        return;
      }

      // Show current chain
      const strategy = opts.strategy || 'quota-aware';
      console.log(chalk.dim('  Strategy: ') + strategy);

      // Try to read from config
      try {
        const { readFileSync } = await import('fs');
        const { resolve } = await import('path');
        const { homedir } = await import('os');
        const configPath = resolve(process.env.TORMENTNEXUS_CONFIG_DIR || resolve(homedir(), '.tormentnexus'), 'config.jsonc');
        const raw = readFileSync(configPath, 'utf8');
        const config = JSON.parse(raw.replace(/\/\/.*$/gm, '').replace(/\/\*[\s\S]*?\*\//g, ''));
        const chain = config.fallback?.chain ?? [];
        if (chain.length > 0) {
          console.log(chalk.dim('  Chain:    ') + chalk.white(chain.join(' → ')));
        } else {
          console.log(chalk.dim('  Chain:    ') + 'not configured (auto-detect from providers)');
          console.log(chalk.dim('  Default:  ') + 'claude-opus-4 → gpt-5 → gemini-2.5-pro → deepseek-v4');
        }
      } catch {
        console.log(chalk.dim('  Chain:    ') + 'not configured (auto-detect from providers)');
        console.log(chalk.dim('  Default:  ') + 'claude-opus-4 → gpt-5 → gemini-2.5-pro → deepseek-v4');
      }
      console.log('');
    });
}
