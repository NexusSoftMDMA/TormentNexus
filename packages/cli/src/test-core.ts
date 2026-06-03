console.log("DEBUG: Starting import test...");
const start = Date.now();
try {
    console.log("DEBUG: Importing @tormentnexus/core...");
    const core = await import('@tormentnexus/core');
    console.log("DEBUG: @tormentnexus/core loaded successfully in " + (Date.now() - start) + "ms");
    console.log("Exports:", Object.keys(core));
} catch (e) {
    console.error("DEBUG: Failed to import @tormentnexus/core", e);
}
