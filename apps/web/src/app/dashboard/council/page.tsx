"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function CouncilRedirect() {
	const router = useRouter();
	useEffect(() => {
		router.replace("/dashboard/swarm?tab=swarm");
	}, [router]);

	return (
		<div className="flex min-h-screen items-center justify-center bg-black text-zinc-400">
			<p>Redirecting to Swarm Control...</p>
		</div>
	);
}
