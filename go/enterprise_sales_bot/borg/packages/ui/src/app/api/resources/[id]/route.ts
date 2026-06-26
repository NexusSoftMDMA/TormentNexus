import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';

const CORE_API_URL = process.env.CORE_API_URL || 'http://localhost:3002';
const CORE_TOKEN = process.env.SUPER_AI_TOKEN || process.env.CORE_API_TOKEN || 'dev-token';

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;
    const res = await fetch(`${CORE_API_URL}/api/resources/${id}`, {
      headers: {
        Authorization: `Bearer ${CORE_TOKEN}`
      }
    });

    const data = await res.json().catch(() => ({}));
    return NextResponse.json(data, { status: res.status });
  } catch (error) {
    return NextResponse.json(
      { error: error instanceof Error ? error.message : 'Failed to fetch resource' },
      { status: 500 }
    );
  }
}
