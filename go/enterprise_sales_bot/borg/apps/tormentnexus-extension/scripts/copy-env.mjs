import { copyFileSync, existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const projectRoot = path.resolve(scriptDir, '..');
const envPath = path.join(projectRoot, '.env');
const exampleEnvPath = path.join(projectRoot, '.example.env');

if (!existsSync(envPath) && existsSync(exampleEnvPath)) {
    copyFileSync(exampleEnvPath, envPath);
    console.log('.example.env has been copied to .env');
}
