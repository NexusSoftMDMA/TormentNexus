#!/usr/bin/env node
import fs from 'node:fs';
import fsp from 'node:fs/promises';
import path from 'node:path';
import { performance } from 'node:perf_hooks';
import { parseConfig } from '../dist/src/config/loader.js';
import { TextContextIndex } from '../dist/src/context/text-index.js';

const DEFAULT_DIR = path.resolve('.mvmt-bench', 'text-index');
const DEFAULT_DOCS = 10_000;
const TOPICS = ['latency', 'oauth', 'connector', 'index', 'search', 'policy', 'mount', 'tunnel'];

const options = parseArgs(process.argv.slice(2));
const benchDir = path.resolve(options.dir ?? DEFAULT_DIR);
const docsDir = path.join(benchDir, 'docs');
const indexPath = path.join(benchDir, 'text-index.json');

main().catch((err) => {
  console.error(err instanceof Error ? err.message : String(err));
  process.exit(1);
});

async function main() {
  prepareBenchDir(benchDir, { force: options.force || !options.dir });

  const generateStart = performance.now();
  await generateCorpus(docsDir, options.docs);
  const generateMs = performance.now() - generateStart;

  const config = parseConfig({
    version: 1,
    mounts: [
      {
        name: 'bench',
        type: 'local_folder',
        path: '/bench',
        root: docsDir,
        writeAccess: false,
      },
    ],
  });
  const index = new TextContextIndex({
    mounts: config.mounts,
    indexPath,
    ...(options.maxFiles ? { maxIndexedFiles: options.maxFiles } : {}),
    ...(options.maxChunks ? { maxIndexedChunks: options.maxChunks } : {}),
    ...(options.maxChunksPerFile ? { maxChunksPerFile: options.maxChunksPerFile } : {}),
  });

  const rebuildStart = performance.now();
  const stats = await index.rebuild();
  const rebuildMs = performance.now() - rebuildStart;

  const searches = [];
  for (const query of ['needle-latency shard-7', 'needle-oauth', 'mounted-context benchmark']) {
    const start = performance.now();
    const results = await index.search(query, ['bench'], 10);
    searches.push({
      query,
      ms: Math.round(performance.now() - start),
      results: results.length,
      top: results[0]?.path,
    });
  }

  const output = {
    docs: options.docs,
    dir: benchDir,
    indexPath,
    indexMb: roundMb(fs.statSync(indexPath).size),
    generatedMs: Math.round(generateMs),
    rebuildMs: Math.round(rebuildMs),
    stats,
    searches,
    memoryMb: memoryMb(),
  };

  if (options.json) {
    console.log(JSON.stringify(output, null, 2));
    return;
  }

  console.log('mvmt text-index benchmark');
  console.log(`  corpus: ${output.docs} docs at ${output.dir}`);
  console.log(`  index:  ${output.indexMb} MB at ${output.indexPath}`);
  console.log(`  build:  generated ${output.generatedMs}ms, indexed ${output.rebuildMs}ms`);
  console.log(`  stats:  ${stats.files} files, ${stats.chunks} chunks${stats.truncated ? ' (truncated)' : ''}`);
  console.log(`  memory: rss ${output.memoryMb.rss}, heap ${output.memoryMb.heapUsed}`);
  for (const search of searches) {
    console.log(`  search: ${JSON.stringify(search.query)} ${search.ms}ms, results=${search.results}, top=${search.top ?? '(none)'}`);
  }
}

function parseArgs(args) {
  const parsed = {
    docs: DEFAULT_DOCS,
    dir: undefined,
    force: false,
    json: false,
    maxFiles: undefined,
    maxChunks: undefined,
    maxChunksPerFile: undefined,
  };

  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    switch (arg) {
      case '--docs':
        parsed.docs = parsePositiveInt(args[++index], '--docs');
        break;
      case '--dir':
        parsed.dir = requireValue(args[++index], '--dir');
        break;
      case '--force':
        parsed.force = true;
        break;
      case '--json':
        parsed.json = true;
        break;
      case '--max-files':
        parsed.maxFiles = parsePositiveInt(args[++index], '--max-files');
        break;
      case '--max-chunks':
        parsed.maxChunks = parsePositiveInt(args[++index], '--max-chunks');
        break;
      case '--max-chunks-per-file':
        parsed.maxChunksPerFile = parsePositiveInt(args[++index], '--max-chunks-per-file');
        break;
      case '--help':
      case '-h':
        printHelp();
        process.exit(0);
      default:
        throw new Error(`Unknown argument: ${arg}`);
    }
  }
  return parsed;
}

function prepareBenchDir(dir, options) {
  if (fs.existsSync(dir)) {
    if (!options.force) {
      throw new Error(`Benchmark dir already exists: ${dir}\nPass --force to replace it.`);
    }
    fs.rmSync(dir, { recursive: true, force: true });
  }
  fs.mkdirSync(dir, { recursive: true });
}

async function generateCorpus(root, count) {
  await fsp.mkdir(root, { recursive: true });
  const batchSize = 100;
  for (let offset = 0; offset < count; offset += batchSize) {
    const writes = [];
    for (let i = offset; i < Math.min(offset + batchSize, count); i += 1) {
      const topic = TOPICS[i % TOPICS.length];
      const shard = i % 29;
      const project = `project-${String(i % 100).padStart(3, '0')}`;
      const dir = path.join(root, project, `shard-${String(shard).padStart(2, '0')}`);
      writes.push(writeDoc(dir, i, topic, shard));
    }
    await Promise.all(writes);
  }
}

async function writeDoc(dir, index, topic, shard) {
  await fsp.mkdir(dir, { recursive: true });
  await fsp.writeFile(
    path.join(dir, `doc-${String(index).padStart(6, '0')}.md`),
    [
      `# Generated document ${index}`,
      '',
      `This fixture covers needle-${topic} shard-${shard}.`,
      `The deterministic body mentions mvmt search indexing benchmark corpus ${index}.`,
      `Related terms: local-files mounted-context agent-routing ${topic}.`,
      'The paragraph intentionally looks like a short note, not random bytes.',
      `Stable token sequence: alpha-${index % 13} beta-${index % 17} gamma-${index % 19}.`,
      '',
    ].join('\n'),
    'utf-8',
  );
}

function parsePositiveInt(value, flag) {
  const raw = requireValue(value, flag);
  const parsed = Number(raw);
  if (!Number.isInteger(parsed) || parsed < 1) {
    throw new Error(`${flag} must be a positive integer`);
  }
  return parsed;
}

function requireValue(value, flag) {
  if (!value || value.startsWith('--')) throw new Error(`${flag} requires a value`);
  return value;
}

function roundMb(bytes) {
  return Math.round((bytes / 1024 / 1024) * 10) / 10;
}

function memoryMb() {
  const usage = process.memoryUsage();
  return {
    rss: `${roundMb(usage.rss)} MB`,
    heapUsed: `${roundMb(usage.heapUsed)} MB`,
  };
}

function printHelp() {
  console.log(`Usage: npm run bench:text-index -- [options]

Options:
  --docs <n>                 Number of generated documents (default ${DEFAULT_DOCS})
  --dir <path>               Benchmark output directory (default ${DEFAULT_DIR})
  --force                    Replace an existing custom --dir
  --max-files <n>            Override TextContextIndex maxIndexedFiles
  --max-chunks <n>           Override TextContextIndex maxIndexedChunks
  --max-chunks-per-file <n>  Override TextContextIndex maxChunksPerFile
  --json                     Print JSON metrics
`);
}
