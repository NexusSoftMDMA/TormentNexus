import { randomUUID } from 'crypto';
import { NextResponse } from 'next/server';
import { authenticateUser } from '@/lib/authStore';

export async function POST(req: Request) {
    try {
        const body = await req.json();
        const email = String(body?.email ?? '').trim();
        const password = String(body?.password ?? '');

        if (!email || !password) {
            return NextResponse.json({ ok: false, error: 'Email and password are required.' }, { status: 400 });
        }

        const result = await authenticateUser({ email, password });
        if (!result.ok) {
            return NextResponse.json({ ok: false, error: 'Invalid email or password.' }, { status: 401 });
        }

        const response = NextResponse.json({ ok: true, user: result.user });
        response.cookies.set('tormentnexus_session', randomUUID(), {
            httpOnly: true,
            sameSite: 'lax',
            secure: false,
            path: '/',
            maxAge: 60 * 60 * 24,
        });
        return response;
    } catch {
        return NextResponse.json({ ok: false, error: 'Invalid request payload.' }, { status: 400 });
    }
}
