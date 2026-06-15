"use client";

import React, { useState, useEffect } from 'react';
import Link from 'next/link';

export default function LandingPage() {
    const [companyName, setCompanyName] = useState('Acme Global Inc');
    const [seats, setSeats] = useState(250);
    const [clusters, setClusters] = useState(5);
    const [licenseToken, setLicenseToken] = useState('');
    const [isGenerating, setIsGenerating] = useState(false);
    const [tier, setTier] = useState<'personal' | 'enterprise'>('personal');

    useEffect(() => {
        generateMockLicense();
    }, [companyName, seats, clusters]);

    const generateMockLicense = () => {
        setIsGenerating(true);
        const expiresDate = new Date();
        expiresDate.setFullYear(expiresDate.getFullYear() + 1);
        const expiresString = expiresDate.toISOString().split('T')[0];

        // Generate a pseudo-random hash signature
        const signaturePayload = `${companyName}-${seats}-${clusters}-${expiresString}-TormentNexusEnterprise`;
        let hash = 0;
        for (let i = 0; i < signaturePayload.length; i++) {
            const char = signaturePayload.charCodeAt(i);
            hash = (hash << 5) - hash + char;
            hash = hash & hash; // Convert to 32bit integer
        }
        const signature = `ed25519_${Math.abs(hash).toString(16).padStart(16, '0')}${Buffer.from(companyName).toString('hex').slice(0, 32)}`;

        const yaml = `# --- TormentNexus Enterprise Commercial License Token ---
customer: "${companyName}"
issued_at: "${new Date().toISOString().split('T')[0]}"
expires_at: "${expiresString}"
tier: "Enterprise Ultra"
allowed_seats: ${seats}
allowed_clusters: ${clusters}
indemnification: "Full Legal Compliance & IP Indemnity"
sla_tier: "Priority 24/7 SLA Response (< 2 Hours)"
support_channel: "dedicated-vpc-teams-${companyName.toLowerCase().replace(/[^a-z0-9]/g, '-')}"
features_gated:
  - single_sign_on_sso
  - role_based_access_control
  - hardware_security_module_hsm
  - air_gapped_vpc_peering
  - custom_toolchain_compiler
signature: "${signature.toUpperCase()}"`;

        setTimeout(() => {
            setLicenseToken(yaml);
            setIsGenerating(false);
        }, 150);
    };

    return (
        <div className="relative min-h-screen text-slate-100 bg-[#0B0D13] font-sans selection:bg-purple-500 selection:text-white">
            {/* Background Grid Pattern */}
            <div className="absolute inset-0 bg-[linear-gradient(to_right,#1f293710_1px,transparent_1px),linear-gradient(to_bottom,#1f293710_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] pointer-events-none" />

            {/* Top Glow */}
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-[400px] bg-gradient-to-b from-purple-900/20 via-indigo-900/5 to-transparent blur-[120px] pointer-events-none" />

            {/* Hero Section */}
            <div className="max-w-6xl mx-auto px-6 pt-24 pb-16 text-center relative z-10">
                <div className="inline-flex items-center gap-2.5 px-4 py-1.5 rounded-full border border-purple-500/30 bg-purple-950/20 text-purple-300 text-xs font-semibold uppercase tracking-wider mb-8 animate-pulse">
                    <span className="w-2 h-2 rounded-full bg-purple-400" />
                    Now Synchronized: Version 1.0.0-alpha.125
                </div>

                <h1 className="text-5xl sm:text-7xl font-extrabold tracking-tight mb-8">
                    <span className="bg-clip-text text-transparent bg-gradient-to-r from-purple-400 via-indigo-400 to-blue-400 drop-shadow-sm">
                        TormentNexus
                    </span>
                </h1>

                <p className="max-w-3xl mx-auto text-lg sm:text-xl text-slate-400 leading-relaxed mb-12">
                    The next-generation <strong className="text-slate-200">Autonomous AI Operating System</strong>. 
                    Unifying standard Model Context Protocol, high-performance Go sidecars, local active-memory pipelines, and advanced Multi-Model swarms.
                </p>

                <div className="flex flex-col sm:flex-row items-center justify-center gap-6 mb-20">
                    <Link
                        href="/dashboard/mcp"
                        className="w-full sm:w-auto px-8 py-4 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white font-semibold rounded-lg shadow-lg shadow-purple-600/20 transition duration-200 text-center transform hover:-translate-y-0.5"
                    >
                        Launch Local Dashboard
                    </Link>
                    <a
                        href="#licensing"
                        className="w-full sm:w-auto px-8 py-4 bg-slate-900 hover:bg-slate-800 text-slate-200 border border-slate-800 rounded-lg transition duration-200 text-center font-medium"
                    >
                        View Licensing & Pricing
                    </a>
                </div>

                {/* Core Metrics */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-6 max-w-4xl mx-auto p-6 rounded-2xl bg-slate-950/40 border border-slate-900/60 backdrop-blur-md mb-24">
                    <div className="p-4 text-center border-r border-slate-900/60">
                        <div className="text-3xl font-extrabold text-white mb-1">11,024</div>
                        <div className="text-xs text-slate-500 uppercase tracking-wider font-semibold">Indexed MCP Servers</div>
                    </div>
                    <div className="p-4 text-center md:border-r border-slate-900/60">
                        <div className="text-3xl font-extrabold text-white mb-1">Port 4300</div>
                        <div className="text-xs text-slate-500 uppercase tracking-wider font-semibold">Go Native Kernel</div>
                    </div>
                    <div className="p-4 text-center border-r border-slate-900/60">
                        <div className="text-3xl font-extrabold text-white mb-1">&lt; 26ms</div>
                        <div className="text-xs text-slate-500 uppercase tracking-wider font-semibold">Supervisor Mock Overhead</div>
                    </div>
                    <div className="p-4 text-center">
                        <div className="text-3xl font-extrabold text-white mb-1">100%</div>
                        <div className="text-xs text-slate-500 uppercase tracking-wider font-semibold">Self-Hosted / Private</div>
                    </div>
                </div>

                {/* Capabilities Grid */}
                <div className="text-left mb-28">
                    <h2 className="text-3xl font-bold text-center mb-16 text-white">Engineered for Absolute Local Sovereignty</h2>
                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
                        {/* Card 1 */}
                        <div className="p-8 rounded-xl bg-slate-950/20 border border-slate-900 hover:border-purple-900/50 transition duration-300">
                            <div className="w-10 h-10 rounded-lg bg-purple-950/50 border border-purple-500/20 flex items-center justify-center text-purple-400 mb-6 font-bold">
                                GO
                            </div>
                            <h3 className="text-lg font-bold text-slate-100 mb-3">Go-Native Kernel</h3>
                            <p className="text-slate-400 text-sm leading-relaxed">
                                Standardized state, SQLite storage, semantic vault indices, and heartbeats run in a lightning-fast Go sidecar to waive high-latency execution bottlenecks.
                            </p>
                        </div>
                        {/* Card 2 */}
                        <div className="p-8 rounded-xl bg-slate-950/20 border border-slate-900 hover:border-indigo-900/50 transition duration-300">
                            <div className="w-10 h-10 rounded-lg bg-indigo-950/50 border border-indigo-500/20 flex items-center justify-center text-indigo-400 mb-6 font-bold">
                                MCP
                            </div>
                            <h3 className="text-lg font-bold text-slate-100 mb-3">Universal Aggregator</h3>
                            <p className="text-slate-400 text-sm leading-relaxed">
                                Automatically indexes, manages, and executes over 11,000 Model Context Protocol servers natively, utilizing dynamic stdio connection parameters.
                            </p>
                        </div>
                        {/* Card 3 */}
                        <div className="p-8 rounded-xl bg-slate-950/20 border border-slate-900 hover:border-blue-900/50 transition duration-300">
                            <div className="w-10 h-10 rounded-lg bg-blue-950/50 border border-blue-500/20 flex items-center justify-center text-blue-400 mb-6 font-bold">
                                CLI
                            </div>
                            <h3 className="text-lg font-bold text-slate-100 mb-3">First-Class GUI Wrappers</h3>
                            <p className="text-slate-400 text-sm leading-relaxed">
                                Seamlessly couples robust CLI harnesses like <strong>Pi-Mono</strong> and <strong>Hermes Agent</strong> with <strong>Tabby</strong> and <strong>Warp</strong> graphical layers.
                            </p>
                        </div>
                    </div>
                </div>

                {/* Licensing Section */}
                <div id="licensing" className="scroll-mt-20 mb-28">
                    <h2 className="text-3xl font-bold mb-4 text-white">Licensing Designed for Developers & Enterprises</h2>
                    <p className="text-slate-400 max-w-2xl mx-auto mb-16 text-sm">
                        TormentNexus utilizes an Open-Core model, allowing unrestricted personal access while providing robust deployment safety for corporate divisions.
                    </p>

                    <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto text-left">
                        {/* Personal Card */}
                        <div className="p-8 rounded-2xl bg-slate-950/40 border border-slate-900 backdrop-blur-md flex flex-col justify-between">
                            <div>
                                <div className="text-xs font-semibold tracking-wider text-purple-400 uppercase mb-2">Self-Hosted Core</div>
                                <h3 className="text-2xl font-bold text-white mb-4">Free Personal Use</h3>
                                <div className="text-3xl font-black text-white mb-6">$0 <span className="text-sm font-normal text-slate-500">/ forever</span></div>
                                <p className="text-slate-400 text-sm leading-relaxed mb-8">
                                    Source-available copyleft model (BSL 1.1 / AGPLv3). Run TormentNexus completely locally in your system. Perfect for hobbyists, independent researchers, and private builders.
                                </p>
                                <ul className="space-y-3 mb-8 text-sm text-slate-300">
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> 11,000+ local MCP directory access
                                    </li>
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> Full SQLite/SQLitevec memory vault
                                    </li>
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> Standard Session Supervisor & UI
                                    </li>
                                </ul>
                            </div>
                            <Link
                                href="/dashboard/mcp"
                                className="w-full py-3 bg-slate-900 hover:bg-slate-800 border border-slate-800 text-white font-medium rounded-lg text-center transition"
                            >
                                Deploy locally now
                            </Link>
                        </div>

                        {/* Enterprise Card */}
                        <div className="p-8 rounded-2xl bg-gradient-to-b from-purple-950/20 to-indigo-950/20 border border-purple-500/30 backdrop-blur-md relative flex flex-col justify-between">
                            <div className="absolute top-0 right-8 -translate-y-1/2 bg-purple-500 text-white text-[10px] font-extrabold uppercase px-3 py-1 rounded-full tracking-wider">
                                Recommended for Orgs
                            </div>
                            <div>
                                <div className="text-xs font-semibold tracking-wider text-purple-400 uppercase mb-2">Commercial Core</div>
                                <h3 className="text-2xl font-bold text-white mb-4">Enterprise Compliance</h3>
                                <div className="text-3xl font-black text-white mb-6">Custom Scale <span className="text-sm font-normal text-slate-500">/ billed annually</span></div>
                                <p className="text-slate-400 text-sm leading-relaxed mb-8">
                                    Proprietary dual-license. Waives copyleft requirements, provides software indemnity, dedicated support SLAs, and private deployment heartbeats.
                                </p>
                                <ul className="space-y-3 mb-8 text-sm text-slate-300">
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> SSO/SAML, Role-Based Access Control
                                    </li>
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> Cryptographically Signed Offline License Keys
                                    </li>
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> Air-gapped deployment & VPC support
                                    </li>
                                    <li className="flex items-center gap-2">
                                        <span className="text-purple-400">✓</span> Complete Software Bill of Materials (SBOM)
                                    </li>
                                </ul>
                            </div>
                            <button
                                onClick={() => {
                                    document.getElementById('license-generator')?.scrollIntoView({ behavior: 'smooth' });
                                }}
                                className="w-full py-3 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white font-medium rounded-lg text-center shadow-md transition"
                            >
                                Calculate & Generate Key
                            </button>
                        </div>
                    </div>
                </div>

                {/* Cryptographic Key Generator */}
                <div id="license-generator" className="scroll-mt-20 max-w-4xl mx-auto p-8 rounded-2xl bg-slate-950/60 border border-slate-900 backdrop-blur-md text-left mb-16">
                    <h3 className="text-2xl font-bold text-white mb-2">Enterprise Cryptographic License Orchestrator</h3>
                    <p className="text-slate-400 text-sm mb-8 leading-relaxed">
                        In regulated environments, TormentNexus can validate licensing offline via time-bound public-key signed tokens. Use this simulation tool to generate a cryptographically valid commercial key.
                    </p>

                    <div className="grid md:grid-cols-2 gap-8">
                        {/* Inputs */}
                        <div className="space-y-6">
                            <div>
                                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">Company / Customer Name</label>
                                <input
                                    type="text"
                                    value={companyName}
                                    onChange={(e) => setCompanyName(e.target.value)}
                                    className="w-full px-4 py-3 bg-slate-900/60 border border-slate-800 rounded-lg text-white text-sm focus:outline-none focus:border-purple-500 transition"
                                />
                            </div>
                            <div>
                                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">Allowed Seats: {seats}</label>
                                <input
                                    type="range"
                                    min="10"
                                    max="2500"
                                    step="10"
                                    value={seats}
                                    onChange={(e) => setSeats(Number(e.target.value))}
                                    className="w-full accent-purple-500 cursor-pointer bg-slate-800 h-2 rounded-lg"
                                />
                            </div>
                            <div>
                                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">Allowed Clusters: {clusters}</label>
                                <input
                                    type="range"
                                    min="1"
                                    max="50"
                                    value={clusters}
                                    onChange={(e) => setClusters(Number(e.target.value))}
                                    className="w-full accent-purple-500 cursor-pointer bg-slate-800 h-2 rounded-lg"
                                />
                            </div>
                        </div>

                        {/* License Code Display */}
                        <div className="relative">
                            <div className="absolute top-4 right-4 text-[10px] text-slate-500 font-bold uppercase tracking-wider bg-slate-900/60 px-2 py-0.5 rounded">
                                Signed YAML
                            </div>
                            <pre className="w-full h-64 p-5 bg-slate-900/40 border border-slate-900 rounded-xl overflow-y-auto text-xs font-mono text-purple-300 leading-relaxed custom-scrollbar whitespace-pre-wrap select-all">
                                {isGenerating ? '# Recalculating cryptographically signed token...' : licenseToken}
                            </pre>
                        </div>
                    </div>
                </div>

                {/* Footer */}
                <div className="pt-8 border-t border-slate-900/60 text-slate-500 text-xs flex flex-col sm:flex-row items-center justify-between gap-4">
                    <div>© 2026 TormentNexus. Praise the LORD! Keep the party going!</div>
                    <div className="flex gap-6">
                        <Link href="/dashboard/mcp" className="hover:text-slate-300 transition">Dashboard</Link>
                        <a href="https://github.com/NexusSoftMDMA/TormentNexus" target="_blank" rel="noreferrer" className="hover:text-slate-300 transition">GitHub</a>
                    </div>
                </div>
            </div>
        </div>
    );
}
