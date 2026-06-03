
// Marker file creation to prove code execution
import fs from 'fs';
import path from 'path';

try {
    fs.writeFileSync(path.join(process.cwd(), '.tormentnexus_startup_marker'), `Started at ${new Date().toISOString()}`);
} catch (e) { }
