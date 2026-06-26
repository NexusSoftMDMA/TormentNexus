import Foundation

enum RemotePythonScript {
    static func wrap<Payload: Encodable>(_ payload: Payload, body: String) throws -> String {
        let encoder = JSONEncoder()
        let data = try encoder.encode(payload)
        let encodedPayload = data.base64EncodedString()

        return """
        import base64
        import json
        import pathlib
        import sys

        payload = json.loads(base64.b64decode("\(encodedPayload)").decode("utf-8"))

        \(sharedHelpers)

        \(body)
        """
    }

    private static let sharedHelpers = """
    import os
    import shutil
    import sqlite3

    def fail(message):
        print(json.dumps({
            "ok": False,
            "error": message,
        }, ensure_ascii=False))
        sys.exit(1)

    def stringify(value):
        if value is None:
            return None
        if isinstance(value, bytes):
            return value.decode("utf-8", errors="replace")
        return str(value)

    def normalize_text(value):
        text = stringify(value)
        if text is None:
            return None
        text = text.strip()
        return text or None

    def choose_table(tables, needle):
        lowered = needle.lower()
        for name in tables:
            if name.lower() == lowered:
                return name
        for name in tables:
            if lowered in name.lower():
                return name
        return None

    def choose_column(columns, choices):
        lowered = {column.lower(): column for column in columns}
        for choice in choices:
            if choice.lower() in lowered:
                return lowered[choice.lower()]
        for choice in choices:
            for column in columns:
                if choice.lower() in column.lower():
                    return column
        return None

    def quote_ident(value):
        return '"' + str(value).replace('"', '""') + '"'

    def quote_text(value):
        return "'" + str(value).replace("'", "''") + "'"

    def connect_sqlite_readonly(path):
        connection = None
        try:
            connection = sqlite3.connect(f"file:{path}?mode=ro", uri=True)
            connection.execute("PRAGMA schema_version").fetchone()
            return connection
        except sqlite3.OperationalError as exc:
            if connection is not None:
                try:
                    connection.close()
                except Exception:
                    pass
            message = str(exc).lower()
            if "unable to open database file" not in message and "readonly database" not in message:
                raise
            return sqlite3.connect(f"file:{path}?mode=ro&immutable=1", uri=True)

    def expand_remote_path(value, home=None, base_dir=None):
        if home is None:
            home = pathlib.Path.home()

        normalized = normalize_text(value)
        if normalized is None:
            return None
        expanded = os.path.expandvars(normalized)
        try:
            path = pathlib.Path(expanded).expanduser()
        except Exception:
            path = pathlib.Path(expanded)

        if not path.is_absolute():
            if expanded == "~":
                return home
            if expanded.startswith("~/"):
                return home / expanded[2:]
            if base_dir is not None:
                return base_dir / path
        return path

    def resolved_hermes_home(request=None):
        request_data = payload if request is None else request
        home = pathlib.Path.home()
        expanded = expand_remote_path(request_data.get("hermes_home"), home)
        if expanded is not None:
            return expanded
        env_home = expand_remote_path(os.environ.get("HERMES_HOME"), home)
        if env_home is not None:
            return env_home
        return home / ".hermes"

    def hermes_search_path(request=None):
        home = pathlib.Path.home()
        hermes_home = resolved_hermes_home(request)
        candidates = [
            hermes_home / "hermes-agent" / "venv" / "bin",
            home / ".local" / "bin",
            home / ".hermes" / "hermes-agent" / "venv" / "bin",
            home / ".cargo" / "bin",
            pathlib.Path("/opt/homebrew/bin"),
            pathlib.Path("/usr/local/bin"),
        ]

        entries = []
        seen = set()
        for candidate in candidates:
            try:
                entry = str(candidate)
            except Exception:
                continue
            if not entry or entry in seen:
                continue
            seen.add(entry)
            entries.append(entry)

        env_path = os.environ.get("PATH", "")
        if env_path:
            entries.append(env_path)
        return os.pathsep.join(entries)

    def find_hermes_binary(request=None):
        candidate = shutil.which("hermes", path=hermes_search_path(request))
        if candidate:
            return candidate
        return None

    def tilde(path, home=None):
        if home is None:
            home = pathlib.Path.home()
        try:
            relative = path.relative_to(home)
            return "~/" + relative.as_posix() if relative.as_posix() != "." else "~"
        except ValueError:
            return path.as_posix()

    def iter_session_store_candidates(hermes_home, home=None, hinted_path=None):
        if home is None:
            home = pathlib.Path.home()

        seen = set()

        def emit(candidate):
            if candidate is None:
                return None
            resolved = str(candidate)
            if resolved in seen or not candidate.is_file():
                return None
            seen.add(resolved)
            return candidate

        hinted_candidate = emit(expand_remote_path(hinted_path, home))
        if hinted_candidate is not None:
            yield hinted_candidate

        preferred = [
            hermes_home / "state.db",
            hermes_home / "state.sqlite",
            hermes_home / "state.sqlite3",
            hermes_home / "store.db",
            hermes_home / "store.sqlite",
            hermes_home / "store.sqlite3",
        ]

        for candidate in preferred:
            candidate = emit(candidate)
            if candidate is not None:
                yield candidate

        for candidate in sorted(
            [
                item
                for pattern in ("*.db", "*.sqlite", "*.sqlite3")
                for item in hermes_home.glob(pattern)
                if item.is_file()
            ],
            key=lambda item: item.stat().st_mtime,
            reverse=True,
        ):
            candidate = emit(candidate)
            if candidate is not None:
                yield candidate

        sessions_dir = hermes_home / "sessions"
        if sessions_dir.exists():
            for candidate in sorted(
                [
                    item
                    for pattern in ("*.db", "*.sqlite", "*.sqlite3")
                    for item in sessions_dir.rglob(pattern)
                    if item.is_file()
                ],
                key=lambda item: item.stat().st_mtime,
                reverse=True,
            ):
                candidate = emit(candidate)
                if candidate is not None:
                    yield candidate
    """
}
