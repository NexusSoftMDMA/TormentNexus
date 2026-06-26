const MAX_TIMEOUT: number = 5000;

interface User {
    id: number;
    name: string;
}

class UserService {
    getUser(userId: number): User {
        return { id: userId, name: "" };
    }
}

function authenticate(token: string): boolean {
    return token.length > 0;
}

type UserID = number;

export const DEFAULT_CONFIG = Object.freeze({ timeout: 5000, retries: 3 });

const ALLOWED_ORIGINS = ["http://localhost:3000", "https://example.com"];

const formatName = (first: string, last: string): string => `${first} ${last}`;
