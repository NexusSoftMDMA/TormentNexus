import { NextResponse } from 'next/server';
import { markResetRequested } from '@/lib/authStore';

export async function POST(req: Request) {
    try {
        const body = await req.json();
        const email = String(body?.email ?? '').trim();

        if (!email || !email.includes('@')) {
            return NextResponse.json({ ok: false, error: 'Please provide a valid email address.' }, { status: 400 });
        }

        const result = await markResetRequested(email);

        const payload: {
            ok: true;
            message: string;
            resetUrl?: string;
        } = {
            ok: true,
            message: 'If this account exists, a reset link has been queued.',
        };

        if (result.resetToken) {
            payload.resetUrl = `/reset-password?token=${encodeURIComponent(result.resetToken)}`;
        }

        return NextResponse.json(payload);
    } catch {
        return NextResponse.json({ ok: false, error: 'Invalid request payload.' }, { status: 400 });
    }
}
