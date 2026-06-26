"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";

export default function MemoryRedirect() {
    const router = useRouter();

    useEffect(() => {
        router.push("/dashboard/brain");
    }, [router]);

    return (
        <div className="h-screen w-screen flex flex-col items-center justify-center bg-black text-zinc-500 gap-3">
            <Loader2 className="h-8 w-8 animate-spin text-pink-500" />
            <p className="text-sm font-mono uppercase tracking-widest">Redirecting to Brain &amp; Memory...</p>
        </div>
    );
}
