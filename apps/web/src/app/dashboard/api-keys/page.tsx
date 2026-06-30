"use client";

import { useEffect, useState } from "react";
import { Loader2, Key, Shield, Plus, Trash2 } from "lucide-react";

export default function ApiKeysPage() {
	const [data, setData] = useState<any>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState("");

	useEffect(() => {
		fetch("/api/go/api/api-keys")
			.then((r) => r.json().catch(() => null))
			.then((d) => {
				setData(d);
				setLoading(false);
			})
			.catch((e) => {
				setError(String(e));
				setLoading(false);
			});
	}, []);

	return (
		<div className="p-6 space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-bold flex items-center gap-2">
						<Key className="w-6 h-6" />
						API Keys & Authentication
					</h1>
					<p
						className="text-zinc-400 text-sm mt-1"
						title="Manage API keys for programmatic access to TormentNexus APIs, OAuth clients, and authentication providers"
					>
						Manage API keys, OAuth clients, and authentication providers used by
						tools and agents to access external services.
					</p>
				</div>
			</div>

			{loading && (
				<div className="flex items-center gap-2 text-zinc-500">
					<Loader2 className="w-4 h-4 animate-spin" /> Loading API keys...
				</div>
			)}
			{error && (
				<div className="text-red-400 bg-red-950/20 rounded-lg p-4 border border-red-900/50">
					{error}
				</div>
			)}
			{data && (
				<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
					{Array.isArray(data) ? (
						data.map((key: any, i: number) => (
							<div
								key={i}
								className="border border-zinc-800 rounded-lg p-4 bg-zinc-900/50"
							>
								<div className="flex items-center justify-between">
									<div className="font-medium text-sm">
										{key.name || key.id || `Key ${i + 1}`}
									</div>
									<span
										className={`text-xs px-2 py-0.5 rounded-full ${key.active !== false ? "bg-emerald-900/30 text-emerald-400" : "bg-zinc-800 text-zinc-500"}`}
									>
										{key.active !== false ? "Active" : "Inactive"}
									</span>
								</div>
								<div className="mt-2 text-xs text-zinc-500 font-mono">
									{key.key || key.token || `••••${(key.id || "").slice(-6)}`}
								</div>
								{key.scopes && (
									<div className="mt-2 text-xs text-zinc-600">
										Scopes:{" "}
										{Array.isArray(key.scopes)
											? key.scopes.join(", ")
											: key.scopes}
									</div>
								)}
							</div>
						))
					) : (
						<pre className="text-sm text-zinc-300 bg-zinc-900 rounded-lg p-4 overflow-auto max-h-[80vh] col-span-2">
							{JSON.stringify(data, null, 2)}
						</pre>
					)}
				</div>
			)}
		</div>
	);
}
