import { NextResponse, NextRequest } from "next/server";

const JULES_API_BASE = "https://jules.googleapis.com/v1alpha";

async function proxyRequest(request: NextRequest, { params }: { params: Promise<{ path?: string[] }> }) {
    try {
        const apiKey = request.headers.get("x-jules-api-key");
        if (!apiKey) {
            return NextResponse.json({ error: "API key required" }, { status: 401 });
        }

        // Join path segments: ['sessions', '123'] -> '/sessions/123'
        const resolvedParams = await params;
        const pathSegments = resolvedParams.path || [];
        const pathStr = pathSegments.length > 0 ? `/${pathSegments.join("/")}` : "";

        // Also append query params from the original request!
        const searchParams = request.nextUrl.searchParams.toString();
        const queryString = searchParams ? `?${searchParams}` : "";

        const url = `${JULES_API_BASE}${pathStr}${queryString}`;

        const contentType = request.headers.get("Content-Type") || "application/json";
        const body = ['GET', 'HEAD'].includes(request.method) ? undefined : await request.text();

        const response = await fetch(url, {
            method: request.method,
            headers: {
                "Content-Type": contentType,
                "X-Goog-Api-Key": apiKey,
            },
            body: body,
        });

        const data = await response.json().catch(() => ({}));

        if (!response.ok) {
            console.error(`[Jules Proxy] Error ${response.status} for ${url}`, data);
        }

        return NextResponse.json(data, { status: response.status });
    } catch (error) {
        console.error(`[Jules API Proxy] Error processing ${request.method} request:`, error);
        return NextResponse.json(
            {
                error: "Proxy error",
                message: error instanceof Error ? error.message : "Unknown",
            },
            { status: 500 },
        );
    }
}

export { proxyRequest as GET, proxyRequest as POST, proxyRequest as DELETE, proxyRequest as PATCH, proxyRequest as PUT };
