
import { GraphPanel } from "@/components/GraphPanel";

export default function CodeGraphPage() {
    return (
        <div className="h-screen w-full flex flex-col">
            <header className="h-14 border-b border-neutral-800 flex items-center px-6 bg-neutral-950">
                <h1 className="font-semibold text-lg text-neutral-200">Deep Code Intelligence</h1>
                <span className="ml-4 text-xs text-neutral-500 px-2 py-1 rounded bg-neutral-900 border border-neutral-800">Dependency Graph</span>
            </header>
            <main className="flex-1 relative">
                <GraphPanel />
            </main>
        </div>
    );
}
