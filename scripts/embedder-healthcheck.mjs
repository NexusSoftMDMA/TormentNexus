import { pipeline, env } from "@xenova/transformers";

// Force offline mode to avoid hitting Hugging Face if the model is already cached
env.allowLocalModels = true;
env.useBrowserCache = false;

const MODEL_ID = "Xenova/all-MiniLM-L6-v2";
const PROBE_TEXT = "healthcheck ping from TormentNexus";
const TIMEOUT_MS = 30_000;

let embedder = null;
let loadMs = 0;
let error = null;

async function loadEmbedder() {
	const t0 = Date.now();
	embedder = await pipeline("feature-extraction", MODEL_ID, {
		progress_callback: (p) => {
			if (p.status === "initiate" || p.status === "download") {
				console.log(
					`  [embedder] ${p.file ?? p.name ?? p.status}: ${p.loaded ?? 0} / ${p.total ?? 0}`,
				);
			}
		},
	});
	loadMs = Date.now() - t0;
}

async function probe() {
	const t0 = Date.now();
	let dims = 0;
	if (typeof embedder === "function") {
		const out = await embedder(PROBE_TEXT, {
			pooling: "mean",
			normalize: true,
		});
		dims = Array.isArray(out?.data) ? out.data.length : 0;
	}
	return { dims, ms: Date.now() - t0 };
}

async function main() {
	console.log("=== Embedder Health Check ===");
	console.log("Model :", MODEL_ID);
	console.log("Cache :", process.env.XENOVA_CACHE_DIR ?? "(default)");

	try {
		await loadEmbedder();
		const { dims, ms } = await probe();
		console.log(`Loaded in ${loadMs}ms | probe ${ms}ms | dims ${dims}`);
		console.log("RESULT=OK");
	} catch (e) {
		error = e;
		console.log(`FAILED after ${loadMs}ms: ${e?.message ?? e}`);
		console.log("RESULT=FAIL");

		// Extra diagnostics
		try {
			const fs = await import("fs");
			const path = await import("path");
			const os = await import("os");

			const candidates = [
				path.join(
					os.homedir(),
					".cache",
					"huggingface",
					"transformers",
					"models",
					MODEL_ID,
				),
				path.join(os.homedir(), ".cache", "xenova", MODEL_ID),
				path.join(
					process.cwd(),
					"packages",
					"memory",
					"node_modules",
					"@xenova",
					"transformers",
					"models",
					MODEL_ID,
				),
				path.join(
					process.cwd(),
					"packages",
					"core",
					"node_modules",
					"@xenova",
					"transformers",
					"models",
					MODEL_ID,
				),
			];
			for (const c of candidates) {
				try {
					const entries = fs.readdirSync(c);
					console.log(`dir ${c} => ${entries.length} entries`);
				} catch {
					console.log(`dir ${c} => missing`);
				}
			}
		} catch {}
	}

	if (error) process.exit(1);
}

main();
