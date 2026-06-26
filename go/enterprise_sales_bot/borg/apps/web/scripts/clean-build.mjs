import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const distDir = path.resolve(__dirname, "..", ".next-build");

if (fs.existsSync(distDir)) {
	const oldDir = `${distDir}-old-${Date.now()}`;
	try {
		fs.renameSync(distDir, oldDir);
		fs.rm(oldDir, { recursive: true, force: true }, (err) => {
			if (err) {
				console.warn(`[clean-build] Background cleanup warning: ${err.message}`);
			}
		});
	} catch (err) {
		try {
			fs.rmSync(distDir, { recursive: true, force: true });
		} catch (e) {
			console.warn(`[clean-build] Warning: Could not clean .next-build: ${e.message}`);
		}
	}
}
