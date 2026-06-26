import { type NextRequest, NextResponse } from "next/server";

const GO_SIDECAR_BASE =
	process.env.TORMENTNEXUS_GO_SIDECAR_URL || "http://127.0.0.1:7778";

/**
 * Go sidecar reverse proxy.
 *
 * Browser requests to /api/go/<path> are forwarded to the Go sidecar
 * at GO_SIDECAR_BASE/<path>.  The path is passed through verbatim so
 * that callers can target any sidecar endpoint — e.g. /api/go/health
 * → http://127.0.0.1:4300/health, or /api/go/api/mcp/status →
 * http://127.0.0.1:4300/api/mcp/status.
 */

export async function GET(
	request: NextRequest,
	{ params }: { params: Promise<{ path: string[] }> },
) {
	const resolvedParams = await params;
	const pathSegments = resolvedParams.path.join("/");
	const targetURL = `${GO_SIDECAR_BASE}/${pathSegments}`;

	try {
		const response = await fetch(targetURL, {
			headers: {
				accept: "application/json",
				...Object.fromEntries(
					Object.entries(request.headers).filter(([key]) =>
						["authorization", "cookie"].includes(key.toLowerCase()),
					),
				),
			},
			signal: AbortSignal.timeout(5000),
		});

		const contentType = response.headers.get("content-type") || "";
		if (contentType.includes("application/json")) {
			const data = await response.json();
			return NextResponse.json(data, { status: response.status });
		}
		// Non-JSON response — pass through as text
		const text = await response.text();
		return new NextResponse(text, {
			status: response.status,
			headers: { "content-type": contentType },
		});
	} catch (error) {
		const message =
			error instanceof Error ? error.message : "Go sidecar unreachable";
		return NextResponse.json(
			{ success: false, error: message, sidecarURL: targetURL },
			{ status: 502 },
		);
	}
}

export async function POST(
	request: NextRequest,
	{ params }: { params: Promise<{ path: string[] }> },
) {
	const resolvedParams = await params;
	const pathSegments = resolvedParams.path.join("/");
	const targetURL = `${GO_SIDECAR_BASE}/${pathSegments}`;

	try {
		let body: string | null = null;
		const contentType = request.headers.get("content-type") || "";
		if (contentType.includes("application/json")) {
			body = await request.text();
		}

		const response = await fetch(targetURL, {
			method: "POST",
			headers: {
				"content-type": contentType || "application/json",
			},
			body,
			signal: AbortSignal.timeout(10000),
		});

		const respContentType = response.headers.get("content-type") || "";
		if (respContentType.includes("application/json")) {
			const data = await response.json();
			return NextResponse.json(data, { status: response.status });
		}
		const text = await response.text();
		return new NextResponse(text, {
			status: response.status,
			headers: { "content-type": respContentType },
		});
	} catch (error) {
		const message =
			error instanceof Error ? error.message : "Go sidecar unreachable";
		return NextResponse.json(
			{ success: false, error: message, sidecarURL: targetURL },
			{ status: 502 },
		);
	}
}
