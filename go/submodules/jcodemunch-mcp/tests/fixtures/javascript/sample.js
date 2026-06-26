const MAX_TIMEOUT = 5000;

/**
 * User service class
 */
class UserService {
    getUser(userId) {
        return { id: userId };
    }
}

function authenticate(token) {
    return token.length > 0;
}

export const DEFAULT_HEADERS = {
    "Content-Type": "application/json",
    "Accept": "application/json",
};

const API_BASE_URL = "https://api.example.com";

const handleClick = () => console.log("clicked");
