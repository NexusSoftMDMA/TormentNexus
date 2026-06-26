import Foundation
import Testing
@testable import HermesDesktop

struct RemotePythonScriptTests {
    @Test
    func sharedHelpersResolveHermesBinaryFromActiveWorkspaceFirst() throws {
        let sourceURL = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .appendingPathComponent("Sources/HermesDesktop/Utilities/RemotePythonScript.swift")
        let source = try String(contentsOf: sourceURL, encoding: .utf8)

        #expect(source.contains("def hermes_search_path(request=None):"))
        #expect(source.contains(#"hermes_home / "hermes-agent" / "venv" / "bin""#))
        #expect(source.contains("def find_hermes_binary(request=None):"))
    }

    @Test
    func kanbanHelperKeepsUpstreamBoardDatabasePaths() throws {
        let sourceURL = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .appendingPathComponent("Sources/HermesDesktop/Services/KanbanBrowserService.swift")
        let source = try String(contentsOf: sourceURL, encoding: .utf8)

        #expect(source.contains(#"return kanban_home_path() / "kanban.db""#))
        #expect(source.contains(#"return board_dir(normalized) / "kanban.db""#))
        #expect(!source.contains(#"return kanban_home_path() / "kanban" / "kanban.db""#))
        #expect(source.contains("def add_hermes_agent_import_paths():"))
        #expect(source.contains(#"home / ".hermes" / "hermes-agent""#))
        #expect(source.contains(#"resolved_hermes_home() / "hermes-agent""#))
        #expect(source.contains(#"venv_lib.glob("python*/site-packages")"#))
        #expect(source.contains("def load_hermes_env_file():"))
        #expect(source.contains(#"active_hermes_home = resolved_hermes_home()"#))
        #expect(source.contains(#"kanban_home_path() / ".env""#))
        #expect(source.contains("load_hermes_env_file()"))
        #expect(source.contains("tasks = direct_tasks(conn, include_archived)"))
        #expect(source.contains("run_columns = table_columns(conn, \"task_runs\")"))
    }

    @Test
    func sessionBrowserRefreshesModelFromMessageMetadata() throws {
        let sourceURL = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .appendingPathComponent("Sources/HermesDesktop/Services/SessionBrowserService.swift")
        let source = try String(contentsOf: sourceURL, encoding: .utf8)

        #expect(source.contains("def latest_model_for_session(context, session_id, fallback=None):"))
        #expect(source.contains("model = latest_model_for_session(context, session_id, model)"))
        #expect(source.contains("metadata.get(\"active_model\")"))
    }

    @Test
    func readonlySQLiteHelperFallsBackForWalDatabaseWithoutWritableSidecars() throws {
        let script = try RemotePythonScript.wrap([String: String](), body:
            """
            import shutil
            import tempfile

            root = pathlib.Path(tempfile.mkdtemp())
            db_path = root / "kanban.db"
            writer = sqlite3.connect(db_path)
            writer.execute("PRAGMA journal_mode=WAL")
            writer.execute("CREATE TABLE tasks(id TEXT PRIMARY KEY)")
            writer.execute("INSERT INTO tasks VALUES (?)", ("T1",))
            writer.commit()
            writer.close()

            os.chmod(root, 0o555)
            try:
                connection = connect_sqlite_readonly(db_path)
                count = connection.execute("SELECT COUNT(*) FROM tasks").fetchone()[0]
                connection.close()
            finally:
                os.chmod(root, 0o755)
                shutil.rmtree(root, ignore_errors=True)

            print(json.dumps({"ok": True, "count": count}, ensure_ascii=False))
            """
        )

        let temporaryDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString, isDirectory: true)
        try FileManager.default.createDirectory(
            at: temporaryDirectory,
            withIntermediateDirectories: true
        )
        defer {
            try? FileManager.default.removeItem(at: temporaryDirectory)
        }

        let scriptURL = temporaryDirectory.appendingPathComponent("readonly-sqlite-helper.py")
        try script.write(to: scriptURL, atomically: true, encoding: .utf8)

        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        process.arguments = ["python3", scriptURL.path]

        let outputPipe = Pipe()
        let errorPipe = Pipe()
        process.standardOutput = outputPipe
        process.standardError = errorPipe

        try process.run()
        process.waitUntilExit()

        let output = String(
            data: outputPipe.fileHandleForReading.readDataToEndOfFile(),
            encoding: .utf8
        ) ?? ""
        let error = String(
            data: errorPipe.fileHandleForReading.readDataToEndOfFile(),
            encoding: .utf8
        ) ?? ""

        #expect(process.terminationStatus == 0)
        #expect(error.isEmpty)
        #expect(output.contains("\"count\": 1"))
    }
}
