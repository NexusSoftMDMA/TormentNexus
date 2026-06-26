import { NextResponse } from 'next/server';
import { resetPasswordWithToken } from '@/lib/authStore';

export async function POST(req: Request) {
    try {
        const body = await req.json();
        const token = String(body?.token ?? '').trim();
        const newPassword = String(body?.password ?? '');

        if (!token || !newPassword) {
            return NextResponse.json({ ok: false, error: 'Token and new password are required.' }, { status: 400 });
        }

        if (newPassword.length < 6) {
            return NextResponse.json({ ok: false, error: 'Password must be at least 6 characters.' }, { status: 400 });
        }

        const result = await resetPasswordWithToken({ token, newPassword });
        if (!result.ok) {
            return NextResponse.json({ ok: false, error: 'Reset token is invalid or expired.' }, { status: 400 });
        }

        return NextResponse.json({ ok: true, message: 'Password reset successful. You can now sign in.' });
    } catch {
        return NextResponse.json({ ok: false, error: 'Invalid request payload.' }, { status: 400 });
    }
}
