"""
Mnemosyne Sync Server
=====================
HTTP sync server using Python stdlib (no FastAPI, no external deps).

Provides endpoints for peer-to-peer memory synchronization:
- POST /sync/pull   — pull events from the server's event log
- POST /sync/push   — push events to the server
- GET  /sync/status — server sync statistics
"""

import base64
import hashlib
import hmac
import importlib
import json
import logging
import os
import sys
import threading
from datetime import datetime, timezone
from http.server import HTTPServer, BaseHTTPRequestHandler
from typing import Optional, Dict, Any, Callable

# Full-suite tests and embedded hosts can leave sys.modules["logging"] in a
# partially shadowed state. Recover the stdlib module instead of failing at
# import time; sync serving should not depend on global module-cache hygiene.
if not hasattr(logging, "getLogger"):
    sys.modules.pop("logging", None)
    logging = importlib.import_module("logging")

logger = logging.getLogger(__name__)


class SyncHTTPHandler(BaseHTTPRequestHandler):
    """HTTP handler for Mnemosyne sync endpoints.

    Relies on the class attribute *sync_engine* (set before the server
    starts, e.g. via run_sync_server()) to dispatch requests.
    """

    # Set externally by run_sync_server()
    sync_engine: Any = None
    api_key: Optional[str] = None
    jwt_secret: Optional[str] = None

    # Silence default HTTP server logs (we use our own logger)
    def log_message(self, fmt: str, *args: Any) -> None:
        logger.debug("HTTP: " + fmt, *args)

    def _send_json(
        self, status: int, data: dict, headers: Optional[dict] = None
    ) -> None:
        """Send a JSON response."""
        body = json.dumps(data, default=str).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Access-Control-Allow-Origin", "*")
        if headers:
            for k, v in headers.items():
                self.send_header(k, v)
        self.end_headers()
        self.wfile.write(body)

    def _send_error(self, status: int, message: str) -> None:
        self._send_json(status, {"error": message})

    def _read_body(self) -> Optional[dict]:
        """Read and parse the request body as JSON."""
        content_length = int(self.headers.get("Content-Length", 0))
        if content_length == 0:
            return None
        try:
            raw = self.rfile.read(content_length)
            return json.loads(raw.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError) as e:
            self._send_error(400, f"Invalid JSON body: {e}")
            return None
        except Exception as e:
            self._send_error(400, f"Failed to read body: {e}")
            return None

    @staticmethod
    def _decode_jwt_part(value: str) -> dict:
        """Decode one base64url-encoded JWT JSON segment."""
        padding = "=" * (-len(value) % 4)
        try:
            raw = base64.urlsafe_b64decode((value + padding).encode("ascii"))
            decoded = json.loads(raw.decode("utf-8"))
        except Exception as e:
            raise ValueError("invalid JWT encoding") from e
        if not isinstance(decoded, dict):
            raise ValueError("invalid JWT JSON")
        return decoded

    def _validate_jwt(self, token: str) -> dict:
        """Validate a compact HS256 JWT using the configured shared secret."""
        if not self.jwt_secret:
            raise ValueError("JWT secret is not configured")

        parts = token.split(".")
        if len(parts) != 3:
            raise ValueError("invalid JWT format")

        header = self._decode_jwt_part(parts[0])
        payload = self._decode_jwt_part(parts[1])
        algorithm = header.get("alg")
        if algorithm != "HS256":
            raise ValueError("unsupported JWT algorithm")

        signing_input = f"{parts[0]}.{parts[1]}".encode("ascii")
        expected = hmac.new(
            self.jwt_secret.encode("utf-8"), signing_input, hashlib.sha256
        ).digest()
        expected_sig = base64.urlsafe_b64encode(expected).decode("ascii").rstrip("=")
        try:
            valid_signature = hmac.compare_digest(parts[2], expected_sig)
        except TypeError as e:
            raise ValueError("invalid JWT signature") from e
        if not valid_signature:
            raise ValueError("invalid JWT signature")

        exp = payload.get("exp")
        if exp is not None:
            try:
                if float(exp) < datetime.now(timezone.utc).timestamp():
                    raise ValueError("token expired")
            except (TypeError, ValueError) as e:
                if isinstance(e, ValueError) and str(e) == "token expired":
                    raise
                raise ValueError("invalid JWT exp") from e
        return payload

    def _check_auth(self) -> bool:
        """Check API key / JWT authentication.

        Returns True if request is authorized, otherwise sends 401
        and returns False.
        """
        if self.api_key is None and self.jwt_secret is None:
            return True  # No auth configured

        if self.api_key:
            auth = self.headers.get("Authorization", "")
            if auth.startswith("Bearer "):
                token = auth[7:]
                if token == self.api_key:
                    return True
            self._send_error(401, "Invalid or missing API key")
            return False

        if self.jwt_secret:
            auth = self.headers.get("Authorization", "")
            if not auth.startswith("Bearer "):
                self._send_error(401, "Missing Bearer token")
                return False

            try:
                self._validate_jwt(auth[7:])
                return True
            except ValueError as e:
                self._send_error(401, f"JWT validation failed: {e}")
                return False

        return True  # No auth configured

    # --- CORS preflight ---
    def do_OPTIONS(self) -> None:
        self.send_response(204)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        self.send_header("Access-Control-Max-Age", "86400")
        self.end_headers()

    # --- POST /sync/pull ---
    def do_POST(self) -> None:
        parsed = self._parse_path(self.path)
        if parsed is None:
            self._send_error(404, f"Not found: {self.path}")
            return

        endpoint, _ = parsed

        if endpoint == "/sync/pull":
            self._handle_pull()
        elif endpoint == "/sync/push":
            self._handle_push()
        else:
            self._send_error(404, f"Not found: {self.path}")

    # --- GET /sync/status ---
    def do_GET(self) -> None:
        parsed = self._parse_path(self.path)
        if parsed is None:
            self._send_error(404, f"Not found: {self.path}")
            return

        endpoint, _ = parsed

        if endpoint == "/sync/status":
            self._handle_status()
        else:
            self._send_error(404, f"Not found: {self.path}")

    @staticmethod
    def _parse_path(path: str):
        """Parse request path, strip query string, return (path, query)."""
        if "?" in path:
            path, query = path.split("?", 1)
        else:
            query = ""
        # Ensure trailing slash doesn't break matching
        path = path.rstrip("/") or "/"
        return path, query

    def _handle_pull(self) -> None:
        """Handle POST /sync/pull — return events since cursor."""
        if not self._check_auth():
            return

        if self.sync_engine is None:
            self._send_error(500, "Sync engine not initialized")
            return

        body = self._read_body()
        if body is None:
            body = {}

        since = body.get("since")
        limit = body.get("limit", 1000)
        device_id = body.get("device_id")

        if limit is not None:
            limit = min(int(limit), 10000)

        try:
            result = self.sync_engine.pull_changes(
                since_cursor=since, limit=limit, device_id=device_id
            )
            self._send_json(200, result)
        except Exception as e:
            logger.exception("Error in pull_changes")
            self._send_error(500, f"Pull failed: {e}")

    def _handle_push(self) -> None:
        """Handle POST /sync/push — accept and apply events."""
        if not self._check_auth():
            return

        if self.sync_engine is None:
            self._send_error(500, "Sync engine not initialized")
            return

        body = self._read_body()
        if body is None:
            self._send_error(400, "Request body required")
            return

        events = body.get("events", [])
        if not isinstance(events, list):
            self._send_error(400, "'events' must be a list")
            return

        try:
            result = self.sync_engine.push_changes(events)
            result["next_cursor"] = datetime.now(timezone.utc).isoformat()
            self._send_json(200, result)
        except Exception as e:
            logger.exception("Error in push_changes")
            self._send_error(500, f"Push failed: {e}")

    def _handle_status(self) -> None:
        """Handle GET /sync/status — return server sync stats."""
        if not self._check_auth():
            return

        if self.sync_engine is None:
            self._send_error(500, "Sync engine not initialized")
            return

        try:
            status = self.sync_engine.get_status()
            self._send_json(200, status)
        except Exception as e:
            logger.exception("Error in get_status")
            self._send_error(500, f"Status failed: {e}")


def run_sync_server(
    host: str = "127.0.0.1",
    port: int = 8765,
    beam_instance=None,
    device_id: Optional[str] = None,
    api_key: Optional[str] = None,
    jwt_secret: Optional[str] = None,
    tls_cert: Optional[str] = None,
    tls_key: Optional[str] = None,
    daemon: bool = False,
) -> HTTPServer:
    """Start a Mnemosyne sync HTTP server.

    Uses only stdlib ``http.server`` — no FastAPI, no external deps.

    Parameters
    ----------
    host : str
        Bind address (default 127.0.0.1).
    port : int
        Bind port (default 8765).
    beam_instance :
        Mnemosyne or BeamMemory instance to sync against.
    device_id : str, optional
        Device ID for the server's sync engine.
    api_key : str, optional
        Shared API key for bearer-token auth.
    jwt_secret : str, optional
        JWT secret for token validation (minimal decode, no lib).
    tls_cert : str, optional
        Path to TLS certificate file (for HTTPS).
    tls_key : str, optional
        Path to TLS key file (for HTTPS).
    daemon : bool
        If True, start in a background thread. Default False (blocking).

    Returns
    -------
    HTTPServer
    """
    # Lazy import to avoid circular at import time
    from mnemosyne.core.sync import SyncEngine

    engine = SyncEngine(beam_instance, device_id=device_id)

    # Attach engine + auth to handler class
    SyncHTTPHandler.sync_engine = engine
    SyncHTTPHandler.api_key = api_key
    SyncHTTPHandler.jwt_secret = jwt_secret

    server = HTTPServer((host, port), SyncHTTPHandler)

    if tls_cert and tls_key:
        try:
            import ssl
            ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
            ctx.load_cert_chain(tls_cert, tls_key)
            server.socket = ctx.wrap_socket(server.socket, server_side=True)
        except Exception as e:
            logger.error("Failed to configure TLS: %s", e)
            raise

    if daemon:
        t = threading.Thread(
            target=server.serve_forever, daemon=True, name="sync-server"
        )
        t.start()
        logger.info(
            "Sync server started on %s:%s (background thread)", host, port
        )
    else:
        logger.info("Sync server listening on %s:%s", host, port)
        try:
            server.serve_forever()
        except KeyboardInterrupt:
            logger.info("Shutting down sync server")
            server.shutdown()

    return server


# ---------------------------------------------------------------------------
# CLI-friendly entry point
# ---------------------------------------------------------------------------

def main(args: Optional[list] = None) -> None:
    """Run the sync server from command-line args (matched to cli.py)."""
    if args is None:
        args = sys.argv[1:]

    import argparse
    parser = argparse.ArgumentParser(description="Mnemosyne Sync Server")
    parser.add_argument("--host", default="127.0.0.1", help="Bind address")
    parser.add_argument("--port", type=int, default=8765, help="Bind port")
    parser.add_argument("--db-path", help="Path to Mnemosyne SQLite database")
    parser.add_argument("--device-id", help="Device identifier for this server")
    parser.add_argument("--api-key", help="API key for bearer-token auth")
    parser.add_argument("--jwt-secret", help="JWT secret for token auth")
    parser.add_argument("--tls-cert", help="TLS certificate file path")
    parser.add_argument("--tls-key", help="TLS key file path")
    parser.add_argument("--debug", action="store_true", help="Enable debug logging")

    parsed = parser.parse_args(args)

    if parsed.debug:
        logging.basicConfig(level=logging.DEBUG)
    else:
        logging.basicConfig(level=logging.INFO)

    # Resolve beam instance
    if parsed.db_path:
        from mnemosyne.core.memory import Mnemosyne
        beam_instance = Mnemosyne(db_path=parsed.db_path)
    else:
        from mnemosyne.core.memory import Mnemosyne
        beam_instance = Mnemosyne()

    run_sync_server(
        host=parsed.host,
        port=parsed.port,
        beam_instance=beam_instance,
        device_id=parsed.device_id,
        api_key=parsed.api_key,
        jwt_secret=parsed.jwt_secret,
        tls_cert=parsed.tls_cert,
        tls_key=parsed.tls_key,
    )


if __name__ == "__main__":
    main()
