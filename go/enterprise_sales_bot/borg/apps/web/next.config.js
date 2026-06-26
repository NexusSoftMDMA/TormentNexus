/** @type {import('next').NextConfig} */
const nextConfig = {
	// Skip static generation for pages that need runtime data
	output: "standalone",
	distDir: ".next-build",

	// Skip static page generation
	skipTrailingSlashRedirect: true,
	// Allow cross-origin dev requests from Go sidecar proxy
	allowedDevOrigins: ["127.0.0.1"],

	// Disable static optimization for all pages
	images: {
		unoptimized: true,
	},

	// Disable strict mode to avoid double renders
	reactStrictMode: false,

	// Transpile packages
	transpilePackages: ["@tormentnexus/ui"],

	// Turbopack config (Next.js 16 default)
	turbopack: {
		root: require("path").resolve(__dirname, "..", ".."),
	},

	// Silence type errors during build
	typescript: {
		ignoreBuildErrors: false,
	},
};

module.exports = nextConfig;
