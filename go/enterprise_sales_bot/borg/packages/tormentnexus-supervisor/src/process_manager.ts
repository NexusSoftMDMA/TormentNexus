import psList from 'ps-list';
// import fkill from 'fkill'; // We might need to add fkill to dependencies if we want easy killing

export class ProcessManager {
    async listProcesses() {
        try {
            const processes = await psList();
            // Filter detailed info to save context? Or return all?
            // Returning top 50 by CPU/Memory might be better for an LLM
            return processes.slice(0, 100);
        } catch (error: any) {
            return `Error listing processes: ${error.message}`;
        }
    }

    async killProcess(pid: number) {
        try {
            process.kill(pid);
            return `Successfully sent signal to PID ${pid}`;
        } catch (error: any) {
            return `Error killing process ${pid}: ${error.message}`;
        }
    }
}
