import { createHash, randomBytes, randomUUID, scryptSync, timingSafeEqual } from 'crypto';
import { promises as fs } from 'fs';
import path from 'path';

type AuthUser = {
    id: string;
    name: string;
    email: string;
    passwordHash: string;
    createdAt: string;
    resetRequestedAt?: string;
    resetTokenHash?: string;
    resetTokenExpiresAt?: string;
};

type AuthSession = {
    id: string;
    userId: string;
    createdAt: string;
    expiresAt: string;
};

type AuthDb = {
    users: AuthUser[];
    sessions: AuthSession[];
};

const STORE_DIR = path.join(process.cwd(), '.tormentnexus-auth');
const STORE_FILE = path.join(STORE_DIR, 'users.json');
const DEFAULT_DB: AuthDb = { users: [], sessions: [] };

function normalizeEmail(email: string): string {
    return email.trim().toLowerCase();
}

function hashPassword(password: string): string {
    // Lightweight local hash for dev auth flow wiring.
    return scryptSync(password, 'tormentnexus-local-salt', 64).toString('hex');
}

function hashResetToken(token: string): string {
    return createHash('sha256').update(token).digest('hex');
}

function verifyPassword(password: string, passwordHash: string): boolean {
    const computed = Buffer.from(hashPassword(password), 'hex');
    const expected = Buffer.from(passwordHash, 'hex');
    if (computed.length !== expected.length) {
        return false;
    }
    return timingSafeEqual(computed, expected);
}

async function ensureStore(): Promise<void> {
    await fs.mkdir(STORE_DIR, { recursive: true });
    try {
        await fs.access(STORE_FILE);
    } catch {
        await fs.writeFile(STORE_FILE, JSON.stringify(DEFAULT_DB, null, 2), 'utf-8');
    }
}

async function readDb(): Promise<AuthDb> {
    await ensureStore();
    const raw = await fs.readFile(STORE_FILE, 'utf-8');
    try {
        const parsed = JSON.parse(raw) as AuthDb;
        return { users: parsed.users ?? [], sessions: parsed.sessions ?? [] };
    } catch {
        return DEFAULT_DB;
    }
}

async function writeDb(db: AuthDb): Promise<void> {
    await ensureStore();
    await fs.writeFile(STORE_FILE, JSON.stringify(db, null, 2), 'utf-8');
}

export async function createUser(input: { name: string; email: string; password: string }) {
    const db = await readDb();
    const normalized = normalizeEmail(input.email);
    const existing = db.users.find((u) => u.email === normalized);
    if (existing) {
        return { ok: false as const, reason: 'EXISTS' as const };
    }

    const user: AuthUser = {
        id: randomUUID(),
        name: input.name.trim(),
        email: normalized,
        passwordHash: hashPassword(input.password),
        createdAt: new Date().toISOString(),
    };

    db.users.push(user);
    await writeDb(db);

    return {
        ok: true as const,
        user: {
            id: user.id,
            name: user.name,
            email: user.email,
        },
    };
}

export async function authenticateUser(input: { email: string; password: string }) {
    const db = await readDb();
    const normalized = normalizeEmail(input.email);
    const user = db.users.find((u) => u.email === normalized);
    if (!user) {
        return { ok: false as const, reason: 'INVALID_CREDENTIALS' as const };
    }
    if (!verifyPassword(input.password, user.passwordHash)) {
        return { ok: false as const, reason: 'INVALID_CREDENTIALS' as const };
    }

    return {
        ok: true as const,
        user: {
            id: user.id,
            name: user.name,
            email: user.email,
        },
    };
}

export async function createSession(input: { userId: string; ttlSeconds?: number }) {
    const db = await readDb();
    const ttlSeconds = input.ttlSeconds ?? 60 * 60 * 24;
    const now = Date.now();

    db.sessions = db.sessions.filter((session) => {
        const expiry = Date.parse(session.expiresAt);
        return Number.isFinite(expiry) && expiry > now;
    });

    const session: AuthSession = {
        id: randomUUID(),
        userId: input.userId,
        createdAt: new Date(now).toISOString(),
        expiresAt: new Date(now + ttlSeconds * 1000).toISOString(),
    };

    db.sessions.push(session);
    await writeDb(db);
    return session;
}

export async function getUserBySession(sessionId: string) {
    if (!sessionId) {
        return null;
    }

    const db = await readDb();
    const now = Date.now();

    db.sessions = db.sessions.filter((session) => {
        const expiry = Date.parse(session.expiresAt);
        return Number.isFinite(expiry) && expiry > now;
    });

    const session = db.sessions.find((entry) => entry.id === sessionId);
    if (!session) {
        await writeDb(db);
        return null;
    }

    const user = db.users.find((entry) => entry.id === session.userId);
    if (!user) {
        db.sessions = db.sessions.filter((entry) => entry.id !== session.id);
        await writeDb(db);
        return null;
    }

    await writeDb(db);
    return {
        id: user.id,
        name: user.name,
        email: user.email,
    };
}

export async function revokeSession(sessionId: string) {
    if (!sessionId) {
        return { ok: true as const };
    }

    const db = await readDb();
    const nextSessions = db.sessions.filter((entry) => entry.id !== sessionId);
    if (nextSessions.length !== db.sessions.length) {
        db.sessions = nextSessions;
        await writeDb(db);
    }

    return { ok: true as const };
}

export async function markResetRequested(email: string) {
    const db = await readDb();
    const normalized = normalizeEmail(email);
    const user = db.users.find((u) => u.email === normalized);
    let resetToken: string | undefined;

    if (user) {
        const token = randomBytes(32).toString('hex');
        const expiresAt = new Date(Date.now() + 1000 * 60 * 30).toISOString(); // 30 minutes

        user.resetRequestedAt = new Date().toISOString();
        user.resetTokenHash = hashResetToken(token);
        user.resetTokenExpiresAt = expiresAt;
        await writeDb(db);
        resetToken = token;
    }

    // Always return success shape to avoid account enumeration.
    return { ok: true as const, resetToken };
}

export async function resetPasswordWithToken(input: { token: string; newPassword: string }) {
    const db = await readDb();
    const tokenHash = hashResetToken(input.token);
    const now = Date.now();

    const user = db.users.find((u) => {
        if (!u.resetTokenHash || !u.resetTokenExpiresAt) {
            return false;
        }
        const expiry = Date.parse(u.resetTokenExpiresAt);
        if (!Number.isFinite(expiry) || expiry < now) {
            return false;
        }
        const expected = Buffer.from(u.resetTokenHash, 'hex');
        const computed = Buffer.from(tokenHash, 'hex');
        if (expected.length !== computed.length) {
            return false;
        }
        return timingSafeEqual(expected, computed);
    });

    if (!user) {
        return { ok: false as const, reason: 'INVALID_OR_EXPIRED' as const };
    }

    user.passwordHash = hashPassword(input.newPassword);
    user.resetTokenHash = undefined;
    user.resetTokenExpiresAt = undefined;
    user.resetRequestedAt = new Date().toISOString();
    await writeDb(db);

    return {
        ok: true as const,
        user: {
            id: user.id,
            name: user.name,
            email: user.email,
        },
    };
}
