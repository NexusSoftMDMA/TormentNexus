import Foundation

final class RemoteHermesService: @unchecked Sendable {
    private let sshTransport: SSHTransport

    init(sshTransport: SSHTransport) {
        self.sshTransport = sshTransport
    }

    func discover(connection: ConnectionProfile) async throws -> RemoteDiscovery {
        let script = try RemotePythonScript.wrap(
            RemoteDiscoveryRequest(
                hermesHome: connection.remoteHermesHomePath,
                profileName: connection.cliHermesProfileName,
                customHomeMode: connection.usesCustomHermesHome
            ),
            body: discoveryScript
        )

        return try await sshTransport.executeJSON(
            on: connection,
            pythonScript: script,
            responseType: RemoteDiscovery.self
        )
    }

    func deleteHermesProfile(connection: ConnectionProfile, profileName: String) async throws -> RemoteProfileDeletionResult {
        let script = try RemotePythonScript.wrap(
            RemoteProfileDeletionRequest(profileName: profileName),
            body: profileDeletionScript
        )

        return try await sshTransport.executeJSON(
            on: connection,
            pythonScript: script,
            responseType: RemoteProfileDeletionResult.self
        )
    }

    private var discoveryScript: String {
        """
        import json
        import os
        import pathlib
        import shutil
        import sqlite3

        def discover_session_store(hermes_home: pathlib.Path):
            if not hermes_home.exists():
                return None

            for candidate in iter_session_store_candidates(hermes_home):
                try:
                    conn = connect_sqlite_readonly(candidate)
                    cursor = conn.execute(
                        "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
                    )
                    tables = [row[0] for row in cursor.fetchall()]
                    session_table = choose_table(tables, "sessions")
                    message_table = choose_table(tables, "messages")
                    if session_table and message_table:
                        conn.close()
                        return {
                            "kind": "sqlite",
                            "path": tilde(candidate, home),
                            "session_table": session_table,
                            "message_table": message_table,
                        }
                    conn.close()
                except Exception:
                    continue

            return None

        def discover_kanban(default_hermes_home: pathlib.Path, home: pathlib.Path):
            kanban_db = default_hermes_home / "kanban.db"
            previous_hermes_home = os.environ.get("HERMES_HOME")
            os.environ["HERMES_HOME"] = str(default_hermes_home)

            has_kanban_module = False
            dispatcher = None
            try:
                import hermes_cli.kanban_db  # noqa: F401
                has_kanban_module = True
            except Exception:
                has_kanban_module = False

            try:
                import hermes_cli.kanban as kanban_cli
                running, message = kanban_cli._check_dispatcher_presence()
                dispatcher = {
                    "running": bool(running),
                    "message": message or None,
                }
            except Exception:
                dispatcher = {
                    "running": None,
                    "message": None,
                }

            if previous_hermes_home is None:
                os.environ.pop("HERMES_HOME", None)
            else:
                os.environ["HERMES_HOME"] = previous_hermes_home

            return {
                "database_path": tilde(kanban_db, home),
                "exists": kanban_db.exists(),
                "host_wide": True,
                "has_hermes_cli": find_hermes_binary() is not None,
                "has_kanban_module": has_kanban_module,
                "dispatcher": dispatcher,
            }

        try:
            home = pathlib.Path.home()
            default_hermes_home = home / ".hermes"
            hermes_home = resolved_hermes_home()
            custom_home_mode = bool(payload.get("custom_home_mode"))
            user_path = hermes_home / "memories" / "USER.md"
            memory_path = hermes_home / "memories" / "MEMORY.md"
            soul_path = hermes_home / "SOUL.md"
            sessions_dir = hermes_home / "sessions"
            cron_jobs_path = hermes_home / "cron" / "jobs.json"
            kanban_database_path = default_hermes_home / "kanban.db"
            profiles_dir = default_hermes_home / "profiles"

            active_profile_name = payload.get("profile_name")
            if hermes_home == default_hermes_home:
                active_profile_name = "default"
            elif not active_profile_name:
                active_profile_name = hermes_home.name

            active_profile = {
                "name": active_profile_name,
                "path": tilde(hermes_home, home),
                "is_default": hermes_home == default_hermes_home,
                "exists": hermes_home.exists(),
            }

            if custom_home_mode:
                available_profiles = [active_profile]
            else:
                available_profiles = [{
                    "name": "default",
                    "path": tilde(default_hermes_home, home),
                    "is_default": True,
                    "exists": default_hermes_home.exists(),
                }]

                if profiles_dir.exists():
                    for item in sorted(
                        [entry for entry in profiles_dir.iterdir() if entry.is_dir()],
                        key=lambda entry: entry.name.lower(),
                    ):
                        available_profiles.append({
                            "name": item.name,
                            "path": tilde(item, home),
                            "is_default": False,
                            "exists": True,
                        })

            result = {
                "ok": True,
                "remote_home": tilde(home, home),
                "hermes_home": tilde(hermes_home, home),
                "active_profile": active_profile,
                "available_profiles": available_profiles,
                "paths": {
                    "user": tilde(user_path, home),
                    "memory": tilde(memory_path, home),
                    "soul": tilde(soul_path, home),
                    "sessions_dir": tilde(sessions_dir, home),
                    "cron_jobs": tilde(cron_jobs_path, home),
                    "kanban_database": tilde(kanban_database_path, home),
                },
                "exists": {
                    "user": user_path.exists(),
                    "memory": memory_path.exists(),
                    "soul": soul_path.exists(),
                    "sessions_dir": sessions_dir.exists(),
                    "cron_jobs": cron_jobs_path.exists(),
                    "kanban_database": kanban_database_path.exists(),
                },
                "session_store": discover_session_store(hermes_home),
                "kanban": discover_kanban(default_hermes_home, home),
            }

            print(json.dumps(result, ensure_ascii=False))
        except Exception as exc:
            fail(f"Unable to discover the remote Hermes workspace: {exc}")
        """
    }

    private var profileDeletionScript: String {
        """
        import json
        import pathlib
        import shutil

        try:
            profile_name = str(payload.get("profile_name") or "").strip()
            if not profile_name:
                fail("Profile name is required.")
            if profile_name == "default":
                fail("The default Hermes profile cannot be deleted from Hermes Desktop.")
            if "/" in profile_name or profile_name in {".", ".."}:
                fail("Profile name must be a profile name, not a path.")

            home = pathlib.Path.home()
            profile_path = home / ".hermes" / "profiles" / profile_name
            profiles_root = (home / ".hermes" / "profiles").resolve()
            resolved_profile_path = profile_path.resolve()
            if profiles_root not in resolved_profile_path.parents:
                fail("Resolved profile path is outside ~/.hermes/profiles.")
            if not profile_path.exists():
                fail(f"Profile '{profile_name}' does not exist on this host.")
            if not profile_path.is_dir():
                fail(f"Profile '{profile_name}' is not a directory.")

            shutil.rmtree(profile_path)
            print(json.dumps({
                "ok": True,
                "profile_name": profile_name,
                "deleted_path": tilde(profile_path, home),
            }, ensure_ascii=False))
        except Exception as exc:
            fail(f"Unable to delete Hermes profile: {exc}")
        """
    }
}

private struct RemoteDiscoveryRequest: Encodable {
    let hermesHome: String
    let profileName: String?
    let customHomeMode: Bool

    enum CodingKeys: String, CodingKey {
        case hermesHome = "hermes_home"
        case profileName = "profile_name"
        case customHomeMode = "custom_home_mode"
    }
}

private struct RemoteProfileDeletionRequest: Encodable {
    let profileName: String

    enum CodingKeys: String, CodingKey {
        case profileName = "profile_name"
    }
}

struct RemoteProfileDeletionResult: Decodable {
    let ok: Bool
    let profileName: String
    let deletedPath: String

    enum CodingKeys: String, CodingKey {
        case ok
        case profileName = "profile_name"
        case deletedPath = "deleted_path"
    }
}
