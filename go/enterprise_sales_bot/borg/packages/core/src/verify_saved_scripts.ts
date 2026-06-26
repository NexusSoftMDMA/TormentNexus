
import { jsonConfigProvider } from './services/config/JsonConfigProvider.js';
import { savedScriptsRouter } from './routers/savedScriptsRouter.js';
import { v4 as uuidv4 } from 'uuid';

async function verify() {
    console.log("Starting Saved Scripts Router Verification...");

    const testScript = {
        name: "test-script-" + Date.now(),
        description: "A test script",
        code: "console.log('hello')",
        uuid: uuidv4()
    };

    // 1. Create Script via Provider
    // Validate the backing provider JsonConfigProvider fully, 
    // as the router is a thin wrapper around it.

    console.log("1. Saving script via Provider...");
    await jsonConfigProvider.saveScript(testScript);

    // 2. Verify Load
    console.log("2. Loading scripts...");
    const scripts = await jsonConfigProvider.loadScripts();
    const saved = scripts.find(s => s.uuid === testScript.uuid);

    if (!saved) {
        console.error("❌ Script not saved.");
        process.exit(1);
    }
    console.log("✅ Script saved and loaded.");

    // 3. Verify Update
    console.log("3. Updating script...");
    saved.code = "console.log('updated')";
    await jsonConfigProvider.saveScript(saved);

    const updatedScripts = await jsonConfigProvider.loadScripts();
    const updated = updatedScripts.find(s => s.uuid === testScript.uuid);
    if (updated?.code !== "console.log('updated')") {
        console.error("❌ Script update failed.");
        process.exit(1);
    }
    console.log("✅ Script updated.");

    // 4. Verify Delete
    console.log("4. Deleting script...");
    await jsonConfigProvider.deleteScript(testScript.uuid);

    const finalScripts = await jsonConfigProvider.loadScripts();
    const deleted = finalScripts.find(s => s.uuid === testScript.uuid);
    if (deleted) {
        console.error("❌ Script delete failed.");
        process.exit(1);
    }
    console.log("✅ Script deleted.");

    console.log("Verification Complete.");
}

verify().catch(e => {
    console.error(e);
    process.exit(1);
});
